# Metrics & Observability Service

Prometheus, Grafana, and Alertmanager configuration.

**Development Plan**: `../../docs/service-plans/08-metrics-observability-service.md`

## Quick Start
```bash
docker-compose -f ../../docker/docker-compose.dev.yml up prometheus grafana alertmanager
```

## Access
- Grafana: http://localhost:3001 (admin/admin)
- Prometheus: http://localhost:9090
- Alertmanager: http://localhost:9093

## Configuration
- `prometheus/prometheus.yml` - Scrape configs
- `alertmanager/alertmanager.yml` - Alert routing
- `grafana/provisioning/` - Dashboards and datasources
