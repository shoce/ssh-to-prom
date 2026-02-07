# ssh-to-prom

Track ssh connection attempts on your host and expose them as enriched prometheus metrics.

![time chart](img/time_chart.png) ![pie chart](img/pie_chart.png)

## Installation

`go get -u github.com/shoce/ssh-to-prom`

## Usage

`ssh-to-prom -f <ssh_log_file> -m <prometheus port>`

The tool can be somewhat configured by passing some command-line flags when using it.

| Flag | Description  |
|---|---|
| `-f` | Location of the failed authentication log files, typically located at `/var/log/auth.log` or `/var/log/secure` in unix based systems. Default: `/var/log/auth.log`. |
| `-m` | The prometheus port to be exposed. Default: `:2112` |
| `-d` | Enable debug logs. Default: `false` |

## Dependencies

- `papertrail/go-tail` (and `fsnotify/fsnotify` as it's dependency) in order to track and read ssh log file changes
- `prometheus/client_golang` for the prometheus exporter client
- `stretchr/testify` for some readable tests

## Contributing

If you want to contribute, just open an issue or pull request and we can work it out together.
