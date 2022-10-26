package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ZipFolder(folder string) ([]byte, error) {
	var compressedBytes bytes.Buffer
	zipWriter := zip.NewWriter(&compressedBytes)

	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() {
			_ = file.Close()
		}()

		// Ensure that `path` is not absolute; it should not start with "/".
		// This snippet happens to work because I don't use
		// absolute paths, but ensure your real-world code
		// transforms path into a zip-root relative path.
		f, err := zipWriter.Create(strings.TrimPrefix(path, folder))
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}

		return nil
	}

	err := filepath.Walk(folder, walker)
	if err != nil {
		return nil, err
	}

	err = zipWriter.Close()

	return compressedBytes.Bytes(), err
}

func UnzipFolder(zippedFolder []byte, destinationFolder string) error {
	zipReader, err := zip.NewReader(bytes.NewReader(zippedFolder), int64(len(zippedFolder)))
	if err != nil {
		return err
	}

	// Iterate over zipped files and extract them to working folder
	for _, file := range zipReader.File {
		err = unzipFile(file, destinationFolder)
		if err != nil {
			return err
		}
	}

	return nil
}

func unzipFile(f *zip.File, destinationFolder string) error {
	// Check if file paths are not vulnerable to Zip Slip
	filePath := filepath.Join(destinationFolder, f.Name)
	if !strings.HasPrefix(filePath, filepath.Clean(destinationFolder)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", filePath)
	}

	// Create directory tree
	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			return err
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	// Create a destination file for unzipped content
	destinationFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer func() {
		_ = destinationFile.Close()
	}()

	// Unzip the content of a file and copy it to the destination file
	zippedFile, err := f.Open()
	if err != nil {
		return err
	}
	defer func() {
		_ = zippedFile.Close()
	}()

	if _, err := io.Copy(destinationFile, zippedFile); err != nil {
		return err
	}
	return nil
}
