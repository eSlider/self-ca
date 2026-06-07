# Linux — Install a self-signed root CA

Linux trust stores are **distribution-specific**. There is no single command that works everywhere.

> **What to install:** the **root CA certificate** in PEM format.

## Required file format

| Requirement | Value |
|-------------|-------|
| Format | PEM (preferred) or DER |
| Extension | `.crt` (Debian/Ubuntu **require** `.crt` in `ca-certificates` dir) |
| Content | Single root CA per file (bundles behave differently per distro) |
| Permissions | Root-owned, world-readable (`644`) |

Identify your family:

```bash
cat /etc/os-release
```

---

## Debian / Ubuntu / Mint / Pop!_OS

Uses `update-ca-certificates`.

```bash
# Copy CA — filename MUST end in .crt
sudo cp ca.crt /usr/local/share/ca-certificates/my-self-ca.crt

# Rebuild system bundle
sudo update-ca-certificates
```

Expected output includes: `1 added, 0 removed`.

Verify:

```bash
openssl verify -CAfile /etc/ssl/certs/ca-certificates.crt <(openssl s_client -connect myhost:443 </dev/null 2>/dev/null | openssl x509)
# Or simply:
curl -v https://myhost.local
```

Remove:

```bash
sudo rm /usr/local/share/ca-certificates/my-self-ca.crt
sudo update-ca-certificates --fresh
```

---

## RHEL / Fedora / Rocky / Alma / CentOS Stream

Uses `update-ca-trust` (not `update-ca-certificates` — that command does not exist on Fedora).

```bash
sudo cp ca.crt /etc/pki/ca-trust/source/anchors/my-self-ca.pem
sudo update-ca-trust extract
```

Verify:

```bash
trust list | grep -i "my-self-ca"
curl -v https://myhost.local
```

Remove:

```bash
sudo rm /etc/pki/ca-trust/source/anchors/my-self-ca.pem
sudo update-ca-trust extract
```

Reference: [RHEL — shared system certificates](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/9/html/securing_networks/using-shared-system-certificates_securing-networks)

---

## Arch / Manjaro / SteamOS

Also uses `update-ca-trust`, different anchor path:

```bash
sudo cp ca.crt /etc/ca-certificates/trust-source/anchors/my-self-ca.crt
sudo update-ca-trust
```

Verify:

```bash
trust list | grep -i "my-self-ca"
```

---

## Alpine Linux

Uses `update-ca-certificates` (BusyBox/OpenRC variant):

```bash
sudo cp ca.crt /usr/local/share/ca-certificates/my-self-ca.crt
sudo update-ca-certificates
```

---

## Mozilla Firefox (separate trust store)

Firefox maintains its **own** CA database on all platforms. System trust alone is **not enough** for Firefox.

### Option A — trust OS CAs (recommended)

1. Open `about:config`
2. Set `security.enterprise_roots.enabled` → `true`
3. Restart Firefox

This makes Firefox honor the system trust store you updated above.

### Option B — import manually

1. **Settings → Privacy & Security → Certificates → View Certificates → Authorities → Import**
2. Select `ca.crt`, check **Trust this CA to identify websites**

Reference: [Mozilla — enterprise roots](https://support.mozilla.org/en-US/kb/setting-certificate-authorities-firefox)

---

## Snap / Flatpak applications (TASK)

Snaps and Flatpaks often ship **isolated** CA bundles. System `update-ca-certificates` may not affect them.

| Packaging | Issue | Task |
|-----------|-------|------|
| Snap (e.g. Chromium) | Uses snap-specific cert bundle | Document `snap connect` / refresh; test per snap |
| Flatpak | Uses p11-kit or internal store | Document `flatpak override --user --env=...` if needed |

---

## Known issues and open tasks

| ID | Issue | Impact | Task |
|----|-------|--------|------|
| LIN-1 | **Two update tools** — `update-ca-certificates` vs `update-ca-trust` | Users run wrong command | Service detects distro family and shows correct script |
| LIN-2 | **`.crt` extension required** on Debian | Silent ignore of CA file | Enforce `.crt` in generated install script |
| LIN-3 | **Firefox isolation** | Browser still untrusted | Always document Firefox section |
| LIN-4 | **Java truststore** — JVM uses `$JAVA_HOME/lib/security/cacerts` | Java apps fail | TASK: optional `keytool -importcert` snippet |
| LIN-5 | **Go/Python venv** — some tools bundle own roots | Dev tools ignore system store | Document for developers |
| LIN-6 | **Container hosts** — trust inside Docker/K8s is separate | CI/CD HTTPS fails | TASK: `docs/linux/containers.md` |
| LIN-7 | **WSL** — Windows host and WSL have separate stores | Dev environment confusion | TASK: cross-link Windows doc |

---

## What self-ca should deliver for Linux

- [ ] Download: `ca.crt`
- [ ] Generated install script per family:
  - `install-debian.sh`
  - `install-rhel.sh`
  - `install-arch.sh`
- [ ] Distro detection helper (parse `/etc/os-release`)
- [ ] Firefox `about:config` instructions
- [ ] Uninstall / rollback commands in same script
