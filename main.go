package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gosom/google-maps-scraper/runner"
	"github.com/gosom/google-maps-scraper/runner/databaserunner"
	"github.com/gosom/google-maps-scraper/runner/filerunner"
	"github.com/gosom/google-maps-scraper/runner/installplaywright"
	"github.com/gosom/google-maps-scraper/runner/lambdaaws"
	"github.com/gosom/google-maps-scraper/runner/webrunner"
)

func init() {
	// CRITICAL: Set Windows-compatible browser launch arguments BEFORE any Playwright initialization
	// These must be set at package init time to ensure they're available when scrapemate/Playwright loads
	if runtime.GOOS == "windows" {
		// Set environment variables that Playwright Go might respect
		// These match the working TypeScript scrapers (Zillow, Facebook)
		if os.Getenv("PLAYWRIGHT_BROWSERS_ARGS") == "" {
			os.Setenv("PLAYWRIGHT_BROWSERS_ARGS", "--no-sandbox --disable-dev-shm-usage --disable-blink-features=AutomationControlled")
		}
		if os.Getenv("CHROMIUM_ARGS") == "" {
			os.Setenv("CHROMIUM_ARGS", "--no-sandbox --disable-dev-shm-usage --disable-blink-features=AutomationControlled")
		}
		// Try additional environment variables that Playwright Go might check
		if os.Getenv("PLAYWRIGHT_CHROMIUM_ARGS") == "" {
			os.Setenv("PLAYWRIGHT_CHROMIUM_ARGS", "--no-sandbox --disable-dev-shm-usage --disable-blink-features=AutomationControlled")
		}
		log.Println("[INIT] Windows detected - set browser compatibility environment variables")
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	runner.Banner()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan

		log.Println("Received signal, shutting down...")

		cancel()
	}()

	cfg := runner.ParseConfig()

	runnerInstance, err := runnerFactory(cfg)
	if err != nil {
		cancel()
		os.Stderr.WriteString(err.Error() + "\n")

		runner.Telemetry().Close()

		os.Exit(1)
	}

	if err := runnerInstance.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		os.Stderr.WriteString(err.Error() + "\n")

		_ = runnerInstance.Close(ctx)
		runner.Telemetry().Close()

		cancel()

		os.Exit(1)
	}

	_ = runnerInstance.Close(ctx)
	runner.Telemetry().Close()

	cancel()

	os.Exit(0)
}

func runnerFactory(cfg *runner.Config) (runner.Runner, error) {
	switch cfg.RunMode {
	case runner.RunModeFile:
		return filerunner.New(cfg)
	case runner.RunModeDatabase, runner.RunModeDatabaseProduce:
		return databaserunner.New(cfg)
	case runner.RunModeInstallPlaywright:
		return installplaywright.New(cfg)
	case runner.RunModeWeb:
		return webrunner.New(cfg)
	case runner.RunModeAwsLambda:
		return lambdaaws.New(cfg)
	case runner.RunModeAwsLambdaInvoker:
		return lambdaaws.NewInvoker(cfg)
	default:
		return nil, fmt.Errorf("%w: %d", runner.ErrInvalidRunMode, cfg.RunMode)
	}
}
