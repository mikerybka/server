
## Setup

```
ssh root@<ip_address>
```

```
echo 'export PATH=$PATH:/usr/local/go/bin' > /etc/profile.d/go.sh
```
Log out and log back in to enable to new environment.

```
wget https://go.dev/dl/go1.19.5.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.19.5.linux-amd64.tar.gz
rm go1.19.5.linux-amd64.tar.gz
go version
```

```
go install github.com/mikerybka/server/cmd/reverseproxy@latest
```

Copy the following files:

/etc/systemd/system/reverseproxy.service
```
[Unit]
Description=reverseproxy

[Service]
ExecStart=/root/go/bin/reverseproxy

[Install]
WantedBy=multi-user.target
```

Run these commands:

```
systemctl daemon-reload
systemctl start reverseproxy
systemctl status reverseproxy
```

## Config

Domain to port mapping is done in `/etc/reverseproxy/config.json`.
