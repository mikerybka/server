
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
go install github.com/mikerybka/server/cmd/server@latest
go install github.com/mikerybka/server/cmd/updater@latest
```

Copy the following files:

/etc/systemd/system/server.service
```
[Unit]
Description=server

[Service]
ExecStart=/root/go/bin/server

[Install]
WantedBy=multi-user.target
```

/etc/systemd/system/updater.service
```
[Unit]
Description=updater

[Service]
ExecStart=/root/go/bin/updater

[Install]
WantedBy=multi-user.target
```

Run these commands:

```
systemctl daemon-reload
systemctl start server
systemctl status server
systemctl start updater
systemctl status updater
```
