# cmd_exporter

this project runs commands on the operating system and exports its output and exit code as a prometheus metric.

Use the `cmd_exporter.yml` file to configure the commands, the metrics prefix and the connection stuff.

Put many commands as you need but remember to keep the "keys" unique. keys are use with the metrics prefix to form the metrics name, like this:

```
metrics_prefix + "_" + key_name
```

For example:

```yaml
metrics_prefix: cmd_exporter_prefix
help_message: basic message goes here
...
commands:
  isready_instance1: /usr/pgsql-11/bin/pg_isready -p 9999
  isready_instance2: pg_isready -p 5432
  cmd2: echo foo
```
Will generate the metrics below:
* `cmd_exporter_prefix_isready_instance1`
* `cmd_exporter_prefix_isready_instance2`
* `cmd_exporter_prefix_cmd2`

Each metric is a [GAUGE](https://prometheus.io/docs/concepts/metric_types/#gauge) that contains:
* an `output` label with the command output
* the value with the `exit code`
> for instance:
>```
> # HELP cmd_exporter_prefix_cmd2 basic message goes here
> # TYPE cmd_exporter_prefix_cmd2 gauge
> cmd_exporter_prefix_cmd2{output="foo"} 0
> 
> # HELP cmd_exporter_prefix_isready_instance1 basic message goes here
> # TYPE cmd_exporter_prefix_isready_instance1 gauge
> cmd_exporter_prefix_isready_instance1{output=""} 0
> 
> # HELP cmd_exporter_prefix_isready_instance2 basic message goes here
> # TYPE cmd_exporter_prefix_isready_instance2 gauge
> cmd_exporter_prefix_isready_instance2{output="/tmp:5432 - no response"} 2
> ```

After that configure your prometheus with something like that:
```yaml
scrape_configs:
  - job_name: 'cmd_exporter'
    static_configs:
    - targets: ['localhost:2112']
```
