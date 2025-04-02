#!/bin/sh

# Use `curl -fsSL https://raw.githubusercontent.com/mikerybka/server/refs/heads/main/scripts/setup.sh | sh`

curl -fsSL "https://raw.githubusercontent.com/mikerybka/server/refs/heads/main/builds/linux/amd64/server" -o "/bin/server" && \
    chmod +x /bin/server && \
    curl -fsSL https://tailscale.com/install.sh | sh && \
    tailscale up && \
    curl -fsSL https://raw.githubusercontent.com/mikerybka/server/refs/heads/main/config/netplan/60-floating-ip.yaml -o /etc/netplan/60-floating-ip.yaml && \
    chmod 600 /etc/netplan/60-floating-ip.yaml && \
    netplan apply && \
    /bin/server setup
