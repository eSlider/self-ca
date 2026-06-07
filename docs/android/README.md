# Android — Install a self-signed root CA

This guide covers trusting a **custom root CA** on Android phones and tablets.

> **What to install:** the **root CA certificate** (`ca.crt`). Leaf/server certs alone are not enough for a private CA.

## Required file format

| Requirement | Value |
|-------------|-------|
| Format | PEM or DER |
| Extension | `.crt` or `.cer` (some OEM file pickers are picky) |
| Content | **Root CA only** — X.509 v3 with `CA:TRUE` |
| Private key | **Never** ship the CA private key |

---

## Manual install (user workflow)

Path names vary by OEM/Android version. Below is the **Google Pixel / stock Android 13+** flow; Samsung and others use similar wording under **Security** or **Biometrics and security**.

### Step 1 — Copy CA to the device

- Download from the self-ca web UI in Chrome, or
- Transfer via USB / Files app / email attachment.

Save to **Downloads** or another location the cert installer can read.

### Step 2 — Install as CA certificate

1. **Settings → Security & privacy** (or **Security**)
2. **More security settings** (if present)
3. **Encryption & credentials**
4. **Install a certificate** → **CA certificate**
5. Read the warning (*"Your network activity may be monitored"*) → **Install anyway**
6. Pick the `ca.crt` file
7. Confirm name (defaults to subject CN)

Android stores this in the **user credential store**, not the system store.

### Step 3 — Verify in browser

1. Open **Chrome** and visit your HTTPS URL.
2. Confirm the connection is trusted (no `NET::ERR_CERT_AUTHORITY_INVALID`).

If Chrome still fails, see [Known issues](#known-issues-and-open-tasks).

### Step 4 — Verify installation location

**Settings → Security → Encryption & credentials → Trusted credentials → User**

Your CA should appear under the **User** tab (not **System**).

---

## System-wide trust (advanced, usually not available)

Apps on Android 7+ (API 24+) **do not trust user-installed CAs by default**. Only the **system CA store** is trusted unless an app opts in via [Network Security Config](https://developer.android.com/privacy-and-security/security-config).

| Method | Requires | Works for most apps? |
|--------|----------|----------------------|
| User CA install (above) | Nothing | **No** — mainly Chrome/browser; many apps ignore user CAs |
| MDM / work profile | Enterprise MDM | **Yes** — can push to system store on managed devices |
| Rooted device + Magisk module | Root | **Yes** — fragile, breaks on OTAs |
| App `network_security_config.xml` | App developer change | **Per-app only** |

**TASK for self-ca docs/UI:** clearly state that user CA install is sufficient for **browsers and debugging**, not for arbitrary third-party apps.

---

## Alternative: per-app trust (developer-owned apps)

If you control the Android app, add `res/xml/network_security_config.xml`:

```xml
<?xml version="1.0" encoding="utf-8"?>
<network-security-config>
  <base-config cleartextTrafficPermitted="false">
    <trust-anchors>
      <certificates src="system" />
      <certificates src="user" />
    </trust-anchors>
  </base-config>
</network-security-config>
```

Reference in `AndroidManifest.xml`:

```xml
<application
    android:networkSecurityConfig="@xml/network_security_config"
    ... />
```

Or embed the CA as `@raw/my_ca` in the same config (works without user install).

---

## Known issues and open tasks

| ID | Issue | Impact | Task |
|----|-------|--------|------|
| AND-1 | **User vs system store** — default install goes to user store | Most native apps reject the CA | Document loudly; offer MDM instructions for enterprise |
| AND-2 | **Chrome on Android 7+** — may not trust user CAs in all versions/OEM builds | Browser testing fails | Verify on target devices; link to Chrome flags only as last resort |
| AND-3 | **Persistent notification** — user CA triggers "Network may be monitored" on some versions | User confusion | Explain in UI — expected Android behavior |
| AND-4 | **Screen lock required** — installing CA may require PIN/pattern | Blocks install on some devices | Mention in prerequisites |
| AND-5 | **Android 14+ APEX system certs** — system store is in immutable APEX modules | Root/Magisk workarounds harder | Do not promise system CA install without MDM/root |
| AND-6 | **Certificate pinning** — banking/social apps pin keys | CA trust irrelevant | Out of scope; document |
| AND-7 | **OEM UI differences** — Samsung, Xiaomi, etc. hide cert settings | Users can't find menu | TASK: OEM-specific screenshots in docs |
| AND-8 | **No `.mobileconfig` equivalent** — no universal one-tap profile format | Harder UX than iOS | TASK: investigate Samsung Knox / Android Enterprise work profile APIs |

---

## What self-ca should deliver for Android

- [ ] Download: `ca.crt` (PEM, `.crt` filename)
- [ ] Optional: DER-encoded `.cer` for picky file pickers
- [ ] In-app checklist with OEM-specific notes
- [ ] Developer snippet: `network_security_config.xml` + `@raw/ca` for owned apps
- [ ] Enterprise doc link: Android Enterprise / MDM CA deployment (future)
