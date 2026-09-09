package tooling

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestScanEscapeHatchesClean(t *testing.T) {
	tempDir := t.TempDir()

	cleanCpp := `#include <iostream>
#include <memory>
#include <vector>
#include <span>

void process(std::span<const int> data) {
    auto ptr = std::make_unique<int>(42);
    std::cout << *ptr << std::endl;
}
`
	if err := os.WriteFile(filepath.Join(tempDir, "main.cpp"), []byte(cleanCpp), 0644); err != nil {
		t.Fatal(err)
	}

	report, err := ScanEscapeHatches(tempDir)
	if err != nil {
		t.Fatalf("unexpected scan error: %v", err)
	}

	if !report.Clean {
		t.Errorf("expected clean report, got clean=false with %d findings", report.TotalFindings)
	}
	if report.TotalFindings != 0 {
		t.Errorf("expected 0 findings, got %d", report.TotalFindings)
	}
}

func TestScanEscapeHatchesViolations(t *testing.T) {
	tempDir := t.TempDir()

	violatingCpp := `#include <iostream>

void dangerous() {
    int* p = reinterpret_cast<int*>(0x1234);
    void* v = (void*)p;
    int* raw = (int*)malloc(sizeof(int));
    free(raw);
    goto cleanup;
cleanup:
    return;
}
`
	if err := os.WriteFile(filepath.Join(tempDir, "violating.cpp"), []byte(violatingCpp), 0644); err != nil {
		t.Fatal(err)
	}

	report, err := ScanEscapeHatches(tempDir)
	if err != nil {
		t.Fatalf("unexpected scan error: %v", err)
	}

	if report.Clean {
		t.Errorf("expected clean=false for violating file")
	}
	if report.UnregisteredFindings < 4 {
		t.Errorf("expected at least 4 unregistered findings, got %d", report.UnregisteredFindings)
	}
}

func TestScanEscapeHatchesRegistry(t *testing.T) {
	tempDir := t.TempDir()

	srcDir := filepath.Join(tempDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	cppCode := `#include <cstddef>
void kernel_entry() {
    int* ptr = reinterpret_cast<int*>(0x4000);
}
`
	if err := os.WriteFile(filepath.Join(srcDir, "kernel.cpp"), []byte(cppCode), 0644); err != nil {
		t.Fatal(err)
	}

	atlasDir := filepath.Join(tempDir, ".atlas")
	if err := os.MkdirAll(atlasDir, 0755); err != nil {
		t.Fatal(err)
	}

	registry := EscapeHatchRegistry{
		SchemaVersion: 1,
		Project:       "systems-test",
		DefaultPolicy: "quarantine",
		EscapeHatches: []EscapeHatchEntry{
			{
				ID:               "EH-001",
				Language:         "cpp",
				File:             "src/kernel.cpp",
				LineRange:        [2]int{1, 10},
				HatchType:        "raw_cast",
				Justification:    "Hardware MMIO register access for low-level HAL",
				SafetyInvariants: "Address 0x4000 mapped by page table in early boot",
				Reviewer:         "systems-lead",
				Status:           "active",
			},
		},
	}

	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(atlasDir, "escape-hatches.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	report, err := ScanEscapeHatches(tempDir)
	if err != nil {
		t.Fatalf("unexpected scan error: %v", err)
	}

	if !report.Clean {
		t.Errorf("expected clean=true because escape hatch was formally registered, got clean=false")
	}
	if report.RegisteredFindings != 1 {
		t.Errorf("expected 1 registered finding, got %d", report.RegisteredFindings)
	}
	if report.UnregisteredFindings != 0 {
		t.Errorf("expected 0 unregistered findings, got %d", report.UnregisteredFindings)
	}
}

func TestScanEscapeHatchesInlineAnnotation(t *testing.T) {
	tempDir := t.TempDir()

	zigCode := `const std = @import("std");

pub fn bufferCast(bytes: [*]u8) *u32 {
    // ATLAS:ESCAPE_HATCH[EH-ZIG-01]
    return @ptrCast(bytes);
}
`
	if err := os.WriteFile(filepath.Join(tempDir, "buffer.zig"), []byte(zigCode), 0644); err != nil {
		t.Fatal(err)
	}

	report, err := ScanEscapeHatches(tempDir)
	if err != nil {
		t.Fatalf("unexpected scan error: %v", err)
	}

	if !report.Clean {
		t.Errorf("expected clean=true with inline annotation, got clean=false")
	}
	if report.RegisteredFindings != 1 {
		t.Errorf("expected 1 registered finding, got %d", report.RegisteredFindings)
	}
	if report.Findings[0].RegistryID != "EH-ZIG-01" {
		t.Errorf("expected registry id EH-ZIG-01, got %s", report.Findings[0].RegistryID)
	}
}

func TestScanEscapeHatchesManagedLanguages(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Python eval
	pyCode := `def execute(code):
    eval(code)
`
	if err := os.WriteFile(filepath.Join(tempDir, "script.py"), []byte(pyCode), 0644); err != nil {
		t.Fatal(err)
	}

	// 2. Kotlin !!
	ktCode := `fun getName(user: User?): String {
    return user!!.name
}
`
	if err := os.WriteFile(filepath.Join(tempDir, "user.kt"), []byte(ktCode), 0644); err != nil {
		t.Fatal(err)
	}

	// 3. PHP eval
	phpCode := `<?php
eval($untrusted_code);
`
	if err := os.WriteFile(filepath.Join(tempDir, "index.php"), []byte(phpCode), 0644); err != nil {
		t.Fatal(err)
	}

	// 4. Bash eval
	shCode := `#!/bin/bash
eval "$1"
`
	if err := os.WriteFile(filepath.Join(tempDir, "runner.sh"), []byte(shCode), 0644); err != nil {
		t.Fatal(err)
	}

	report, err := ScanEscapeHatches(tempDir)
	if err != nil {
		t.Fatalf("unexpected scan error: %v", err)
	}

	if report.Clean {
		t.Errorf("expected clean=false for managed language violations")
	}
	if report.UnregisteredFindings < 4 {
		t.Errorf("expected at least 4 unregistered findings, got %d", report.UnregisteredFindings)
	}
}

func TestScanEscapeHatchesDomain(t *testing.T) {
	tempDir := t.TempDir()

	// 1. HTML javascript: URI and innerHTML
	htmlCode := `<!DOCTYPE html>
<html>
<body>
<a href="javascript:alert(1)">Click</a>
<div id="target"></div>
<script>
document.getElementById('target').innerHTML = "<p>danger</p>";
</script>
</body>
</html>`
	if err := os.WriteFile(filepath.Join(tempDir, "index.html"), []byte(htmlCode), 0644); err != nil {
		t.Fatal(err)
	}

	// 2. CSS !important
	cssCode := `.btn {
    background-color: red !important;
}`
	if err := os.WriteFile(filepath.Join(tempDir, "style.css"), []byte(cssCode), 0644); err != nil {
		t.Fatal(err)
	}

	// 3. SQL SELECT *
	sqlCode := `SELECT * FROM users WHERE active = 1;`
	if err := os.WriteFile(filepath.Join(tempDir, "query.sql"), []byte(sqlCode), 0644); err != nil {
		t.Fatal(err)
	}

	// 4. Dockerfile USER root and :latest
	dockerfileCode := `FROM alpine:latest
USER root
RUN apk update
`
	if err := os.WriteFile(filepath.Join(tempDir, "Dockerfile"), []byte(dockerfileCode), 0644); err != nil {
		t.Fatal(err)
	}

	// 5. Terraform open ingress 0.0.0.0/0
	tfCode := `resource "aws_security_group" "allow_all" {
  name        = "allow_all"
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}`
	if err := os.WriteFile(filepath.Join(tempDir, "main.tf"), []byte(tfCode), 0644); err != nil {
		t.Fatal(err)
	}

	report, err := ScanEscapeHatches(tempDir)
	if err != nil {
		t.Fatalf("unexpected scan error: %v", err)
	}

	if report.Clean {
		t.Errorf("expected clean=false for domain violations")
	}
	if report.UnregisteredFindings < 6 {
		t.Errorf("expected at least 6 unregistered domain findings, got %d", report.UnregisteredFindings)
	}
}
