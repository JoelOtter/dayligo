package dayligo

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	backupFileName = "backup.daylio"
)

func ReadBackupFromFile(path string) (*Backup, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open daylio zip file: %w", err)
	}
	tmpDir, err := os.MkdirTemp("", "dayligo")
	if err != nil {
		return nil, fmt.Errorf("failed to create tempdir: %w", err)
	}
	var fileContents []byte
	for _, file := range reader.File {
		filePath := filepath.Join(tmpDir, file.Name)
		fmt.Println("Unzipping file", filePath)
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
				return nil, fmt.Errorf("failed to create directory %s: %w", filePath, err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return nil, fmt.Errorf("failed to create directory structure: %w", err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return nil, fmt.Errorf("failed to open destination file %s: %w", filePath, err)
		}

		fileInArchive, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file in archive %s: %w", filePath, err)
		}
		fileInArchiveContents, err := ioutil.ReadAll(fileInArchive)
		if err != nil {
			return nil, fmt.Errorf("failed to read file in archive %s: %w", filePath, err)
		}
		if file.Name == backupFileName {
			fileContents = fileInArchiveContents
		}

		if _, err := io.Copy(dstFile, bytes.NewReader(fileInArchiveContents)); err != nil {
			return nil, fmt.Errorf("failed to copy file during unzip: %w", err)
		}

		if err := dstFile.Close(); err != nil {
			return nil, fmt.Errorf("failed to close destination file: %w", err)
		}
		if err := fileInArchive.Close(); err != nil {
			return nil, fmt.Errorf("failed to close file in archive: %w", err)
		}
	}
	if err := reader.Close(); err != nil {
		return nil, fmt.Errorf("failed to close daylio file: %w", err)
	}
	jsonBytes, err := base64.StdEncoding.DecodeString(string(fileContents))
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64-encoded JSON: %w", err)
	}
	fmt.Printf("JSON bytes length: %d\n", len(jsonBytes))
	var backup Backup
	if err := json.Unmarshal(jsonBytes, &backup); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(jsonBytes, &rawMap); err != nil {
		return nil, fmt.Errorf("failed to parse internal raw map: %w", err)
	}
	backup.rawMap = rawMap
	backup.tempDirPath = tmpDir
	return &backup, nil
}

// Close cleans up created temp directory.
func (b *Backup) Close() error {
	fmt.Println("Deleting temp dir", b.tempDirPath)
	if err := os.RemoveAll(b.tempDirPath); err != nil {
		return fmt.Errorf("failed to delete tempdir %s: %w", b.tempDirPath, err)
	}
	return nil
}

func (b *Backup) WriteToFile(path string) error {
	for key, val := range getKeyMapping(b) {
		var err error
		b.rawMap[key], err = json.Marshal(val)
		if err != nil {
			return fmt.Errorf("failed to marshal %s to JSON: %w", key, err)
		}
	}
	jsonBackup, err := json.Marshal(b.rawMap)
	if err != nil {
		return fmt.Errorf("failed to marshal raw map to JSON: %w", err)
	}
	base64Bytes := bytes.NewBuffer(nil)
	encoder := base64.NewEncoder(base64.StdEncoding, base64Bytes)
	if _, err := encoder.Write(jsonBackup); err != nil {
		return fmt.Errorf("failed to encode base 64 bytes: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("failed to close base64 encoder: %w", err)
	}

	dstFile, err := os.OpenFile(filepath.Join(b.tempDirPath, backupFileName), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to open destination file %s: %w", backupFileName, err)
	}
	if _, err := dstFile.Write(base64Bytes.Bytes()); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}
	if err := dstFile.Close(); err != nil {
		return fmt.Errorf("failed to close backup file: %w", err)
	}

	archive, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}
	writer := zip.NewWriter(archive)

	if err := filepath.Walk(b.tempDirPath, func(path string, info fs.FileInfo, topErr error) error {
		fmt.Println("Crawling", path)
		if topErr != nil {
			return topErr
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open source file: %w", err)
		}
		internalPath, err := filepath.Rel(b.tempDirPath, path)
		if err != nil {
			return fmt.Errorf("failed to work out internal path in archive: %w", err)
		}
		f, err := writer.Create(internalPath)
		if err != nil {
			return fmt.Errorf("failed to create destination file: %w", err)
		}
		if _, err := io.Copy(f, file); err != nil {
			return fmt.Errorf("failed to copy file: %w", err)
		}

		if err := file.Close(); err != nil {
			return fmt.Errorf("failed to close source file: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to compress temp dir to archive: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close archive writer: %w", err)
	}
	if err := archive.Close(); err != nil {
		return fmt.Errorf("failed to close archive file: %w", err)
	}

	return nil
}

func getKeyMapping(backup *Backup) map[string]interface{} {
	return map[string]interface{}{
		"dayEntries":       &backup.DayEntries,
		"goals":            &backup.Goals,
		"goalEntries":      &backup.GoalEntries,
		"goalSuccessWeeks": &backup.GoalSuccessWeeks,
	}
}
