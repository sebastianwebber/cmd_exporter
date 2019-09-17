# cmd_exporter

this project runs a command and exports its output and exit code as a prometheus metric.

Use the `cmd_exporter.yml` file to configure the command, the metric and connection stuff.

After that configure your prometheus with something like that:
```yaml
scrape_configs:
  - job_name: 'cmd_exporter'
    static_configs:
    - targets: ['localhost:2112']
```
