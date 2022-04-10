# Observability 

# Первый запуск Prometheus

Перейдем в директорию `prometheus-first` и запустим:

```bash
docker run -d -p 9090:9090 \
    --name gb-prometheus \
    -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml \
    -v $(pwd)/prometheus.rules.yml:/etc/prometheus/prometheus.rules.yml \
    -v $(pwd)/alert.rules.yml:/etc/prometheus/alert.rules.yml \
    prom/prometheus:v2.25.0
```