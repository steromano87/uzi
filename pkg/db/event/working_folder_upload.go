package event

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"github.com/steromano87/harkonnen/v1/pkg/db"
	"gorm.io/datatypes"
	"io"
	"os"
	"path/filepath"
	"time"
)

type WorkingFolderUpload struct {
	Timestamp              time.Time
	Origin                 string
	Destination            string
	Base64CompressedFolder string
}

func NewWorkingFolderUpload(origin string, destination string, folderPath string) (*WorkingFolderUpload, error) {
	var compressedBytes bytes.Buffer
	zipWriter := zip.NewWriter(&compressedBytes)
	defer func(zipWriter *zip.Writer) {
		_ = zipWriter.Close()
	}(zipWriter)

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
		f, err := zipWriter.Create(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}

		return nil
	}

	err := filepath.Walk(folderPath, walker)
	if err != nil {
		return nil, err
	}

	event := new(WorkingFolderUpload)
	event.Timestamp = time.Now()
	event.Origin = origin
	event.Destination = destination
	event.Base64CompressedFolder = base64.StdEncoding.EncodeToString(compressedBytes.Bytes())

	return event, nil
}

func (w *WorkingFolderUpload) ToDBEvent() db.Event {
	return db.Event{
		Timestamp:   datatypes.Date(w.Timestamp),
		Origin:      w.Origin,
		Destination: w.Destination,
		Kind:        db.WorkingFolderInitEvent,
		Data:        datatypes.JSONMap{"base64EncodedWorkingFolder": w.Base64CompressedFolder},
	}
}
