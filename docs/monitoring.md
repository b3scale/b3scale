# Monitoring

Metrics are exported in a prometheus compatible format under `/metrics`.

!!! Warning
    Make sure to limit visibility to this path so only your prometheus infrastructure can retrieve them. Failure to do so will leak information about frontends and backends to the public.

## Metrics

B3scale exposes the following metrics:

* `b3scale_meeting_attendees`: Number of attendees in the cluster
* `b3scale_meeting_durations`: Duration of meetings in the cluster
* `b3scale_backend_meetings`: Number of meetings per backend
* `b3scale_frontend_attendees`: Number of attendees per frontend

## Scraping the endpoint

The following config will scrape only the b3scale native metrics, skipping over all meta data metrics.

```yaml
scrape_configs:
  - job_name: 'b3scale'
    static_configs:
      - targets: ['<your_host>:443']  # Replace <your_host>
    metrics_path: '/metrics'
    metric_relabel_configs:
      - source_labels: [__name__]
        regex: b3scale_.*
        action: keep
```

 If you want to ingest all metrics, skip the `metric_relabel_configs` section.