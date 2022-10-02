# GSocks5

![Version](https://img.shields.io/github/tag/z0rr0/gsocks5.svg)
![License](https://img.shields.io/github/license/z0rr0/gsocks5.svg)

It's a simple socks5 server based on [go-socks5](https://github.com/armon/go-socks5)
with authentication and custom dns resolving.

It also uses [github.com/serjs/socks5-server](https://github.com/serjs/socks5-server) ideas.

## Build

```sh
go build -ldflags "-X main.Tag=`git tags | tail -1`" .
```

## Run

```sh
...
```

Parameters

```
Usage of ./gsocks5:
  -auth string
        authentication file
  -debug
        debug mode
  -dns string
        custom DNS server
  -host string
        server host
  -port uint
        port to listen on (default 1080)
  -timeout uint
        context timeout (seconds) (default 5)
  -version
        show version
```

## License

This source code is governed by a MIT license that can be found
in the [LICENSE](https://github.com/z0rr0/gsocks5/blob/main/LICENSE) file.
