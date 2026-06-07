package export

import (
	"fmt"
	"strings"
)

func LinuxInstallScript(in CAInput) (string, error) {
	if strings.TrimSpace(in.CertPEM) == "" {
		return "", fmt.Errorf("cert PEM required")
	}
	name := in.CommonName
	if name == "" {
		name = "self-ca"
	}
	safeName := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '-'
	}, name)

	var b strings.Builder
	b.WriteString("#!/bin/sh\n")
	b.WriteString("# Self-CA Linux install script — auto-detects distro family\n")
	b.WriteString("set -eu\n\n")
	b.WriteString("CA_FILE=\"$(mktemp)\"\n")
	b.WriteString("trap 'rm -f \"$CA_FILE\"' EXIT\n")
	b.WriteString("cat > \"$CA_FILE\" <<'SELF_CA_PEM'\n")
	b.WriteString(strings.TrimSpace(in.CertPEM))
	b.WriteString("\nSELF_CA_PEM\n\n")
	b.WriteString("if [ -f /etc/os-release ]; then\n  . /etc/os-release\nfi\n\n")
	b.WriteString("ID=\"${ID:-unknown}\"\n")
	b.WriteString("case \"$ID\" in\n")
	b.WriteString("  debian|ubuntu|linuxmint|pop|elementary|zorin|kali|raspbian)\n")
	b.WriteString(fmt.Sprintf("    sudo cp \"$CA_FILE\" \"/usr/local/share/ca-certificates/%s.crt\"\n", safeName))
	b.WriteString("    sudo update-ca-certificates\n")
	b.WriteString("    ;;\n")
	b.WriteString("  fedora|rhel|centos|rocky|almalinux|ol|amzn)\n")
	b.WriteString(fmt.Sprintf("    sudo cp \"$CA_FILE\" \"/etc/pki/ca-trust/source/anchors/%s.pem\"\n", safeName))
	b.WriteString("    sudo update-ca-trust extract\n")
	b.WriteString("    ;;\n")
	b.WriteString("  arch|manjaro|garuda|steamos|endeavouros)\n")
	b.WriteString(fmt.Sprintf("    sudo cp \"$CA_FILE\" \"/etc/ca-certificates/trust-source/anchors/%s.crt\"\n", safeName))
	b.WriteString("    sudo update-ca-trust\n")
	b.WriteString("    ;;\n")
	b.WriteString("  alpine)\n")
	b.WriteString(fmt.Sprintf("    sudo cp \"$CA_FILE\" \"/usr/local/share/ca-certificates/%s.crt\"\n", safeName))
	b.WriteString("    sudo update-ca-trust\n")
	b.WriteString("    ;;\n")
	b.WriteString("  *)\n")
	b.WriteString("    echo \"Unsupported distro: $ID — see docs/linux/README.md\"\n")
	b.WriteString("    exit 1\n")
	b.WriteString("    ;;\n")
	b.WriteString("esac\n\n")
	b.WriteString("echo \"Installed ")
	b.WriteString(name)
	b.WriteString(" — Firefox may need security.enterprise_roots.enabled=true\"\n")
	return b.String(), nil
}
