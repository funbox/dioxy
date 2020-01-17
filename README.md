[About](#about) • [Installation](#installation) • [Getting started](#getting-started) • [Usage](#usage) • [License](#license)

# About

`dioxy` is simple aggregating proxy for MQTT broker metrics in JSON format. We develop and maintain this software 
mostly for collecting CO2, temperature and humidity metrics from MT8060 in pair with ESP8266 microcontroller. This project is a
part of health measurement system for office workers.

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

Set up dioxy configuration to aggregate MQTT metrics.

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
