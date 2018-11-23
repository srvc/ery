# ery
[![Build Status](https://travis-ci.com/srvc/ery.svg?branch=master)](https://travis-ci.com/srvc/ery)
[![Go Report Card](https://goreportcard.com/badge/github.com/srvc/ery)](https://goreportcard.com/report/github.com/srvc/ery)
[![license](https://img.shields.io/github/license/srvc/ery.svg)](./LICENSE)
[![Go project version](https://badge.fury.io/go/github.com%2Fsrvc%2Fery.svg)](https://badge.fury.io/go/github.com%2Fsrvc%2Fery)

:mag: Discover services in local.

[![asciicast](https://asciinema.org/a/212433.svg)](https://asciinema.org/a/212433)


## Usage
### For local processes
You should create a configuration file. you can generate it via `ery init` command.

```sh
# when your current working directory is "~/src/github.com/yourname/awesomeapp",
# a default hostname uses `awesomeapp.yourname.ery`.
ery init
```

any commands prefixed by `ery` sets `PORT` to environment variables at random.

```sh
ery rails s
```

### For docker containers
`ery` reads exposed ports automatically. You have only to set a hostname through label of the container.

```sh
# you can access the rails server with "http://yourapp.ery"
$ docker run \
  --rm \
  --label tools.srvc.ery.hostname=yourapp.ery \
  -p 80:12345 \
  yourapp \
  rails s -p 80
```


## Installation
1. Install `ery`
    - linux
        - ```
          curl -Lo ery https://github.com/srvc/ery/releases/download/v0.0.1/ery_linux_amd64 && chmod +x ery && sudo mv ery /usr/local/bin
          ```
    - macOS
        - ```
          curl -Lo ery https://github.com/srvc/ery/releases/download/v0.0.1/ery_darwin_amd64 && chmod +x ery && sudo mv ery /usr/local/bin
          ```
    - build yourself
        - `go get github.com/srvc/ery/cmd/ery`
1. Configure nameserver
    - linux
        - `sudo sh -c 'echo "nameserver 127.0.0.1" >> /etc/resolv.conf'`
    - macOS
        - `sudo sh -c 'echo "nameserver 127.0.0.1" >> /etc/resolver/ery'`
        - if you wanna use other TLDs, you should replace "ery" to others on above command
1. Add loopback address aliases (macOS only)
    - `for i in $(seq 1 255); do /sbin/ifconfig lo0 alias 127.0.0.$i; done`
1. Register as a startup process
    - `sudo ery daemon install`
    - `sudo ery daemon start`


## Author
- Masayuki Izumi ([@izumin5210](https://github.com/izumin5210))


## License
ery is licensed under the MIT License. See [LICENSE](./LICENSE)
