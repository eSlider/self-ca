# iOS / iPadOS — Install a self-signed root CA

This guide covers trusting a **custom root CA** on iPhone and iPad so Safari, Mail, and other system TLS clients accept certificates signed by your CA.

> **What to install:** the **root CA certificate** (`ca.crt`), not the server/leaf certificate alone. Browsers validate the chain up to a trusted anchor.

## Required file format

| Requirement | Value |
|-------------|-------|
| Format | PEM (Base64, `-----BEGIN CERTIFICATE-----`) |
| Extension | `.crt`, `.cer`, or `.pem` (avoid `.mobileconfig` unless you ship a profile) |
| Content | **Root CA only** — must have `CA:TRUE` in Basic Constraints |
| Private key | **Never** distribute the CA private key to devices |

The service should offer a direct download link (HTTPS or AirDrop-friendly) and optionally a `.mobileconfig` profile (see [Automated delivery](#automated-delivery-task)).

---

## Manual install (user workflow)

### Step 1 — Transfer the CA certificate to the device

Choose one:

1. **Safari download** — open the CA download URL from the self-ca web UI.
2. **AirDrop** — send `ca.crt` from a Mac.
3. **Email / Files** — attach the certificate and open it on the device.

iOS shows *"This website is trying to download a configuration profile"* or prompts to install a certificate.

### Step 2 — Install the profile

1. Open **Settings**.
2. Tap **Profile Downloaded** (top banner), or go to **General → VPN & Device Management** (older iOS: **General → Profiles & Device Management**).
3. Tap the profile, review details, tap **Install**.
4. Enter passcode / confirm.

At this point the CA is installed but **not yet trusted for SSL/TLS**.

### Step 3 — Enable full trust (mandatory)

Apple requires an explicit second step for manually installed root CAs:

1. **Settings → General → About → Certificate Trust Settings**
2. Under **Enable Full Trust for Root Certificates**, toggle **ON** for your CA.
3. Confirm the security warning.

Without this step, HTTPS sites signed by your CA will still fail with certificate errors.

### Step 4 — Verify

1. Open **Safari** and navigate to your HTTPS endpoint (e.g. `https://myserver.local`).
2. Confirm no certificate warning.
3. Optional: **Settings → General → About → Certificate Trust Settings** — your CA should appear as fully trusted.

---

## Automated delivery (TASK)

The web service should generate an Apple configuration profile (`.mobileconfig`) containing:

- Payload type: `com.apple.security.root`
- Payload content: Base64-encoded DER of the root CA

Reference: [Apple — Certificates payload settings](https://support.apple.com/guide/deployment/certificates-payload-settings-dep91d2eb26/web)

**Benefits:** single-tap install, clearer UX than raw `.crt`.

**Limitation:** even with `.mobileconfig`, **manual full-trust toggle is still required** unless deployed via MDM/Apple Configurator (see Known issues).

---

## Known issues and open tasks

| ID | Issue | Impact | Task |
|----|-------|--------|------|
| IOS-1 | **Two-step trust** — install profile + enable Certificate Trust Settings | Users skip step 3 and think install failed | UI must show step 3 prominently; block "done" until acknowledged |
| IOS-2 | **MDM auto-trust** — only MDM/Configurator profiles get automatic SSL trust | Manual `.mobileconfig` still needs step 3 | Document MDM path for enterprise; don't claim one-click trust |
| IOS-3 | **Supervised vs unsupervised** — unsupervised devices show extra warnings | UX friction for BYOD | Show Apple’s exact warning text in docs/UI |
| IOS-4 | **App-specific trust** — some apps use certificate pinning | CA trust does not help pinned apps | Document that pinning bypass is out of scope |
| IOS-5 | **Wi‑Fi / VPN profiles** — 802.1X may need intermediate certs in same profile | Enterprise Wi‑Fi fails with root-only | TASK: support intermediate chain in mobileconfig bundle |
| IOS-6 | **Certificate expiry** — expired CA breaks all trust silently over time | Production outages | TASK: expiry alerts + re-issue workflow in service |

---

## What self-ca should deliver for iOS

- [ ] Download: `ca.crt` (PEM)
- [ ] Download: `ca.mobileconfig` (root CA profile)
- [ ] QR code → Safari download URL
- [ ] Step-by-step checklist in UI (install → trust → verify)
- [ ] Deep link hint: `App-Prefs:root=General&path=About/CERTIFICATE_TRUST_SETTINGS` (may change between iOS versions — verify before relying on it)
