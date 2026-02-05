# Prometheus/Mimir Metadata API Guide

This guide shows you how to query metadata from your Grafana Cloud Prometheus/Mimir instance using the HTTP API.

## Quick Start

Run the pre-built script to see all metadata:

```bash
./scripts/query-prometheus-metadata.sh
```

## API Endpoint

Your Grafana Cloud Prometheus endpoint:
```
https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom
```

**Note:** Grafana Cloud uses `/api/prom/api/v1/...` instead of the standard `/api/v1/...`

## Authentication

All requests require HTTP Basic Auth:
```bash
curl -u "USER_ID:API_TOKEN" "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/..."
```

## Metadata API Endpoints

### 1. Get All Metric Names

Get all metrics in your Prometheus instance:

```bash
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/label/__name__/values"
```

**Response:**
```json
{
  "status": "success",
  "data": [
    "farkle_active_games",
    "farkle_game_banks_total",
    "farkle_game_rolls_total",
    "k6_http_reqs_total",
    ...
  ]
}
```

**Filter to specific metrics:**
```bash
# Only Farkle metrics
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/label/__name__/values" \
  | jq -r '.data[] | select(startswith("farkle_"))'

# Only K6 metrics
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/label/__name__/values" \
  | jq -r '.data[] | select(startswith("k6_"))'
```

### 2. Get All Label Names

Get all label names across all metrics:

```bash
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/labels"
```

**Response:**
```json
{
  "status": "success",
  "data": [
    "__name__",
    "endpoint",
    "environment",
    "instance",
    "job",
    "method",
    "service",
    "status",
    ...
  ]
}
```

### 3. Get Label Values

Get all values for a specific label:

```bash
# Get all service values
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/label/service/values"

# Get all endpoint values
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/label/endpoint/values"

# Get all status code values
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/label/status/values"
```

**Response:**
```json
{
  "status": "success",
  "data": [
    "/api/roll",
    "/api/bank",
    "/api/reset",
    "/api/set-player-name",
    "/metrics",
    ...
  ]
}
```

**With time range:**
```bash
END=$(date +%s)
START=$((END - 3600))  # Last hour

curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/label/endpoint/values?start=${START}&end=${END}"
```

### 4. Get Series Metadata

Get all time series for a specific metric:

```bash
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/series?match[]=farkle_http_requests_total"
```

**Response:**
```json
{
  "status": "success",
  "data": [
    {
      "__name__": "farkle_http_requests_total",
      "endpoint": "/api/roll",
      "environment": "production",
      "instance": "localhost:8080",
      "job": "prometheus.scrape.farkle_metrics",
      "method": "POST",
      "service": "farkle-game",
      "status": "200"
    },
    {
      "__name__": "farkle_http_requests_total",
      "endpoint": "/api/bank",
      "environment": "production",
      "method": "GET",
      "service": "farkle-game",
      "status": "200"
    },
    ...
  ]
}
```

**Multiple matchers:**
```bash
# Get series with specific labels
curl -u "USER:PASS" \
  'https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/series?match[]={service="farkle-game",endpoint="/api/roll"}'

# Multiple metric patterns
curl -u "USER:PASS" \
  'https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/series?match[]=farkle_game_*&match[]=k6_http_*'
```

**With time range:**
```bash
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/series?match[]=farkle_http_requests_total&start=${START}&end=${END}"
```

### 5. Get Metric Metadata (Type, Help, Unit)

Get metadata about a specific metric:

```bash
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/metadata?metric=farkle_http_requests_total"
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "farkle_http_requests_total": [
      {
        "type": "counter",
        "help": "Total number of HTTP requests",
        "unit": ""
      }
    ]
  }
}
```

**Get metadata for all metrics:**
```bash
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/metadata"
```

## Querying Actual Metric Values

### Instant Query

Get current value of a metric:

```bash
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/query?query=farkle_game_rolls_total"
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "resultType": "vector",
    "result": [
      {
        "metric": {
          "__name__": "farkle_game_rolls_total",
          "environment": "production",
          "instance": "localhost:8080",
          "job": "prometheus.scrape.farkle_metrics",
          "service": "farkle-game"
        },
        "value": [1738770000, "12345"]
      }
    ]
  }
}
```

**With PromQL functions:**
```bash
# Rate of HTTP requests
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/query?query=rate(farkle_http_requests_total[5m])"

# Sum by endpoint
curl -u "USER:PASS" \
  'https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/query?query=sum%20by%20(endpoint)%20(farkle_http_requests_total)'

# P95 latency
curl -u "USER:PASS" \
  'https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/query?query=histogram_quantile(0.95,%20rate(farkle_http_request_duration_seconds_bucket[5m]))'
```

### Range Query

Get metric values over a time range:

```bash
END=$(date +%s)
START=$((END - 3600))  # Last hour

curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/query_range?query=farkle_game_rolls_total&start=${START}&end=${END}&step=60"
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "resultType": "matrix",
    "result": [
      {
        "metric": {
          "__name__": "farkle_game_rolls_total",
          "service": "farkle-game"
        },
        "values": [
          [1738766400, "1000"],
          [1738766460, "1050"],
          [1738766520, "1100"],
          ...
        ]
      }
    ]
  }
}
```

**Parameters:**
- `query` - PromQL query
- `start` - Start timestamp (Unix epoch)
- `end` - End timestamp (Unix epoch)
- `step` - Query resolution step width (in seconds)

## Example Queries for Farkle Metrics

### Game Activity

```bash
# Total rolls
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/query?query=farkle_game_rolls_total"

# Roll rate (rolls per second)
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/query?query=rate(farkle_game_rolls_total[1m])"

# Total farkles
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/query?query=farkle_game_farkles_total"

# Farkle rate (farkles per roll)
curl -u "USER:PASS" \
  'https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/query?query=farkle_game_farkles_total%20%2F%20farkle_game_rolls_total'
```

### HTTP Performance

```bash
# Request rate by endpoint
curl -u "USER:PASS" \
  'https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/query?query=sum%20by%20(endpoint)%20(rate(farkle_http_requests_total[5m]))'

# P95 response time
curl -u "USER:PASS" \
  'https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/query?query=histogram_quantile(0.95,%20rate(farkle_http_request_duration_seconds_bucket[5m]))'

# Error rate
curl -u "USER:PASS" \
  'https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/query?query=sum(rate(farkle_http_requests_total{status=~"5.."}[5m]))%20%2F%20sum(rate(farkle_http_requests_total[5m]))'
```

### K6 Load Testing

```bash
# Current VUs
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/query?query=k6_vus"

# Request rate
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/query?query=rate(k6_http_reqs_total[1m])"

# P99 duration
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/query?query=k6_http_req_duration_p99"

# Error rate
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/query?query=k6_http_req_failed_rate"
```

## Using curl with Grafana Cloud

### Environment Variables

Store credentials in environment variables:

```bash
export PROM_URL="https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom"
export PROM_USER="1890848"
export PROM_PASS="glc_..."

# Query
curl -u "${PROM_USER}:${PROM_PASS}" \
  "${PROM_URL}/api/v1/label/__name__/values"
```

### Using .netrc

Store credentials in `~/.netrc` for automatic auth:

```bash
cat >> ~/.netrc <<EOF
machine prometheus-prod-13-prod-us-east-0.grafana.net
login 1890848
password glc_...
EOF

chmod 600 ~/.netrc

# Query without -u flag
curl -n "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/labels"
```

## Advanced Queries

### URL Encoding

For complex PromQL queries, use URL encoding:

```bash
# Without encoding (may fail)
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/query?query=sum by (endpoint) (rate(farkle_http_requests_total[5m]))"

# With encoding (reliable)
QUERY=$(echo 'sum by (endpoint) (rate(farkle_http_requests_total[5m]))' | jq -sRr @uri)
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/query?query=${QUERY}"
```

### Pagination

Some endpoints support pagination:

```bash
# Get first 100 series
curl -u "USER:PASS" \
  "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/series?match[]=farkle_*&limit=100"
```

### Multiple Matchers

Combine multiple label matchers:

```bash
# AND conditions
curl -u "USER:PASS" \
  'https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/series?match[]={service="farkle-game",endpoint="/api/roll",status="200"}'

# OR conditions (multiple match[] parameters)
curl -u "USER:PASS" \
  'https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/series?match[]=farkle_game_rolls_total&match[]=farkle_game_banks_total'
```

## Programmatic Access

### Python Example

```python
import requests
from urllib.parse import urlencode

PROM_URL = "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom"
PROM_USER = "1890848"
PROM_PASS = "glc_..."

def query_prometheus(query):
    url = f"{PROM_URL}/api/v1/query"
    params = {"query": query}
    response = requests.get(
        url,
        params=params,
        auth=(PROM_USER, PROM_PASS)
    )
    return response.json()

# Example usage
result = query_prometheus("farkle_game_rolls_total")
print(result)
```

### Go Example

```go
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
)

const (
    promURL  = "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom"
    promUser = "1890848"
    promPass = "glc_..."
)

func queryPrometheus(query string) (map[string]interface{}, error) {
    u := fmt.Sprintf("%s/api/v1/query?query=%s", promURL, url.QueryEscape(query))

    req, err := http.NewRequest("GET", u, nil)
    if err != nil {
        return nil, err
    }

    req.SetBasicAuth(promUser, promPass)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    var result map[string]interface{}
    err = json.Unmarshal(body, &result)
    return result, err
}

func main() {
    result, err := queryPrometheus("farkle_game_rolls_total")
    if err != nil {
        panic(err)
    }

    fmt.Printf("%+v\n", result)
}
```

## Error Handling

### Common Error Codes

- **401 Unauthorized**: Invalid credentials
- **403 Forbidden**: Valid credentials but insufficient permissions
- **404 Not Found**: Endpoint doesn't exist (check path includes `/api/prom`)
- **422 Unprocessable Entity**: Invalid PromQL query
- **503 Service Unavailable**: Grafana Cloud is temporarily unavailable

### Error Response Format

```json
{
  "status": "error",
  "errorType": "bad_data",
  "error": "invalid parameter 'query': parse error at char 5: unexpected character: '!'"
}
```

### Debugging

```bash
# Verbose output
curl -v -u "USER:PASS" "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/labels"

# Show HTTP headers only
curl -I -u "USER:PASS" "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/api/v1/labels"

# Show timing information
curl -w "@curl-format.txt" -o /dev/null -s -u "USER:PASS" "..."
```

**curl-format.txt:**
```
    time_namelookup:  %{time_namelookup}\n
       time_connect:  %{time_connect}\n
    time_appconnect:  %{time_appconnect}\n
   time_pretransfer:  %{time_pretransfer}\n
      time_redirect:  %{time_redirect}\n
 time_starttransfer:  %{time_starttransfer}\n
                    ----------\n
         time_total:  %{time_total}\n
```

## Resources

- **Prometheus HTTP API**: https://prometheus.io/docs/prometheus/latest/querying/api/
- **PromQL Guide**: https://prometheus.io/docs/prometheus/latest/querying/basics/
- **Grafana Cloud Docs**: https://grafana.com/docs/grafana-cloud/
- **Mimir HTTP API**: https://grafana.com/docs/mimir/latest/references/http-api/

## Scripts

- **[`scripts/query-prometheus-metadata.sh`](../scripts/query-prometheus-metadata.sh)** - Run all metadata queries
- **[`scripts/test-prometheus-metadata.sh`](../scripts/test-prometheus-metadata.sh)** - Comprehensive metadata exploration

Run any script to see live data from your Grafana Cloud instance!
