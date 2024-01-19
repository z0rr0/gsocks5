# GSocks5

![Go](https://github.com/z0rr0/gsocks5/workflows/Go/badge.svg)
![Version](https://img.shields.io/github/tag/z0rr0/gsocks5.svg)
![License](https://img.shields.io/github/license/z0rr0/gsocks5.svg)

It's a simple socks5 server based on [go-socks5](https://github.com/armon/go-socks5)
with authentication and custom dns resolving.

It also uses [github.com/serjs/socks5-server](https://github.com/serjs/socks5-server) ideas.

Parameters:

```
Usage of ./gsocks5:
  -auth value
        authentication file
  -concurrent value
        number of concurrent connections in range [1, 10000] (default 100)
  -debug
        debug mode
  -dns string
        custom DNS server
  -host string
        server host
  -port value
        TCP port number to listen on in range [1, 65535] (default 1080)
  -timeout duration
        context timeout (default 5s)
  -version
        show version
```

For custom DNS server you can use:

- google [public DNS](https://developers.google.com/speed/public-dns/): `8.8.8.8`, `8.8.4.4`
- cloudflare [public DNS](https://www.cloudflare.com/learning/dns/what-is-1.1.1.1/): `1.1.1.1`

DockerHub image [z0rr0/gsocks5](https://hub.docker.com/repository/docker/z0rr0/gsocks5).

## Build

Local build:

```sh
# static binary
make build

# docker image
make docker
```

## Run

Local:

```sh
# without host parameter it listens on all interfaces
./gsocks5 -host 127.0.0.1 -port 1080
```

Docker:

```sh
# run container with custom parameters
# -dns can be omitted, then it uses default host DSN resolver
#
# for example there is a file "data/users.txt" with users passwords
# > cat data/users.txt
# user1 password1
# user2 password2

docker run -d \
  --name gsocks5 \
  -u $UID:$UID \
  --log-opt max-size=10m \
  --memory 64m \
  -p 1181:1080 \
  -v $PWD/data:/data/auth:ro \
  --restart always \
  z0rr0/gsocks5:latest -auth /data/auth/users.txt -dns 8.8.8.8
```

## Check

```sh
curl --socks5 <IP>:<PORT> <TARGET_URL>
```

Where

- `<IP>` is your server IP (localhost if you run server locally).
- `<PORT>` is your server port (1080 by default).
- `<TARGET_URL>` is URL you want to check, for example https://fwtf.xyz/short

With authentication:

```sh
curl --socks5 <IP>:<PORT> -U <USER>:<PASSWORD> <TARGET_URL>
```

## License

This source code is governed by a MIT license that can be found
in the [LICENSE](https://github.com/z0rr0/gsocks5/blob/main/LICENSE) file.
