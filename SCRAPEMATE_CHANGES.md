# Scrapemate Browser Args Support - Implementation Summary

## Changes Made to Scrapemate Fork

We've successfully added browser launch arguments support to the scrapemate fork, allowing us to customize browser launch options for Windows compatibility.

### 1. Added Browser Args to Config (`scrapemateapp/config.go`)

**Added to `jsOptions` struct:**
- `BrowserArgs []string` - Custom browser launch arguments
- `ReplaceDefaultArgs bool` - Flag to replace or append to defaults

**New Functions:**
- `WithBrowserArgs(args []string)` - Set custom browser args
- `ReplaceDefaultArgs()` - Replace all default args with custom ones

### 2. Updated JS Fetcher (`adapters/fetchers/jshttp/jshttp.go`)

**Added to `JSFetcherOptions`:**
- `BrowserArgs []string`
- `ReplaceDefaultArgs bool`

**Updated `newBrowser()` function:**
- Now accepts `customArgs []string` and `replaceDefaultArgs bool`
- Removed `--single-process` from defaults (causes Windows issues)
- Logic to either append or replace default args

**Key Fix:**
- **Removed `--single-process`** from default browser args - this was causing Windows stability issues

### 3. Updated Google Maps Scraper

**Modified `runner/filerunner/filerunner.go`:**
- Added Windows detection
- Ready to use `WithBrowserArgs()` if needed (defaults already have required args)

## Browser Args in Defaults

Scrapemate defaults already include Windows-compatible args:
- `--no-sandbox` ✅
- `--disable-dev-shm-usage` ✅  
- `--disable-blink-features=AutomationControlled` ✅

**Removed problematic arg:**
- `--single-process` ❌ (causes crashes on Windows)

## Usage

To use custom browser args in Google Maps scraper:

```go
opts = append(opts, scrapemateapp.WithJS(
    scrapemateapp.DisableImages(),
    scrapemateapp.WithBrowserArgs([]string{
        "--custom-arg",
        "--another-arg",
    }),
))
```

## Testing

1. Rebuild: `cd backend/submodules/google-maps-scraper && go build -o google-maps-scraper.exe`
2. Run with a simple query
3. Check if browser stays open without hanging

## Next Steps

The main fix is removing `--single-process` from defaults. Test to see if this resolves the hanging issue.

