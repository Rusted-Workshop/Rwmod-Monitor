package archiver

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Archiver struct {
	monitorDir string
}

func NewArchiver(monitorDir string) *Archiver {
	return &Archiver{
		monitorDir: monitorDir,
	}
}

func (a *Archiver) Archive(dirPath string) (string, error) {
	dirName := filepath.Base(dirPath)
	timestamp := time.Now().UTC().Unix()
	archiveName := fmt.Sprintf("%s-%d.rwmod", dirName, timestamp)
	archivePath := filepath.Join(a.monitorDir, archiveName)

	zipFile, err := os.Create(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to create archive: %w", err)
	}
	defer zipFile.Close()

	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		return a.addFileToZip(writer, path, dirPath)
	})

	if err != nil {
		os.Remove(archivePath)
		return "", fmt.Errorf("failed to archive directory: %w", err)
	}

	return archivePath, nil
}

func (a *Archiver) addFileToZip(writer *zip.Writer, filePath, baseDir string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	relPath, err := filepath.Rel(baseDir, filePath)
	if err != nil {
		return err
	}

	zipEntry, err := writer.Create(relPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(zipEntry, file)
	return err
}
