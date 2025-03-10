# Athom Exporter

---

A Prometheus exporter for collecting metrics from the **AthomTech Tastmota Switzerland Plug** based on Tasmota (might also work with other Tasmota devices).

## Features 🚀
- **Exports metrics in OpenMetrics format**: provides energy usage metrics to scrape for prometheus or other OpenMetrics compatible services like Grafana Alloy.
- **Graceful error handling**: Provides meaningful HTTP responses and logs errors.
- **Environment variable support**: Customize config and bind address using environment variables.

---

## Scrape Configuration

Athom exporter configuration works just like blackbox exporter configuration.
Prometheus will scrape the exporter and the target of the exporter is defined via a query parameter.

### Prometheus Example Config

```yaml
scrape_configs:
  - job_name: 'athom'
    metrics_path: /metrics
    static_configs:
      - targets:
        - http://plug.host
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: athom-exporter:5573
```

### Grafana Alloy Example Config

```alloy
discovery.relabel "athom" {
  targets = [
      {
        address = "http://plug.host",
      },
  ]

  rule {
    action = "replace"
    source_labels = ["address"]
    target_label = "__param_target"
  }

  rule {
    action = "replace"
    source_labels = ["__param_target"]
    target_label = "instance"
  }

  rule {
    action = "replace"
    target_label = "__address__"
    replacement = "athom.default.svc.cluster.local:5573"
  }
}

prometheus.scrape "athom" {
  clustering {
    enabled = true
  }

  scrape_interval = "60s"
  targets    = discovery.relabel.athom.output
  forward_to = [prometheus.remote_write.mimir.receiver]
}
```

## Installation 🛠️

### From Source
Build the binary
```bash
git clone git@github.com:lirionex/athom-exporter.git

cd athom-exporter

go build -o athom-exporter
```

Run with env vars
```bash
export BIND_ADDRESS=":5573"
./athom-exporter
```
### Docker

```bash
docker run -d \
  -p 5573:5573 \
  --name athom-exporter \
  ghcr.io/lirionex/athom-exporter/athom-exporter:latest
```

### Docker Compose

```yaml
services:
    ical-proxy:
        image: ghcr.io/lirionex/athom-exporter/athom-exporter:latest
        container_name: athom-exporter
        ports:
          - "5573:5573"
```

### Kubernetes

```yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: athom
  labels:
    app: athom
spec:
  replicas: 1
  selector:
    matchLabels:
      app: athom
  template:
    metadata:
      labels:
        app: athom
    spec:
      containers:
        - name: athom
          image: ghcr.io/lirionex/athom-exporter/athom-exporter:latest
          ports:
            - containerPort: 5573
---
apiVersion: v1
kind: Service
metadata:
  name: athom
spec:
  selector:
    app: athom
  ports:
    - protocol: TCP
      port: 5573
      targetPort: 5573
```