// utils_test.go

package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

func TestUnzip(t *testing.T) {
	// Define the source and destination paths
	src := "testdata/test.zip"
	dest := "testdata/unzipped"

	// Clean up any previous test data
	defer func() {
		os.RemoveAll(dest)
	}()

	// Unzip the test.zip file
	if err := utils.Unzip(src, dest); err != nil {
		t.Fatalf("error unzipping file: %v", err)
	}

	// Check if the unzipped files exist
	expectedFiles := []string{"file1.txt", "file2.txt"}
	for _, file := range expectedFiles {
		filePath := filepath.Join(dest, file)
		if _, err := os.Stat(filePath); err != nil {
			t.Errorf("expected file %s not found: %v", file, err)
		}
	}
}

func TestUntar(t *testing.T) {
	// Define the source and destination paths
	src := "testdata/test.tar.gz"
	dest := "testdata/untarred"

	// Clean up any previous test data
	defer func() {
		os.RemoveAll(dest)
	}()

	// Untar the test.tar.gz file
	if err := utils.Untar(src, dest); err != nil {
		t.Fatalf("error untarring file: %v", err)
	}

	// Check if the untarred files exist
	expectedFiles := []string{"file1.txt", "file2.txt"}
	for _, file := range expectedFiles {
		filePath := filepath.Join(dest, file)
		if _, err := os.Stat(filePath); err != nil {
			t.Errorf("expected file %s not found: %v", file, err)
		}
	}
}
