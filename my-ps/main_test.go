package main

import (
	"strings"
	"testing"
)

func TestParseProcessStatus(t *testing.T) {
	t.Run("parses all known fields", func(t *testing.T) {
		input := strings.NewReader(`Name:	bash
State:	S (sleeping)
PPid:	1
Uid:	1000	1000	1000	1000
Threads:	1
VmRSS:	4096 kB
`)
		stat, err := parseProcessStatus(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stat.name != "bash" {
			t.Errorf("name: got %q, want %q", stat.name, "bash")
		}
		if stat.state != "S (sleeping)" {
			t.Errorf("state: got %q, want %q", stat.state, "S (sleeping)")
		}
		if stat.ppid != "1" {
			t.Errorf("ppid: got %q, want %q", stat.ppid, "1")
		}
		if stat.realUID != "1000" {
			t.Errorf("realUID: got %q, want %q", stat.realUID, "1000")
		}
		if stat.threads != "1" {
			t.Errorf("threads: got %q, want %q", stat.threads, "1")
		}
		if stat.vmRSS != "4096 kB" {
			t.Errorf("vmRSS: got %q, want %q", stat.vmRSS, "4096 kB")
		}
	})

	t.Run("ignores unknown fields", func(t *testing.T) {
		input := strings.NewReader(`Name:	myproc
Unknown:	ignored
PPid:	42
`)
		stat, err := parseProcessStatus(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stat.name != "myproc" {
			t.Errorf("name: got %q, want %q", stat.name, "myproc")
		}
		if stat.ppid != "42" {
			t.Errorf("ppid: got %q, want %q", stat.ppid, "42")
		}
	})

	t.Run("skips malformed lines without colon", func(t *testing.T) {
		input := strings.NewReader(`Name:	myproc
this line has no colon
PPid:	7
`)
		stat, err := parseProcessStatus(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stat.name != "myproc" {
			t.Errorf("name: got %q, want %q", stat.name, "myproc")
		}
		if stat.ppid != "7" {
			t.Errorf("ppid: got %q, want %q", stat.ppid, "7")
		}
	})

	t.Run("empty Uid value does not panic", func(t *testing.T) {
		input := strings.NewReader("Uid:\n")
		stat, err := parseProcessStatus(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stat.realUID != "" {
			t.Errorf("realUID: got %q, want empty string", stat.realUID)
		}
	})

	t.Run("missing fields return zero values", func(t *testing.T) {
		input := strings.NewReader("Name:\tminimal\n")
		stat, err := parseProcessStatus(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stat.ppid != "" || stat.realUID != "" || stat.threads != "" || stat.vmRSS != "" {
			t.Errorf("expected empty fields for missing keys, got %+v", stat)
		}
	})
}

func TestGetProcessStatSelf(t *testing.T) {
	stat, err := getProcessStat("self")
	if err != nil {
		t.Fatalf("unexpected error reading /proc/self/status: %v", err)
	}
	if stat.name == "" {
		t.Error("name should not be empty for the current process")
	}
	if stat.ppid == "" {
		t.Error("ppid should not be empty for the current process")
	}
	if stat.realUID == "" {
		t.Error("realUID should not be empty for the current process")
	}
}
