package cheatsheettest_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	cheatsheettest "github.com/tuananhlai/prototypes/cheatsheet-test"
)

func TestMarshalJSON_Golden(t *testing.T) {
	type payload struct {
		Name         string            `json:"name"`
		Age          int               `json:"age"`
		Tags         []string          `json:"tags"`
		Attrs        map[string]string `json:"attrs"`
		IsActive     bool              `json:"is_active"`
		CreatedAt    time.Time         `json:"created_at"`
		Country      *string           `json:"country"`
		AnnualSalary *int              `json:"annual_salary,omitempty"`
	}

	// Make the input deterministic:
	// - Avoid maps if you need stable key order across Go versions/encoders.
	// - If you use maps, you may still get stable output in practice, but it's not
	//   a contract you should rely on for golden files.
	in := payload{
		Name: "Jonathan",
		Age:  30,
		Tags: []string{"go", "testing"},
		Attrs: map[string]string{
			"team": "platform",
			"role": "dev",
		},
		IsActive:     true,
		CreatedAt:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Country:      nil,
		AnnualSalary: nil,
	}

	got, err := cheatsheettest.MarshalJSON(in)
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}

	assertGolden(t, got, "marshal_json")
}

func assertGolden(t *testing.T, got []byte, goldenFileName string) {
	t.Helper()
	goldenFilePath := filepath.Join("testdata", goldenFileName+".golden")

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(goldenFilePath), 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(goldenFilePath, got, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
	}

	want, err := os.ReadFile(goldenFilePath)
	if err != nil {
		t.Fatalf("read golden %s: %v (tip: run UPDATE_GOLDEN=1 go test ./... to create/update)", goldenFilePath, err)
	}

	if string(got) != string(want) {
		t.Fatalf("golden mismatch\n--- got:\n%s\n--- want:\n%s", got, want)
	}
}
