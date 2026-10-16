# Retry and Backoff Configuration

This document explains how to configure retry and backoff settings for the Elasticsearch client to handle transient errors like connection resets, timeouts, and server errors.

## Configuration Options

The `Config` struct now includes the following retry-related fields:

- `RetryCount`: Maximum number of retry attempts (default: 0 = no retries for a
  single address; automatically set to the number of addresses when more than one
  address is configured)
- `RetryWaitTime`: Minimum wait time between retries (default: 100ms)
- `RetryMaxWaitTime`: Maximum wait time between retries (default: 2s)
- `RetryConditions`: Custom retry conditions (optional; when retries are enabled
  and this is empty, `DefaultRetryConditions()` is applied automatically)

## Default Retry Behavior

When `RetryCount > 0`, the client automatically retries on:

1. **Network errors** including:
   - "connection reset by peer" (your specific error case)
   - "connection refused"
   - "timeout" / "i/o timeout"
   - "network is unreachable"
   - "broken pipe"
   - "EOF"

2. **HTTP status codes**:
   - 429 Too Many Requests
   - 5xx server errors (except 501 Not Implemented)

When retries are enabled and no custom `RetryConditions` are supplied — including
when retries are auto-enabled because multiple addresses are configured — these
default conditions are installed automatically.

## Usage Examples

### Basic Retry Configuration

```go
client, err := elasticsearch.New(&elasticsearch.Config{
    Addresses: []string{"https://localhost:9200"},
    Username: "admin",
    Password: "admin",
    RetryCount: 3,
    RetryWaitTime: 500 * time.Millisecond,
    RetryMaxWaitTime: 5 * time.Second,
}, logger)
```

### Advanced Retry with Custom Conditions

```go
client, err := elasticsearch.New(&elasticsearch.Config{
    Addresses: []string{"https://localhost:9200"},
    Username: "admin",
    Password: "admin",
    RetryCount: 3,
    RetryWaitTime: 500 * time.Millisecond,
    RetryMaxWaitTime: 5 * time.Second,
    // Use built-in conditions optimized for Point-in-Time searches
    RetryConditions: elasticsearch.PITSearchRetryConditions(),
}, logger)
```

### Custom Retry Conditions

```go
customConditions := []resty.RetryConditionFunc{
    // Add your custom logic here
    func(res *resty.Response, err error) bool {
        if err != nil && strings.Contains(err.Error(), "specific error") {
            return true
        }
        return false
    },
}

client, err := elasticsearch.New(&elasticsearch.Config{
    Addresses:       []string{"https://localhost:9200"},
    RetryCount:       2,
    RetryConditions:  customConditions,
}, logger)
```

## Failover across multiple addresses

When `Config.Addresses` contains more than one node URL, requests are distributed
across the addresses in round-robin order and retry/failover is **enabled
automatically**:

- If `RetryCount` is left at its zero value (the default), `New` sets it to the
  number of addresses. Each retry attempt targets the next address in the
  round-robin cycle, so with N addresses the request makes up to N+1 attempts and
  every node is tried at least once before the request gives up.
- If `RetryConditions` is empty, `DefaultRetryConditions()` is installed
  automatically (network errors, 429, and 5xx except 501).

```go
// Failover is automatic: no RetryCount/RetryConditions required.
client, err := elasticsearch.New(&elasticsearch.Config{
    Addresses: []string{"https://node1:9200", "https://node2:9200"},
}, logger)
```

Explicit configuration always wins: an explicit `RetryCount > 0` is used as-is
(it is not raised to the number of addresses), and caller-provided
`RetryConditions` replace the defaults entirely.

With a **single** address, retries remain opt-in: `RetryCount == 0` (the default)
returns a failed request immediately. There is currently no opt-out field for
multi-address configurations — configure a single address if you need strict
no-retry semantics.

## Handling Your Specific Error

For the error you mentioned:
```
ERROR HTTP request failed error=read tcp 172.18.0.2:54492->10.221.84.160:443: read: connection reset by peer service=search
```

This will be automatically handled by the default retry conditions when you set `RetryCount > 0`, or automatically when multiple addresses are configured. The client will retry the request up to the specified number of times with exponential backoff.

## Built-in Helper Functions

- `elasticsearch.DefaultRetryConditions()`: Returns standard retry conditions
- `elasticsearch.PITSearchRetryConditions()`: Returns conditions optimized for long-running Point-in-Time searches, including Elasticsearch-specific transient errors like:
  - `search_phase_execution_exception`
  - `too_many_buckets_exception`
  - `circuit_breaking_exception`

## Backoff Strategy

The client uses capped exponential backoff with jitter (based on AWS recommendations):
- Wait time doubles with each retry attempt
- Random jitter is added to prevent thundering herd
- Wait time is capped at `RetryMaxWaitTime`