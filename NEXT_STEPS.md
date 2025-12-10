# Next Steps: Fixing Browser Hang Issue

## Summary
The Google Maps scraper browser keeps hanging on Windows. We've tried setting environment variables, but scrapemate doesn't expose browser launch arguments through its API.

## Completed
✅ Enhanced debug logging
✅ Windows OS detection
✅ Environment variables set in `main.go` `init()` function
✅ Environment variables set in `filerunner.go` as backup

## Current Issue
The scraper is still hanging, which means environment variables aren't being respected by Playwright Go/scrapemate.

## Recommended Next Step: Fork Scrapemate

Since scrapemate doesn't support browser launch arguments, we need to:

### Option 1: Fork Scrapemate (Recommended)
1. Fork https://github.com/gosom/scrapemate
2. Add `WithBrowserArgs([]string)` option function
3. Pass args to Playwright Go's browser launch
4. Update `go.mod` with replace directive

### Option 2: Use Local Replace (For Testing)
1. Clone scrapemate locally
2. Make changes
3. Uncomment/modify replace in `go.mod`:
   ```go
   replace github.com/gosom/scrapemate v0.9.6 => ../scrapemate
   ```

### Option 3: Wait for scrapemate Update
Open an issue/PR on scrapemate to add browser args support.

## Immediate Action
Since the command is hanging, we should:
1. Stop the hanging process
2. Decide on approach (fork vs local modify)
3. Implement browser args support in scrapemate
4. Test with Windows-compatible args

## Browser Args Needed (Matching Working Scrapers)
```go
[]string{
    "--no-sandbox",
    "--disable-dev-shm-usage", 
    "--disable-blink-features=AutomationControlled",
}
```

