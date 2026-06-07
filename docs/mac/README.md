# macOS — Install a self-signed root CA

This guide covers trusting a **custom root CA** on macOS so Safari, Chrome, and most system APIs (Security.framework / Network.framework) accept your certificates.

> **What to install:** the **root CA certificate**. The leaf/server cert is not sufficient unless you also trust the issuing CA.

## Required file format

| Requirement | Value |
|-------------|-------|
| Format | PEM (`.pem`, `.crt`) or DER (`.cer`) |
| Keychain target | **System** (all users, admin) or **login** (current user) |
| Trust setting | **Always Trust** for SSL/TLS |
| Private key | **Never** distribute the CA private key |

---

## Manual install — Keychain Access (GUI)

### Step 1 — Download and open

1. Download `ca.crt` from the self-ca web UI.
2. Double-click the file — **Keychain Access** opens.

### Step 2 — Choose keychain

When prompted:

- **login** — current user only (no admin password for import, but trust change may prompt)
- **System** — all users (**requires admin password**)

For shared Macs or system-wide dev servers, prefer **System**.

### Step 3 — Set trust (critical)

Installing alone is **not enough**. You must set trust:

1. Open **Keychain Access** (Spotlight → "Keychain Access")
2. Select **login** or **System** keychain → category **Certificates**
3. Find your CA (match Common Name)
4. Double-click → expand **Trust**
5. Set **Secure Sockets Layer (SSL)** → **Always Trust**
6. Close window → enter password to save

Without SSL trust set to Always Trust, browsers show certificate errors even though the cert is present.

### Step 4 — Verify

```bash
# Should show your CA
security find-certificate -a -c "Your CA Common Name" ~/Library/Keychains/login.keychain-db

# Test HTTPS
curl -v https://myhost.local
open https://myhost.local   # Safari
```

---

## CLI install — security command

### Current user (login keychain)

```bash
# Add certificate
security add-trusted-cert -d -r trustRoot -k ~/Library/Keychains/login.keychain-db ca.crt
```

Flags:

- `-d` — add to admin cert store (for login keychain, marks as trusted root)
- `-r trustRoot` — trust as root CA
- `-k` — target keychain

### System keychain (all users, requires admin)

```bash
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ca.crt
```

Reference: `man security` → `add-trusted-cert`

---

## CLI install — split add + trust (if add-trusted-cert fails)

Some macOS versions behave better with explicit trust settings:

```bash
# Import only
security import ca.crt -k ~/Library/Keychains/login.keychain-db -T /usr/bin/curl

# Then set trust via Keychain Access GUI, or:
sudo security add-trusted-cert -d -r trustRoot -p ssl -k /Library/Keychains/System.keychain ca.crt
```

---

## Configuration profile (optional, TASK)

Like iOS, macOS accepts `.mobileconfig` with `com.apple.security.root` payload. Useful for one-click install via MDM or manual profile install.

Reference: [Apple — Certificates payload settings](https://support.apple.com/guide/deployment/certificates-payload-settings-dep91d2eb26/web)

On macOS, profiles still may require user approval and admin password for System keychain.

---

## Firefox (separate store)

Firefox does not use the macOS system keychain by default.

1. `about:config` → `security.enterprise_roots.enabled` = `true`, **or**
2. **Settings → Privacy & Security → Certificates → View Certificates → Authorities → Import**

See [Linux Firefox section](../linux/README.md#mozilla-firefox-separate-trust-store) for details (same Firefox behavior on all OSes).

---

## Known issues and open tasks

| ID | Issue | Impact | Task |
|----|-------|--------|------|
| MAC-1 | **Trust not set by default** — import ≠ trust | #1 macOS support issue | UI must highlight "Always Trust" step |
| MAC-2 | **System vs login keychain** — wrong keychain = other users unaffected | Shared machine confusion | Explain scope in download page |
| MAC-3 | **Gatekeeper / quarantine** — downloaded files flagged | Double-click blocked | TASK: serve over HTTPS; `xattr -d com.apple.quarantine` doc |
| MAC-4 | **SIP / MDM** — managed Macs may block custom roots | Install fails silently | Document enterprise MDM path |
| MAC-5 | **Firefox / Chrome** — Chrome uses system store; Firefox may not | Split behavior | Cross-link Firefox docs |
| MAC-6 | **Java / Node native** — separate trust stores possible | Dev tool failures | TASK: keytool snippet |
| MAC-7 | **Certificate pinning in apps** | CA trust insufficient | Out of scope |

---

## What self-ca should deliver for macOS

- [ ] Download: `ca.crt` (PEM)
- [ ] Download: `install-ca.sh` wrapping `security add-trusted-cert`
- [ ] Download: optional `ca.mobileconfig`
- [ ] UI checklist: import → set Always Trust → verify in Safari
- [ ] Distinguish login vs System keychain in script flags
