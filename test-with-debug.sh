#!/bin/bash
# Enhanced test script with comprehensive debugging for Google Maps Scraper

set -e

echo "=== Google Maps Scraper - Enhanced Debug Test ==="
echo "Timestamp: $(date)"
echo ""

# Set environment variables for better debugging
export DISABLE_TELEMETRY=1
export DEBUG=pw:api
export PLAYWRIGHT_DEBUG=1

# Enable verbose logging
export SCRAPEMATE_LOG_LEVEL=debug

echo "Environment variables set:"
echo "  DISABLE_TELEMETRY=1"
echo "  DEBUG=pw:api"
echo "  PLAYWRIGHT_DEBUG=1"
echo "  SCRAPEMATE_LOG_LEVEL=debug"
echo ""

echo "Running scraper with enhanced logging..."
echo ""

# Run with timeout and capture all output
timeout 60 ./google-maps-scraper.exe \
  -input test-query.txt \
  -results test-results-debug.csv \
  -depth 1 \
  -c 1 \
  -exit-on-inactivity 2m \
  2>&1 | tee scraper-debug-output.log

echo ""
echo "=== Test Complete ==="
echo "Check scraper-debug-output.log for detailed logs"

