package utils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {

	// Create a temporary directory for test data.
	err := os.MkdirAll("testdata", 0755)
	if err != nil {
		panic(err)
	}

	// Create necessary test data files.
	createTestDataFiles()

}

func teardown() {
	// Remove the temporary directory and its contents.
	err := os.RemoveAll("testdata")
	if err != nil {
		panic(err)
	}
}

func createTestDataFiles() {

	// Create a test.zip file with mock content.
	zipFile, err := os.Create("testdata/test.zip")
	if err != nil {
		panic(err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Write files to the ZIP archive
	for _, fileName := range []string{"file1.txt", "file2.txt"} {
		fileWriter, err := zipWriter.Create(fileName)
		if err != nil {
			panic(err)
		}
		_, err = fileWriter.Write([]byte("This is a test file."))
		if err != nil {
			panic(err)
		}
	}

	// Prepare to create a test.tar.gz file with mock content.
	tarGzFileName := "testdata/test.tar.gz"
	tarGzFile, err := os.Create(tarGzFileName)
	if err != nil {
		panic(err)
	}
	defer tarGzFile.Close()

	gzWriter := gzip.NewWriter(tarGzFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Create and write files to the TAR archive before compressing it
	for _, fileName := range []string{"file1.txt", "file2.txt"} {
		header := &tar.Header{
			Name: fileName,
			Mode: 0600,
			Size: int64(len("This is a test file.")),
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			panic(err)
		}
		if _, err := tarWriter.Write([]byte("This is a test file.")); err != nil {
			panic(err)
		}
	}

}
