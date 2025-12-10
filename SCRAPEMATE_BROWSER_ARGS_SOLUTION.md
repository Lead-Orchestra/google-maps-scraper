# Scrapemate Browser Launch Arguments Solution

## Problem
Scrapemate doesn't expose browser launch arguments through its API. The documentation shows `scrapemateapp.WithJS()` accepts options like `Headfull()` and `DisableImages()`, but no way to pass custom browser launch args.

## Current Approach (May Not Work)
We've set environment variables in `main.go` `init()` function, but Playwright Go may not respect them.

## Recommended Solution: Fork Scrapemate

Since scrapemate is a Go module dependency, we should:

1. **Fork the scrapemate repository** (https://github.com/gosom/scrapemate)
2. **Add browser launch arguments support** to scrapemate's Playwright initialization
3. **Update go.mod** with a replace directive pointing to our fork

### Implementation Steps

1. Fork scrapemate on GitHub
2. Find where scrapemate initializes Playwright browser (likely in `scrapemateapp` package)
3. Add a `WithBrowserArgs([]string)` option function
4. Pass these args to Playwright Go's `Chromium.Launch()` call
5. Update `go.mod`:
   ```go
   replace github.com/gosom/scrapemate v0.9.6 => github.com/YOUR_USERNAME/scrapemate v0.9.6
   ```

### Alternative: Use Replace with Local Path

If we want to modify locally first:

1. Clone scrapemate locally
2. Add browser args support
3. Use local replace in `go.mod`:
   ```go
   replace github.com/gosom/scrapemate v0.9.6 => ../scrapemate
   ```

## Code Changes Needed in Scrapemate

In scrapemate's Playwright initialization code, we need to:

```go
// Add to scrapemateapp.Config
type Config struct {
    // ... existing fields
    BrowserArgs []string
}

// Add option function
func WithBrowserArgs(args []string) func(*Config) error {
    return func(cfg *Config) error {
        cfg.BrowserArgs = args
        return nil
    }
}

// In Playwright browser launch
browser, err := playwright.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
    Headless: playwright.Bool(headless),
    Args: cfg.BrowserArgs, // Pass through our args
})
```

## Usage in Google Maps Scraper

Once scrapemate supports browser args:

```go
opts = append(opts, scrapemateapp.WithJS(
    scrapemateapp.DisableImages(),
))
opts = append(opts, scrapemateapp.WithBrowserArgs([]string{
    "--no-sandbox",
    "--disable-dev-shm-usage",
    "--disable-blink-features=AutomationControlled",
}))
```

## References
- Scrapemate docs: https://gosom.github.io/scrapemate/#/
- Scrapemate repo: https://github.com/gosom/scrapemate
- Playwright Go: https://github.com/playwright-community/playwright-go

