# TailScale Device Observability

A Prometheus exporter written in Go that polls the [Tailscale API](https://tailscale.com/api) and exposes per-device health metrics for your tailnet. Plug it into any standard Prometheus + Grafana stack to get visibility into device connectivity, key expiry, and client version drift across your network.

---

## Why this exists

Tailscale gives you a great admin panel, but if you're already running Prometheus and Grafana for infrastructure observability, you probably want your tailnet health living in the same place as everything else — with alerting, dashboards, and historical data. This exporter bridges that gap.

---

## Metrics

| Metric | Type | Labels | Description |
|---|---|---|---|
| `tailscale_device_online` | Gauge | `id`, `hostname`, `os` | `1` if the device was seen in the last 5 minutes, `0` otherwise |
| `tailscale_device_last_seen_timestamp_seconds` | Gauge | `id`, `hostname`, `os` | Unix timestamp of last seen time |
| `tailscale_device_key_expiry_seconds` | Gauge | `id`, `hostname`, `os` | Seconds until auth key expires (negative = already expired) |
| `tailscale_device_update_available` | Gauge | `id`, `hostname`, `os` | `1` if a Tailscale client update is available |
| `tailscale_devices_total` | Gauge | — | Total devices registered in the tailnet |

---

## Getting started

### Prerequisites

- Go 1.22+
- A Tailscale account with [API access enabled](https://login.tailscale.com/admin/settings/keys)
- Prometheus (local or remote)

### 1. Generate a Tailscale API key

Go to **Admin Console → Settings → Keys** and generate a read-only API key. You'll also need your tailnet name (the domain shown in the admin console, e.g. `yourorg.github`).

### 2. Clone and run

```bash
git clone https://github.com/wguner/tailscale-exporter.git
cd tailscale-exporter

go mod tidy
go run . 
```

Set the required environment variables:

```bash
export TAILSCALE_API_KEY=tskey-api-...
export TAILSCALE_TAILNET=yourorg.github
export LISTEN_ADDR=:9100   # optional, defaults to :9100

go run .
```

Visit `http://localhost:9100/metrics` to see your tailnet metrics.

### 3. Run with Docker

```bash
docker build -t tailscale-exporter .

docker run \
  -e TAILSCALE_API_KEY=tskey-api-... \
  -e TAILSCALE_TAILNET=yourorg.github \
  -p 9100:9100 \
  tailscale-exporter
```

---

## Prometheus configuration

Add a scrape job to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'tailscale'
    static_configs:
      - targets: ['localhost:9100']
```

---

## Example alert rules

```yaml
groups:
  - name: tailscale
    rules:
      - alert: TailscaleDeviceOffline
        expr: tailscale_device_online == 0
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Tailscale device offline: {{ $labels.hostname }}"
          description: "{{ $labels.hostname }} ({{ $labels.os }}) has not been seen for over 10 minutes."

      - alert: TailscaleKeyExpiringSoon
        expr: tailscale_device_key_expiry_seconds < 86400
        labels:
          severity: warning
        annotations:
          summary: "Tailscale key expiring: {{ $labels.hostname }}"
          description: "Auth key for {{ $labels.hostname }} expires in less than 24 hours."
```

---

## Project structure

```
tailscale-exporter/
├── main.go               # HTTP server, /metrics and /healthz endpoints
├── tailscale/
│   └── client.go         # Tailscale API client and Device type
├── collector/
│   └── devices.go        # prometheus.Collector implementation
├── Dockerfile
└── README.md
```

---

## Environment variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `TAILSCALE_API_KEY` | Yes | — | Tailscale API key (read-only scope is sufficient) |
| `TAILSCALE_TAILNET` | Yes | — | Your tailnet name (e.g. `yourorg.github`) |
| `LISTEN_ADDR` | No | `:9100` | Address and port for the metrics server |

---

## Endpoints

| Endpoint | Description |
|---|---|
| `/metrics` | Prometheus metrics |
| `/healthz` | Health check — returns `200 ok` |

---

## License

MIT
