package export

import (
	"fmt"
	"strings"
)

func WindowsPowerShell(in CAInput) (string, error) {
	if strings.TrimSpace(in.CertPEM) == "" {
		return "", fmt.Errorf("cert PEM required")
	}
	name := in.CommonName
	if name == "" {
		name = "Self-CA"
	}
	var b strings.Builder
	b.WriteString("# Self-CA install script — run in PowerShell\n")
	b.WriteString("# Installs root CA to CurrentUser\\Root (no admin required)\n\n")
	b.WriteString("$certPath = Join-Path $env:TEMP \"self-ca.crt\"\n")
	b.WriteString("@'\n")
	b.WriteString(strings.TrimSpace(in.CertPEM))
	b.WriteString("\n'@ | Set-Content -Path $certPath -Encoding ascii\n\n")
	b.WriteString("Import-Certificate -FilePath $certPath -CertStoreLocation Cert:\\CurrentUser\\Root\n")
	b.WriteString("Write-Host \"Installed ")
	b.WriteString(name)
	b.WriteString(" to Trusted Root (Current User)\"\n")
	return b.String(), nil
}

func WindowsBatch(in CAInput) (string, error) {
	if strings.TrimSpace(in.CertPEM) == "" {
		return "", fmt.Errorf("cert PEM required")
	}
	var b strings.Builder
	b.WriteString("@echo off\n")
	b.WriteString("REM Self-CA install — requires admin for LocalMachine\\Root\n")
	b.WriteString("set CERT=%TEMP%\\self-ca.crt\n")
	b.WriteString("(\n")
	for _, line := range strings.Split(strings.TrimSpace(in.CertPEM), "\n") {
		b.WriteString("echo ")
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString(") > \"%CERT%\"\n")
	b.WriteString("certutil -addstore Root \"%CERT%\"\n")
	b.WriteString("echo Installed root CA. Verify in certmgr.msc\n")
	return b.String(), nil
}
