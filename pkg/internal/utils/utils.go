// utils/file.go

package utils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Unzip extracts a zip file to a destination directory.
func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

// Untar extracts a tar.gz file to a destination directory.
func Untar(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Ensure the directory is created
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			// Ensure the parent directory exists or is created
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			outFile, err := os.Create(target)
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close() // Close here instead of defer to handle error properly
				return err
			}
			outFile.Close() // Successful, so we can close the file here
		default:
			fmt.Printf("Unsupported tar type: %v\n", header.Typeflag)
		}
	}

	return nil
}

func GenerateSha256Hash[T any](data T) string {
	// Convert data to a string assuming it implements fmt.Stringer or similar
	// For structs, you might want to serialize them to JSON or another stable format
	dataString := fmt.Sprintf("%v", data)

	// Compute SHA-256 hash of the data string
	hash := sha256.Sum256([]byte(dataString))

	// Return the hexadecimal string representation of the hash
	return hex.EncodeToString(hash[:])
}

func GenerateUniqueHash() string {
	// Combine the current time and random data for the hash input
	currentTime := time.Now().UnixNano()
	randomBytes := make([]byte, 16) // 128 bits of random data
	_, err := rand.Read(randomBytes)
	if err != nil {
		// Handle random generator failure
		// In a real application, consider how to handle this error properly.
		panic("random number generator failed")
	}

	// Convert both pieces of data to byte slices and concatenate
	hashInput := append([]byte(fmt.Sprintf("%d", currentTime)), randomBytes...)

	// Compute SHA-256 hash
	hash := sha256.Sum256(hashInput)

	// Return the hexadecimal string representation of the hash
	return hex.EncodeToString(hash[:])
}

// DecodeJSON is a convenience function that decodes JSON from an io.Reader into a destination object.
func DecodeJSON(r io.Reader, dest interface{}) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(dest)
}
