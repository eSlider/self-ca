package export

import (
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"
)

type CAInput struct {
	CommonName string
	CertPEM    string
}

func pemDER(certPEM string) ([]byte, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("invalid certificate PEM")
	}
	return block.Bytes, nil
}

func MobileConfig(in CAInput) (string, error) {
	der, err := pemDER(in.CertPEM)
	if err != nil {
		return "", err
	}
	name := in.CommonName
	if name == "" {
		name = "Self-CA Root"
	}
	payloadUUID := newUUID()
	profileUUID := newUUID()
	certB64 := base64.StdEncoding.EncodeToString(der)

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">` + "\n")
	b.WriteString(`<plist version="1.0">` + "\n")
	b.WriteString("<dict>\n")
	b.WriteString("  <key>PayloadContent</key>\n  <array>\n    <dict>\n")
	writePlistString(&b, "PayloadCertificateFileName", "ca.crt", 6)
	writePlistData(&b, "PayloadContent", certB64, 6)
	writePlistString(&b, "PayloadDescription", "Self-CA root certificate", 6)
	writePlistString(&b, "PayloadDisplayName", name, 6)
	writePlistString(&b, "PayloadIdentifier", "io.selfca.root", 6)
	writePlistString(&b, "PayloadType", "com.apple.security.root", 6)
	writePlistString(&b, "PayloadUUID", payloadUUID, 6)
	b.WriteString("      <key>PayloadVersion</key>\n      <integer>1</integer>\n")
	b.WriteString("    </dict>\n  </array>\n")
	writePlistString(&b, "PayloadDisplayName", name+" Root CA", 2)
	writePlistString(&b, "PayloadIdentifier", "io.selfca.profile", 2)
	b.WriteString("  <key>PayloadRemovalDisallowed</key>\n  <false/>\n")
	writePlistString(&b, "PayloadType", "Configuration", 2)
	writePlistString(&b, "PayloadUUID", profileUUID, 2)
	b.WriteString("  <key>PayloadVersion</key>\n  <integer>1</integer>\n")
	b.WriteString("</dict>\n</plist>\n")
	return b.String(), nil
}

func writePlistString(b *strings.Builder, key, value string, indent int) {
	pad := strings.Repeat(" ", indent)
	b.WriteString(pad + "<key>" + key + "</key>\n")
	b.WriteString(pad + "<string>" + escapeXML(value) + "</string>\n")
}

func writePlistData(b *strings.Builder, key, b64 string, indent int) {
	pad := strings.Repeat(" ", indent)
	b.WriteString(pad + "<key>" + key + "</key>\n")
	b.WriteString(pad + "<data>" + b64 + "</data>\n")
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
