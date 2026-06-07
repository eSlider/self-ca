package export

import (
	"fmt"
	"strings"
)

func AndroidNetworkSecurity(in CAInput) (string, error) {
	if strings.TrimSpace(in.CertPEM) == "" {
		return "", fmt.Errorf("cert PEM required")
	}
	name := in.CommonName
	if name == "" {
		name = "Self-CA"
	}
	var b strings.Builder
	b.WriteString("<!-- Self-CA Android network security config snippet -->\n")
	b.WriteString("<!-- For app developers: save PEM as res/raw/self_ca.pem and reference below -->\n")
	b.WriteString("<!-- User CA install alone does NOT trust most apps since Android 7 — see docs/android/README.md -->\n\n")
	b.WriteString("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n")
	b.WriteString("<network-security-config>\n")
	b.WriteString("  <base-config cleartextTrafficPermitted=\"false\">\n")
	b.WriteString("    <trust-anchors>\n")
	b.WriteString("      <certificates src=\"system\" />\n")
	b.WriteString("      <certificates src=\"user\" />\n")
	b.WriteString("      <!-- Optional: embed CA in app -->\n")
	b.WriteString("      <!-- <certificates src=\"@raw/self_ca\" /> -->\n")
	b.WriteString("    </trust-anchors>\n")
	b.WriteString("  </base-config>\n")
	b.WriteString("</network-security-config>\n\n")
	b.WriteString("<!-- PEM certificate (")
	b.WriteString(name)
	b.WriteString("): -->\n")
	b.WriteString("<!--\n")
	b.WriteString(strings.TrimSpace(in.CertPEM))
	b.WriteString("\n-->\n")
	return b.String(), nil
}

// PlatformGuide returns install steps for UI display.
func PlatformGuide(platform string) []string {
	switch platform {
	case "ios":
		return []string{
			"Download ca.mobileconfig or ca.crt",
			"Install profile in Settings",
			"Settings → General → About → Certificate Trust Settings → enable full trust",
		}
	case "android":
		return []string{
			"Download ca.crt",
			"Settings → Security → Install CA certificate",
			"Note: most apps ignore user CAs — see network_security_config.xml for dev apps",
		}
	case "windows":
		return []string{
			"Download install-ca.ps1 (Current User) or install-ca.bat (admin)",
			"Run script — cert must land in Trusted Root, not Personal",
		}
	case "linux":
		return []string{
			"Download install-ca.sh",
			"chmod +x install-ca.sh && ./install-ca.sh",
			"Firefox: set security.enterprise_roots.enabled=true",
		}
	case "mac":
		return []string{
			"Download ca.mobileconfig or ca.crt",
			"Double-click → Keychain Access",
			"Set Always Trust for SSL on the CA certificate",
		}
	default:
		return nil
	}
}
