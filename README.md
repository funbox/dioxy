[About](#about) • [Installation](#installation) • [Getting started](#getting-started) • [Usage](#usage) • [License](#license)

# About

`dioxy` is simple aggregating proxy for MQTT broker metrics in JSON format. We develop and maintain this software 
mostly for collecting CO2, temperature and humidity metrics from MT8060 in pair with ESP8266 microcontroller. This project is a
part of health measurement system for office workers.

## How it works

When you run `dioxy`, it connects to the MQTT broker, collects messages from it
(using the topic prefix) and stores them in memory. Each next-coming message replaces
the previous one and updates the time when it was received.

`dioxy` also monitors orphaned metrics and periodically cleans up an obsolete
metrics that have not been updated for some time. You can also configure these
options as well.

`dioxy` provides a simple HTTP server to let external systems grub these aggregated
metrics from the status page. Every time you request this page, application
serializes Golang memory struct into JSON message and responses with 200 OK.

Otherwise if something went wrong, the application returns
500 Internal Server Error with empty reply.

## MQTT message format

MQTT message should be in following format:

```
/{topic_prefix}/{metrics} {value}
```

JSON representation is:

```
{
  "{topic_prefix}": {
    "metrics": "{metrics}",
    "value": "{value}",
    "updated_at": "{updated_at}"
  }
}
```

`topic_prefix` is the string that begins each MQTT message. `metrics` is the measurement name and `value` its value.
Every measurement also includes `updated_at` field - particular time when the last message received from.

## Installation

### From prebuilt package for RHEL7/CentOS7

You can find RPM packages attached to releases on [Release page](https://github.com/gongled/dioxy/releases).

### From the source code

Fetch Go dependencies for compiling binary from the source code.

```shell
make deps
```

Build binary `dioxy`.

```shell
make all
```

Done.

## Getting started

Set up `dioxy` configuration to aggregate MQTT metrics.

```shell
[mqtt]

  # MQTT broker IP
  ip: mqtt.example.tld

  # MQTT broker port
  port: 1883 

  # MQTT username to authenticate to
  user: username

  # MQTT password to authenticate to
  password: keepinsecret

  # MQTT topic to listen to
  topic: /devices/MT8060/#

[store]

  # Time in seconds to delete obsolete data (TTL)
  ttl: 86400

  # Time in seconds between looking for an obsolete data
  clean-interval: 1500

[http]

  # HTTP server IP
  ip:

  # HTTP server port
  port: 33407

[log]

  # Log file dir
  dir: /var/log/dioxy/

  # Path to log file
  file: {log:dir}/dioxy.log

  # Log permissions
  perms: 600

  # Default log level (debug/info/warn/error/crit)
  level: debug
```

Launch systemd unit and make sure it will be launched after reboot.

```shell
[sudo] systemctl start dioxy.service
[sudo] systemctl enable dioxy.service
```

Check if metrics are collecting well.

```
curl -sL http://127.0.0.1:33407/ | jq .
```

Done.

## Restrictions

We do not provide any mechanism to assosiate metrics with aggregating function.
`dioxy` replaces the last value every time we receive a next one.

We rely that your monitoring system is processing data on their own if needed.

## Usage

```
Usage: dioxy {options}

Options

  --config, -c file .. Path to configuraion file
  --no-color, -nc .... Disable colors in output
  --help, -h ......... Show this help message
  --version, -v ...... Show version
```

## License

Released under the MIT license (see [LICENSE](LICENSE))

[![Sponsored by FunBox](https://funbox.ru/badges/sponsored_by_funbox_grayscale.svg)](https://funbox.ru)
