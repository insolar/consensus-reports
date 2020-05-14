# Consensus-reports
Tool for consensus reports generation
It has two parts: metric replicator and ...

## Metric replicator
It is a tool, that replicates metrics values from one source (prometheus) to another (webdav) for a proper store.
It uses:
- prometheus query language to get aggregated values: `quantile` and `histogram_quantile`
- webdav server to store json files

All metric's formulas for aggregation are stored like `consensusProperty` structs.

Config sets host for prometheus, webdav auth options, time periods from which to grab metric's values, etc.
 
### Run metric replicator
```
go run cmd/metricreplicator/main.go --cfg=cmd/metricreplicator/config.yml
```

Use `--rm=false` option if you want to save created file locally. Option is `true` by default.

After the work local tmp directory will look like:
```
$ ls /tmp/metricreplicator/
config.json		network_size_10.json	network_size_5.json
```

File for each time range in config.

Remote directory will look the same, except the name will be from config file.
