# Windows Browser Fix - Launch Arguments

## Issue
The browser closes unexpectedly on Windows with error: `playwright: target closed: Target page, context or browser has been closed`

## Root Cause
Windows requires specific browser launch arguments for Playwright to work reliably. Our working TypeScript scrapers use:
- `--no-sandbox`
- `--disable-dev-shm-usage`
- `--disable-blink-features=AutomationControlled`

The Go-based Google Maps scraper using scrapemate doesn't currently set these.

## Solution
We need to ensure Playwright Go bindings use these Windows-compatible launch arguments. Since scrapemate doesn't directly expose browser launch args, we'll need to:

1. Set environment variables that Playwright Go respects
2. Or modify scrapemate configuration if possible
3. Or patch the browser launch in the Go code

## Comparison with Working Scrapers

### Facebook Scraper (TypeScript - WORKS)
```typescript
const launchArgs = [
    "--no-sandbox",
    "--disable-blink-features=AutomationControlled",
    "--disable-dev-shm-usage",
    ...fingerprintArgs,
    ...proxyArgs,
];

browser = await chromium.launch({
    headless,
    args: launchArgs,
});
```

### Google Maps Scraper (Go - CURRENT)
```go
opts = append(opts, scrapemateapp.WithJS(scrapemateapp.DisableImages()))
// No browser launch args specified!
```

## Implementation

### Changes Made to `runner/filerunner/filerunner.go`:
1. Added Windows OS detection
2. Set environment variables `PLAYWRIGHT_BROWSERS_ARGS` and `CHROMIUM_ARGS` with Windows-compatible flags
3. Added debug logging for Windows compatibility setup

### Environment Variables Set:
- `PLAYWRIGHT_BROWSERS_ARGS="--no-sandbox --disable-dev-shm-usage --disable-blink-features=AutomationControlled"`
- `CHROMIUM_ARGS="--no-sandbox --disable-dev-shm-usage --disable-blink-features=AutomationControlled"`

### Limitations:
**Note**: This approach relies on Playwright Go bindings respecting these environment variables. If scrapemate doesn't pass them through, this may not work. 

### Alternative Solutions (if environment variables don't work):
1. **Modify scrapemate library**: Add support for custom browser launch arguments
2. **Custom browser launcher**: Create a custom browser launcher that wraps Playwright with proper args
3. **Fork/modify scrapemate**: Add browser args support to scrapemate's configuration

### Testing:
Run the scraper and check debug logs for:
- `[DEBUG] Detected Windows OS - setting compatibility environment variables`
- Browser should stay open without "target closed" errors

If issues persist, we'll need to modify scrapemate or use a different approach.

