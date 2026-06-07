package export_test

import (
	"strings"
	"testing"

	"github.com/eSlider/self-ca/internal/ca"
	"github.com/eSlider/self-ca/internal/export"
)

func sampleCA(t *testing.T) export.CAInput {
	t.Helper()
	generated, err := ca.GenerateCA(ca.CAOptions{CommonName: "Test CA", Organization: "Acme"})
	if err != nil {
		t.Fatalf("GenerateCA: %v", err)
	}
	return export.CAInput{
		CommonName: "Test CA",
		CertPEM:    string(generated.CertPEM),
	}
}

func TestMobileConfig(t *testing.T) {
	in := sampleCA(t)
	out, err := export.MobileConfig(in)
	if err != nil {
		t.Fatalf("MobileConfig: %v", err)
	}
	for _, want := range []string{
		"<?xml",
		"com.apple.security.root",
		"PayloadContent",
		"<data>",
		"Test CA",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in output", want)
		}
	}
}

func TestWindowsPowerShell(t *testing.T) {
	in := sampleCA(t)
	out, err := export.WindowsPowerShell(in)
	if err != nil {
		t.Fatalf("WindowsPowerShell: %v", err)
	}
	if !strings.Contains(out, "Import-Certificate") {
		t.Fatal("expected Import-Certificate")
	}
	if !strings.Contains(out, "BEGIN CERTIFICATE") {
		t.Fatal("expected embedded PEM")
	}
}

func TestWindowsBatch(t *testing.T) {
	in := sampleCA(t)
	out, err := export.WindowsBatch(in)
	if err != nil {
		t.Fatalf("WindowsBatch: %v", err)
	}
	if !strings.Contains(out, "certutil -addstore Root") {
		t.Fatal("expected certutil command")
	}
}

func TestLinuxInstallScript(t *testing.T) {
	in := sampleCA(t)
	out, err := export.LinuxInstallScript(in)
	if err != nil {
		t.Fatalf("LinuxInstallScript: %v", err)
	}
	for _, want := range []string{"#!/bin/sh", "/etc/os-release", "update-ca-certificates", "update-ca-trust"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q", want)
		}
	}
}

func TestAndroidNetworkSecurity(t *testing.T) {
	in := sampleCA(t)
	out, err := export.AndroidNetworkSecurity(in)
	if err != nil {
		t.Fatalf("AndroidNetworkSecurity: %v", err)
	}
	if !strings.Contains(out, "network-security-config") {
		t.Fatal("expected network-security-config")
	}
}

func TestPlatformGuide(t *testing.T) {
	guide := export.PlatformGuide("ios")
	if len(guide) == 0 {
		t.Fatal("expected ios guide steps")
	}
}
