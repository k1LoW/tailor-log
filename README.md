# tailor-log

`tailor-log` is a tool for handling logs from a workspace on the Tailor Platform.

Key features of `tailor-log` are:

- **Ingest** - Fetch function and pipeline resolver logs from Tailor Platform and send them to external services. Currently supports Datadog with custom tags and service names. Includes position management to avoid duplicate processing.
- **Stream** - Monitor logs from Tailor Platform in real-time with configurable fetch intervals.

## As a GitHub Action

### Usage

```yaml
name: Ingest logs from Tailor Platform
on:
  schedule:
    - cron: '*/5 * * * *' # Run every 5 minutes
  workflow_dispatch:

jobs:
  ingest:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - uses: k1LoW/tailor-log@v0.5.0
        with:
          tailor-token: ${{ secrets.TAILOR_TOKEN }}
          github-token: ${{ secrets.GITHUB_TOKEN }}
          datadog-api-key: ${{ secrets.DD_API_KEY }}
          datadog-site: ${{ secrets.DD_SITE }}
          workspace-id: 'your-workspace-id'
          datadog-service: 'your-service-name'
          datadog-tags: 'env:production,team:backend'
          inputs: '*' # Fetch all logs (default)
```

## As a Standalone CLI

### Usage

Ingest logs from Tailor Platform and send them to Datadog:

```console
$ tailor-log ingest --workspace-id your-workspace-id \
  --datadog-service your-service-name \
  --datadog-tag env:production \
  --datadog-tag team:backend \
  --input '*' \
  --pos file
```

Stream logs from Tailor Platform in real-time:

```console
$ tailor-log stream --workspace-id your-workspace-id \
  --input '*' \
  --fetch-interval 5sec
```
### Input Filtering

The `--input` option allows you to filter which logs to fetch using wildcard patterns. Each resource type has a specific key format:

#### Function Logs

Function logs use the format: `function:{scriptName}`

Examples:
```console
# Fetch all function logs
$ tailor-log ingest --input 'function:*'

# Fetch logs from a specific function
$ tailor-log ingest --input 'function:myFunction'

# Fetch logs from functions matching a pattern
$ tailor-log ingest --input 'function:user-*'
```

#### Pipeline Resolver Logs

Pipeline resolver logs use the format: `pipeline:{namespaceName}:resolver:{resolverName}`

Examples:
```console
# Fetch all pipeline resolver logs
$ tailor-log ingest --input 'pipeline:*'

# Fetch logs from a specific namespace
$ tailor-log ingest --input 'pipeline:myNamespace:*'

# Fetch logs from a specific resolver
$ tailor-log ingest --input 'pipeline:myNamespace:resolver:myResolver'

# Fetch logs from resolvers matching a pattern
$ tailor-log ingest --input 'pipeline:*:resolver:user-*'
```

#### Multiple Patterns

You can specify multiple `--input` options to match different patterns:

```console
$ tailor-log ingest \
  --input 'function:user-*' \
  --input 'pipeline:production:resolver:*' \
  --input 'function:admin-service'
```

If no `--input` option is specified, the default value `'*'` is used, which matches all logs.


Environment variables:

- `TAILOR_TOKEN` - Tailor Platform token (required)
- `DD_API_KEY` - Datadog API key (required for ingest command)
- `DD_SITE` - Datadog site (required for ingest command)

### Install

**Using Homebrew (macOS):**

```console
$ brew install k1LoW/tap/tailor-log
```

**Using `go install`:**

```console
$ go install github.com/k1LoW/tailor-log@latest
```

**Using `gh-setup` (GitHub Actions):**

```yaml
- uses: k1LoW/gh-setup@v1
  with:
    repo: k1LoW/tailor-log
```

**Download from GitHub Releases:**

Download the binary from [releases page](https://github.com/k1LoW/tailor-log/releases) and place it in your `$PATH`.

**Using `deb`, `rpm`, or `apk`:**

Download the package from [releases page](https://github.com/k1LoW/tailor-log/releases) and install it.
