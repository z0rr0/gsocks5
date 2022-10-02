# GSocks5

![Version](https://img.shields.io/github/tag/z0rr0/gsocks5.svg)
![License](https://img.shields.io/github/license/z0rr0/gsocks5.svg)

It's a simple socks5 server based on [go-socks5](https://github.com/armon/go-socks5) 
with authentication and custom dns resolving.

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
        host to listen on (default "127.0.0.1")
  -password string
        password for basic auth
  -port uint
        port to listen on (default 8080)
  -root string
        root directory to serve (default ".")
  -timeout duration
        timeout for requests (default 5s)
  -user string
        username for basic auth
  -version
        show version
```

## License

This source code is governed by a MIT license that can be found
in the [LICENSE](https://github.com/z0rr0/gsocks5/blob/main/LICENSE) file.
