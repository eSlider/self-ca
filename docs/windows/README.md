# Windows — Install a self-signed root CA

This guide covers trusting a **custom root CA** on Windows 10/11 so Edge, Chrome, and most WinHTTP/Schannel clients accept your certificates.

> **What to install:** the **root CA certificate**. Install into **Trusted Root Certification Authorities**, not Personal/My.

## Required file format

| Requirement | Value |
|-------------|-------|
| Format | PEM or DER |
| Extension | `.crt` or `.cer` (PEM content with `.crt` works) |
| Store target | `LocalMachine\Root` (all users) or `CurrentUser\Root` (current user only) |
| Private key | **Never** distribute CA private key via download UI |

---

## Manual install — GUI (recommended for end users)

### Step 1 — Download the CA

Download `ca.crt` from the self-ca web UI.

### Step 2 — Open Certificate Import Wizard

**Option A — from File Explorer**

1. Double-click `ca.crt`.
2. Click **Install Certificate…**
3. Choose store scope:
   - **Local Machine** — all users (requires Administrator)
   - **Current User** — only your account (no admin)

**Option B — MMC snap-in**

1. `Win + R` → `certmgr.msc` (current user) or `certlm.msc` (local machine, admin)
2. Expand **Trusted Root Certification Authorities → Certificates**
3. Right-click → **All Tasks → Import…**
4. Select `ca.crt`, finish wizard

### Step 3 — Select the correct store

When prompted for certificate store:

- Select **Place all certificates in the following store**
- Browse → **Trusted Root Certification Authorities**
- **Do not** leave it in Personal — that won't establish trust for TLS

### Step 4 — Verify

1. Open **Edge** or **Chrome** → visit your HTTPS URL.
2. Or run in PowerShell:

```powershell
# Replace thumbprint after import — find it in certmgr
Get-ChildItem Cert:\LocalMachine\Root | Where-Object { $_.Subject -like "*Your CA Name*" }
```

---

## CLI install — PowerShell (admin for LocalMachine)

Run **PowerShell as Administrator** for machine-wide trust:

```powershell
Import-Certificate `
  -FilePath "$env:USERPROFILE\Downloads\ca.crt" `
  -CertStoreLocation Cert:\LocalMachine\Root
```

Current-user only (no admin):

```powershell
Import-Certificate `
  -FilePath "$env:USERPROFILE\Downloads\ca.crt" `
  -CertStoreLocation Cert:\CurrentUser\Root
```

Reference: [Import-Certificate](https://learn.microsoft.com/en-us/powershell/module/pki/import-certificate)

---

## CLI install — certutil (admin)

```cmd
certutil -addstore Root "C:\path\to\ca.crt"
```

For enterprise AD environments:

```cmd
certutil -enterprise -f -addstore Root "C:\path\to\ca.crt"
```

Reference: [certutil -addstore](https://learn.microsoft.com/en-us/windows-server/administration/windows-commands/certutil)

---

## Group Policy (enterprise, TASK)

For domain-joined PCs, deploy via GPO:

**Computer Configuration → Windows Settings → Security Settings → Public Key Policies → Trusted Root Certification Authorities**

Import the CA once; all domain machines receive it.

TASK: document GPO export/import steps in a future `docs/windows/enterprise.md`.

---

## Known issues and open tasks

| ID | Issue | Impact | Task |
|----|-------|--------|------|
| WIN-1 | **Wrong store** — users import into Personal/My | TLS still fails | UI wizard must pre-select Root store; docs emphasize this |
| WIN-2 | **Admin required for LocalMachine** — Current User scope only affects one account | Confusion on shared PCs | Explain scope in download page |
| WIN-3 | **Intermediate vs root** — some tools import CA into Intermediate Authorities | Chain validation fails | TASK: detect and warn if cert lacks CA:TRUE |
| WIN-4 | **Firefox separate store** — Firefox uses NSS, not Schannel | Firefox untrusted even after system install | Link to [Firefox steps](../linux/README.md#mozilla-firefox-separate-trust-store) |
| WIN-5 | **Corporate policy** — IT may block user root installs | Install silently fails or is reverted | Document enterprise GPO path |
| WIN-6 | **`.pfx` confusion** — users download server cert+key bundle | Private key exposure risk | Never offer PFX for CA trust; only `.crt` for root |
| WIN-7 | **Certificate pinning** — some apps bypass system store | CA install insufficient | Document out-of-scope |

---

## What self-ca should deliver for Windows

- [ ] Download: `ca.crt` (PEM)
- [ ] Optional: `install-ca.ps1` helper (CurrentUser + LocalMachine modes)
- [ ] Optional: `install-ca.bat` wrapping certutil for admin users
- [ ] Post-install verification: PowerShell one-liner with expected subject/thumbprint
- [ ] Clear warning: LocalMachine requires Administrator
