# XConnect Gateway

Independent Linux relay/service for XConnect Zero. It is deliberately separate from the XConnect One controlled-client CLI and from XConnect App.

The runtime performs `join → session renewal → signed gateway config sync → WireGuard/Xray apply → ACK`. XConnect Zero `accounts` is the only configuration authority. The binary rejects unsigned, expired, cross-network, or cross-gateway configuration.

## Runtime boundary

- `xconnect-gateway`: owns enrollment, protected local state, signed configuration verification, generated runtime files, apply and ACK.
- external `WireGuard` and `Xray`: data plane processes.
- GitOps: non-sensitive UAT topology and release selection.
- Vault: TLS key material and other environment secrets; expected UAT path is `kv/data/uat/xconnect-one`.
- `accounts`: per-user isolated networks, devices, invites, policy and signed configuration.
- `portal`: user-facing `/panel/xconnect-zero` BFF/UI. It never receives device credentials or private keys.

## Linux host

Install `wireguard-tools`, `xray`, the systemd units under `packaging/systemd`, and the released binary at `/usr/local/bin/xconnect-gateway`. Provision the TLS certificate and key at `/etc/xconnect-gateway/tls.crt` and `/etc/xconnect-gateway/tls.key` from Vault without writing either value to Git.

```sh
sudo xconnect-gateway diagnose
sudo xconnect-gateway join --gateway-id gw-uat-1 'xconnect://join/REDACTED?controller=https%3A%2F%2Faccounts-uat.onwalk.net'
sudo xconnect-gateway up
sudo systemctl enable --now xconnect-gateway-sync.timer
```

The invitation is one-time sensitive data. Do not pass it as a GitHub Actions input or print it in logs.
