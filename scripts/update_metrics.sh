#!/bin/bash

# Simulate updating metrics.json with random values

METRICS_FILE="../config/metrics.json"

# Generate random metrics
RESPONSE_TIME_MS=$(shuf -i 100-500 -n 1)
ERROR_RATE=$(echo "scale=2; $(shuf -i 1-5 -n 1)/100" | bc)
VULNERABILITY_COUNT=$(shuf -i 0-15 -n 1)
INTRUSION_ATTEMPTS=$(shuf -i 0-10 -n 1)
CRITICAL_ASSETS=$(shuf -i 1-5 -n 1)
HIGH_VALUE_ASSETS=$(shuf -i 1-10 -n 1)

# Update metrics.json using jq
jq --arg rt "$RESPONSE_TIME_MS" \
   --arg er "$ERROR_RATE" \
   --arg vc "$VULNERABILITY_COUNT" \
   --arg ia "$INTRUSION_ATTEMPTS" \
   --arg ca "$CRITICAL_ASSETS" \
   --arg ha "$HIGH_VALUE_ASSETS" \
   '.quality_of_service.response_time_ms = ($rt | tonumber) |
    .quality_of_service.error_rate = ($er | tonumber) |
    .security_metrics.vulnerability_count = ($vc | tonumber) |
    .security_metrics.intrusion_attempts = ($ia | tonumber) |
    .asset_value.critical_assets = ($ca | tonumber) |
    .asset_value.high_value_assets = ($ha | tonumber)' \
    "$METRICS_FILE" > tmp.$$.json && mv tmp.$$.json "$METRICS_FILE"

echo "Metrics updated: ResponseTime=${RESPONSE_TIME_MS}ms, ErrorRate=${ERROR_RATE}, Vulnerabilities=${VULNERABILITY_COUNT}, Intrusions=${INTRUSION_ATTEMPTS}, CriticalAssets=${CRITICAL_ASSETS}, HighValueAssets=${HIGH_VALUE_ASSETS}"