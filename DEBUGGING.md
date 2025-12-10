# Google Maps Scraper - Debugging Guide

## Enhanced Logging Added

Comprehensive debug logging has been added to the `filerunner` package to help diagnose issues during scraping operations.

## Debug Logs

The scraper now outputs detailed `[DEBUG]` and `[ERROR]` log messages at key stages:

### Initialization Logs
- Configuration parameters (concurrency, depth, debug mode, fast mode)
- Input file setup
- Writer configuration (CSV/JSON output)
- Browser mode configuration (headless/headful)

### Execution Logs
- Seed job creation count and examples
- Scrapemate app initialization
- Job processing status
- Error details with context

### Configuration Logs
- Browser mode (headless vs headful)
- Proxy configuration
- Page reuse settings
- Concurrency settings

## Running with Enhanced Logging

The scraper now automatically includes debug logs. To see them:

```bash
./google-maps-scraper.exe -input test-query.txt -results results.csv -depth 1 -c 1 -exit-on-inactivity 2m
```

## Known Issues

### Browser Closing Unexpectedly

If you see errors like:
```
playwright: target closed: Target page, context or browser has been closed
```

This can be caused by:
1. **Windows Security/Antivirus** - May block browser processes
2. **Resource Constraints** - Insufficient memory/CPU
3. **Playwright Compatibility** - Browser driver issues on Windows

### Troubleshooting Steps

1. **Check debug logs** - Look for `[DEBUG]` and `[ERROR]` messages to see where it fails
2. **Try headless mode** - Ensure `-debug` flag is NOT set (default is headless)
3. **Reduce concurrency** - Use `-c 1` for single-threaded operation
4. **Check browser installation** - Ensure Playwright browsers are downloaded
5. **Run with elevated permissions** - May be needed on Windows

## Environment Variables

You can also set these environment variables for additional debugging:

- `DISABLE_TELEMETRY=1` - Disable telemetry collection
- `PLAYWRIGHT_DEBUG=1` - Enable Playwright debug output
- `DEBUG=pw:api` - Enable Playwright API debugging

Example:
```bash
DISABLE_TELEMETRY=1 ./google-maps-scraper.exe -input queries.txt -results output.csv
```

## Log Output Format

Logs follow this format:
- `[DEBUG]` - Informational debug messages
- `[ERROR]` - Error messages with context
- Standard scrapemate JSON logs for job processing

All logs are output to stdout/stderr and can be redirected to a file:
```bash
./google-maps-scraper.exe -input queries.txt -results output.csv 2>&1 | tee scraper.log
```

