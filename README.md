
## Setup

```sh
ssh root@<ip_address>
```

### Install git

```sh
apt update
apt install git
```

### Install Go

```sh
echo 'export PATH=$PATH:/usr/local/go/bin' > /etc/profile.d/go.sh
echo 'export PATH=$PATH:/root/go/bin' > /etc/profile.d/go.sh
```

Log out and log back in to enable to new environment.

```sh
wget https://go.dev/dl/go1.19.5.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.19.5.linux-amd64.tar.gz
rm go1.19.5.linux-amd64.tar.gz
go version
```

### Install custom infrastructure

```sh
GOPROXY=direct go install github.com/mikerybka/server/cmd/appmand@latest
GOPROXY=direct go install github.com/mikerybka/server/cmd/appman@latest
GOPROXY=direct go install github.com/mikerybka/server/cmd/reverseproxy@latest
```

### Copy the following files:

#### /etc/systemd/system/appmand.service
```
[Unit]
Description=appmand

[Service]
ExecStart=/root/go/bin/appmand

[Install]
WantedBy=multi-user.target
```

#### /etc/systemd/system/reverseproxy.service
```
[Unit]
Description=reverseproxy

[Service]
ExecStart=/root/go/bin/reverseproxy

[Install]
WantedBy=multi-user.target
```

### Start services

```sh
systemctl daemon-reload
systemctl start appmand
systemctl start reverseproxy
```

## Config

### Add an app

```sh
appman add-app <appID>
```

- `appID` is the go url of the app. For example, `github.com/mikerybka/server/cmd/reverseproxy`.

On success, a port is returned.

### Set a domain

```sh
appman set-domain <domain> <port>
```

- `domain` is a domain to host the app. For example, `brass.dev`.
- `port` is the port to listen on. For example, `8080`.
