# Device validation checklist

Manual checklist for validating platform install guides against real hardware. Run after changing exporters or docs.

## Prerequisites

- [ ] self-ca running (`go run . -config config.yml`)
- [ ] At least one CA created via UI or API
- [ ] Test device on same LAN (or use port-forward / tunnel)
- [ ] Record self-ca version / git commit

## iOS / iPadOS

Device: _______________  iOS version: _______________

- [ ] Download `ca.mobileconfig` from Platform install tab
- [ ] Install profile — appears under Settings → General → VPN & Device Management
- [ ] Enable full trust: Settings → General → About → Certificate Trust Settings
- [ ] Visit HTTPS site signed by issued leaf cert — no warning
- [ ] Notes / blockers: _______________

## Android

Device: _______________  Android version: _______________

- [ ] Download `ca.crt` and install via Settings → Security → Encryption & credentials → Install a certificate → CA certificate
- [ ] Chrome trusts HTTPS site signed by leaf cert
- [ ] Third-party app **without** network security config does **not** trust (expected)
- [ ] Dev app with `network_security_config.xml` snippet trusts (optional)
- [ ] Notes / blockers: _______________

## Windows

Device: _______________  Windows version: _______________

- [ ] Run `install-ca.ps1` — cert in `Cert:\CurrentUser\Root`
- [ ] Edge/Chrome trusts HTTPS site
- [ ] Optional: `install-ca.bat` as admin → LocalMachine Root
- [ ] Notes / blockers: _______________

## Linux (Debian family)

Distro: _______________

- [ ] Run `install-ca.sh` — completes without error
- [ ] `curl https://<leaf-host>` succeeds with system curl
- [ ] Firefox: test with `security.enterprise_roots.enabled=true`
- [ ] Notes / blockers: _______________

## Linux (RHEL family)

Distro: _______________

- [ ] `install-ca.sh` uses `update-ca-trust extract`
- [ ] System curl trusts leaf cert
- [ ] Notes / blockers: _______________

## macOS

Device: _______________  macOS version: _______________

- [ ] Install `ca.mobileconfig` or `ca.crt`
- [ ] Keychain Access → CA → Always Trust for SSL
- [ ] Safari trusts HTTPS site
- [ ] Notes / blockers: _______________

## Sign-off

| Reviewer | Date | Result |
|----------|------|--------|
| | | Pass / Fail |

Update platform README TASK tables if any step fails.
