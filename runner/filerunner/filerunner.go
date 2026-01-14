package filerunner

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gosom/google-maps-scraper/deduper"
	"github.com/gosom/google-maps-scraper/exiter"
	"github.com/gosom/google-maps-scraper/runner"
	"github.com/gosom/google-maps-scraper/tlmt"
	"github.com/gosom/scrapemate"
	"github.com/gosom/scrapemate/adapters/writers/csvwriter"
	"github.com/gosom/scrapemate/adapters/writers/jsonwriter"
	"github.com/gosom/scrapemate/scrapemateapp"
)

type fileRunner struct {
	cfg     *runner.Config
	input   io.Reader
	writers []scrapemate.ResultWriter
	app     *scrapemateapp.ScrapemateApp
	outfile *os.File
}

func New(cfg *runner.Config) (runner.Runner, error) {
	if cfg.RunMode != runner.RunModeFile {
		return nil, fmt.Errorf("%w: %d", runner.ErrInvalidRunMode, cfg.RunMode)
	}

	log.Printf("[DEBUG] Initializing file runner with config: concurrency=%d, depth=%d, debug=%v, fastmode=%v",
		cfg.Concurrency, cfg.MaxDepth, cfg.Debug, cfg.FastMode)

	ans := &fileRunner{
		cfg: cfg,
	}

	log.Printf("[DEBUG] Setting up input from: %s", cfg.InputFile)
	if err := ans.setInput(); err != nil {
		log.Printf("[ERROR] Failed to set input: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] Setting up writers for: %s (JSON=%v)", cfg.ResultsFile, cfg.JSON)
	if err := ans.setWriters(); err != nil {
		log.Printf("[ERROR] Failed to set writers: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] Setting up scrapemate app")
	if err := ans.setApp(); err != nil {
		log.Printf("[ERROR] Failed to set app: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] File runner initialized successfully")
	return ans, nil
}

func (r *fileRunner) Run(ctx context.Context) (err error) {
	var seedJobs []scrapemate.IJob

	t0 := time.Now().UTC()

	log.Printf("[DEBUG] Starting file runner execution")

	defer func() {
		elapsed := time.Now().UTC().Sub(t0)
		log.Printf("[DEBUG] File runner execution completed. Duration: %v, Jobs: %d", elapsed, len(seedJobs))
		
		params := map[string]any{
			"job_count": len(seedJobs),
			"duration":  elapsed.String(),
		}

		if err != nil {
			log.Printf("[ERROR] File runner error: %v", err)
			params["error"] = err.Error()
		}

		evt := tlmt.NewEvent("file_runner", params)

		_ = runner.Telemetry().Send(ctx, evt)
	}()

	log.Printf("[DEBUG] Creating deduper and exit monitor")
	dedup := deduper.New()
	exitMonitor := exiter.New()

	log.Printf("[DEBUG] Creating seed jobs from input (FastMode=%v, Lang=%s, Depth=%d)",
		r.cfg.FastMode, r.cfg.LangCode, r.cfg.MaxDepth)
	seedJobs, err = runner.CreateSeedJobs(
		r.cfg.FastMode,
		r.cfg.LangCode,
		r.input,
		r.cfg.MaxDepth,
		r.cfg.Email,
		r.cfg.GeoCoordinates,
		r.cfg.Zoom,
		r.cfg.Radius,
		dedup,
		exitMonitor,
		r.cfg.ExtraReviews,
	)
	if err != nil {
		log.Printf("[ERROR] Failed to create seed jobs: %v", err)
		return err
	}

	log.Printf("[DEBUG] Created %d seed jobs", len(seedJobs))
	for i, job := range seedJobs {
		if i < 3 { // Log first 3 jobs as examples
			log.Printf("[DEBUG] Seed job %d: %s", i+1, job.GetURL())
		}
	}

	exitMonitor.SetSeedCount(len(seedJobs))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	exitMonitor.SetCancelFunc(cancel)

	log.Printf("[DEBUG] Starting exit monitor")
	go exitMonitor.Run(ctx)

	log.Printf("[DEBUG] Starting scrapemate app with %d seed jobs", len(seedJobs))
	err = r.app.Start(ctx, seedJobs...)
	
	if err != nil {
		log.Printf("[ERROR] Scrapemate app error: %v", err)
	} else {
		log.Printf("[DEBUG] Scrapemate app completed successfully")
	}

	return err
}

func (r *fileRunner) Close(context.Context) error {
	if r.app != nil {
		return r.app.Close()
	}

	if r.input != nil {
		if closer, ok := r.input.(io.Closer); ok {
			return closer.Close()
		}
	}

	if r.outfile != nil {
		return r.outfile.Close()
	}

	return nil
}

func (r *fileRunner) setInput() error {
	switch r.cfg.InputFile {
	case "stdin":
		r.input = os.Stdin
	default:
		f, err := os.Open(r.cfg.InputFile)
		if err != nil {
			return err
		}

		r.input = f
	}

	return nil
}

func (r *fileRunner) setWriters() error {
	if r.cfg.CustomWriter != "" {
		parts := strings.Split(r.cfg.CustomWriter, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid custom writer format: %s", r.cfg.CustomWriter)
		}

		dir, pluginName := parts[0], parts[1]

		customWriter, err := runner.LoadCustomWriter(dir, pluginName)
		if err != nil {
			return err
		}

		r.writers = append(r.writers, customWriter)
	} else {
		var resultsWriter io.Writer

		switch r.cfg.ResultsFile {
		case "stdout":
			resultsWriter = os.Stdout
		default:
			f, err := os.Create(r.cfg.ResultsFile)
			if err != nil {
				return err
			}

			r.outfile = f

			resultsWriter = r.outfile
		}

		csvWriter := csvwriter.NewCsvWriter(csv.NewWriter(resultsWriter))

		if r.cfg.JSON {
			r.writers = append(r.writers, jsonwriter.NewJSONWriter(resultsWriter))
		} else {
			r.writers = append(r.writers, csvWriter)
		}
	}

	return nil
}

func (r *fileRunner) setApp() error {
	// Set Windows-compatible browser launch arguments via environment variables
	// These are critical for Windows stability and match our working TypeScript scrapers
	if runtime.GOOS == "windows" {
		log.Printf("[DEBUG] Detected Windows OS - setting compatibility environment variables")
		// Set environment variables that Playwright Go might respect
		// Note: These may need to be set before Playwright initialization
		if os.Getenv("PLAYWRIGHT_BROWSERS_ARGS") == "" {
			// Format: space-separated browser args
			os.Setenv("PLAYWRIGHT_BROWSERS_ARGS", "--no-sandbox --disable-dev-shm-usage --disable-blink-features=AutomationControlled")
			log.Printf("[DEBUG] Set PLAYWRIGHT_BROWSERS_ARGS for Windows compatibility")
		}
		// Alternative: Try setting chromium-specific args if supported
		if os.Getenv("CHROMIUM_ARGS") == "" {
			os.Setenv("CHROMIUM_ARGS", "--no-sandbox --disable-dev-shm-usage --disable-blink-features=AutomationControlled")
			log.Printf("[DEBUG] Set CHROMIUM_ARGS for Windows compatibility")
		}
	}

	log.Printf("[DEBUG] Configuring scrapemate app options")
	opts := []func(*scrapemateapp.Config) error{
		// scrapemateapp.WithCache("leveldb", "cache"),
		scrapemateapp.WithConcurrency(r.cfg.Concurrency),
		scrapemateapp.WithExitOnInactivity(r.cfg.ExitOnInactivityDuration),
	}

	log.Printf("[DEBUG] App config: concurrency=%d, exitOnInactivity=%v", 
		r.cfg.Concurrency, r.cfg.ExitOnInactivityDuration)

	if len(r.cfg.Proxies) > 0 {
		log.Printf("[DEBUG] Configuring %d proxy(ies)", len(r.cfg.Proxies))
		opts = append(opts,
			scrapemateapp.WithProxies(r.cfg.Proxies),
		)
	}

	if !r.cfg.FastMode {
		if r.cfg.Debug {
			log.Printf("[DEBUG] Configuring headful browser mode (Debug enabled)")
			opts = append(opts, scrapemateapp.WithJS(
				scrapemateapp.Headfull(),
				scrapemateapp.DisableImages(),
			),
			)
		} else {
			log.Printf("[DEBUG] Configuring headless browser mode")
			// Note: Windows-compatible browser launch arguments are already included in scrapemate defaults
			// (--no-sandbox, --disable-dev-shm-usage, --disable-blink-features=AutomationControlled)
			opts = append(opts, scrapemateapp.WithJS(scrapemateapp.DisableImages()))
		}
	} else {
		log.Printf("[DEBUG] Configuring fast mode with stealth (firefox)")
		opts = append(opts, scrapemateapp.WithStealth("firefox"))
	}

	if !r.cfg.DisablePageReuse {
		log.Printf("[DEBUG] Enabling page reuse (limit=2, browser limit=200)")
		opts = append(opts,
			scrapemateapp.WithPageReuseLimit(2),
			scrapemateapp.WithPageReuseLimit(200),
		)
	}

	log.Printf("[DEBUG] Creating scrapemate config with %d writer(s)", len(r.writers))
	matecfg, err := scrapemateapp.NewConfig(
		r.writers,
		opts...,
	)
	if err != nil {
		log.Printf("[ERROR] Failed to create scrapemate config: %v", err)
		return err
	}

	log.Printf("[DEBUG] Initializing scrapemate app")
	r.app, err = scrapemateapp.NewScrapeMateApp(matecfg)
	if err != nil {
		log.Printf("[ERROR] Failed to create scrapemate app: %v", err)
		return err
	}

	log.Printf("[DEBUG] Scrapemate app initialized successfully")
	return nil
}
