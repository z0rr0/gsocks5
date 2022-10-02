# GSocks5

![Go](https://github.com/z0rr0/gsocks5/workflows/Go/badge.svg)
![Version](https://img.shields.io/github/tag/z0rr0/gsocks5.svg)
![License](https://img.shields.io/github/license/z0rr0/gsocks5.svg)

It's a simple socks5 server based on [go-socks5](https://github.com/armon/go-socks5)
with authentication and custom dns resolving.

It also uses [github.com/serjs/socks5-server](https://github.com/serjs/socks5-server) ideas.

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

For custom DNS server you can use:

- google [public DNS](https://developers.google.com/speed/public-dns/): `8.8.8.8`, `8.8.4.4`
- cloudflare [public DNS](https://www.cloudflare.com/learning/dns/what-is-1.1.1.1/): `1.1.1.1`
- mullvad public DNS from list [github.com/mullvad/dns-blocklists](https://github.com/mullvad/dns-blocklists#custom-dns-entries)

DockerHub image [z0rr0/gsocks5](https://hub.docker.com/repository/docker/z0rr0/gsocks5).

## Build

Local build:

```sh
go build -ldflags "-X main.Tag=`git tag --sort=version:refname | tail -1`" .
```

Docker image:

```sh
docker build -t z0rr0/gsocks5 .
```

## Run

Local:

```sh
chmod u+x gsocks5
# show parameters
./gsocks5 -help
```

Docker:

```sh
docker run -d --name gsocks5 -p 1080:1080 z0rr0/gsocks5
```

With authentication:

```sh
# for example there is a directory ".local" with users passwords
# > cat .local/users.txt
# user1 pass1
# user2 pass2

docker run -d --name gsocks5 -p 1080:1080 -v $PWD/.local:/data/auth z0rr0/gsocks5 -auth /data/auth/users.txt
```

With authentication, google DNS, custom port 30080 and timeout 30 seconds:

```sh
docker run -d --name gsocks5 --restart always \
  -p 30080:1080 \
  -v $PWD/.local:/data/auth \
  z0rr0/gsocks5 -auth /data/auth/users.txt \
  -dns 8.8.8.8 \
  -timeout 30
````

## Check

```sh
curl --socks5 <IP>:<PORT> https://am.i.mullvad.net/ip
```

Where

- `<IP>` is your server IP (localhost if you run server locally).
- `<PORT>` is your server port (1080 by default).

With authentication:

```sh
curl --socks5 <IP>:<PORT> -U <user>:<password> https://am.i.mullvad.net/ip
```

## License

This source code is governed by a MIT license that can be found
in the [LICENSE](https://github.com/z0rr0/gsocks5/blob/main/LICENSE) file.
