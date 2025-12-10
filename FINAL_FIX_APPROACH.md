# Final Fix Approach: Browser Launch Arguments

## Issue Summary
The Google Maps scraper browser keeps closing on Windows. Environment variables set in filerunner.go don't work because Playwright initializes before they're set.

## Solution Implemented

### 1. Set Environment Variables in `init()` Function
Added an `init()` function in `main.go` that runs BEFORE any package imports initialize Playwright. This ensures Windows-compatible browser args are set early enough.

**Changes:**
- Added `runtime` import
- Added `init()` function that sets:
  - `PLAYWRIGHT_BROWSERS_ARGS`
  - `CHROMIUM_ARGS`  
  - `PLAYWRIGHT_CHROMIUM_ARGS`
- All set to: `--no-sandbox --disable-dev-shm-usage --disable-blink-features=AutomationControlled`

### 2. Alternative: Fork scrapemate (If init() doesn't work)

If environment variables still don't work, we can:

1. **Fork scrapemate** repository
2. **Add browser launch args support** to scrapemate's Playwright initialization
3. **Use go.mod replace directive**:
   ```go
   replace github.com/gosom/scrapemate v0.9.6 => ../scrapemate-fork
   ```

## Testing
1. Rebuild: `go build -o google-maps-scraper.exe`
2. Run with a simple query
3. Check logs for `[INIT] Windows detected` message
4. Verify browser stays open without "target closed" errors

## If Still Not Working

The ultimate solution would be to fork scrapemate and add direct browser launch args support, similar to how TypeScript scrapers pass args to `chromium.launch({ args: [...] })`.

This would require modifying scrapemate's Playwright initialization code to accept and pass through browser launch arguments.

