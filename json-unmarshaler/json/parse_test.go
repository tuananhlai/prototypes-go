package json

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	testDataDir := "testdata"

	files, err := os.ReadDir(testDataDir)
	if err != nil {
		t.Fatalf("Error reading test data directory: %v", err)
	}

	for _, file := range files {
		fileName := file.Name()

		if strings.HasPrefix(fileName, "fail") {
			t.Run(fileName, func(t *testing.T) {
				filePath := filepath.Join(testDataDir, fileName)
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("Error reading file %s: %v", fileName, err)
				}

				_, parseErr := Parse(string(content))
				if parseErr == nil {
					t.Errorf("Expected an error for file %s, but got none", fileName)
				}
			})
		} else {
			t.Run(fileName, func(t *testing.T) {
				filePath := filepath.Join(testDataDir, fileName)
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("Error reading file %s: %v", fileName, err)
				}

				_, parseErr := Parse(string(content))
				if parseErr != nil {
					t.Errorf("Unexpected error for file %s: %v", fileName, parseErr)
				}
			})
		}
	}
}
