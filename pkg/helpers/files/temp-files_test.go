package files

import (
	"github.com/pkg/errors"
	"log"
	"os"
	"testing"
)

func TestGarbageCollectTemporaryFiles(t *testing.T) {
	tests := []struct {
		name     string
		numFiles int
		n        int
		want     int
		err      error
	}{
		{
			name:     "WithValidInputs",
			numFiles: 5,
			n:        2,
			want:     3,
			err:      nil,
		},
		{
			name:     "WithNoMatchingFiles",
			numFiles: 0,
			n:        2,
			want:     0,
			err:      nil,
		},
		{
			name:     "WithLessThanNFiles",
			numFiles: 1,
			n:        2,
			want:     0,
			err:      nil,
		},
		{
			name:     "WithExactlyNFiles",
			numFiles: 2,
			n:        2,
			want:     0,
			err:      nil,
		},
		{
			name:     "WithDirectories",
			numFiles: 5,
			n:        2,
			want:     3,
			err:      nil,
		},
		{
			name:     "WithZeroN",
			numFiles: 5,
			n:        0,
			want:     5,
			err:      nil,
		},
		{
			name:     "WithNegativeN",
			numFiles: 5,
			n:        -1,
			want:     5,
			err:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, _, err := initTest(t, "", "temp", tt.numFiles, "file*")
			if err != nil {
				log.Fatalf("Failed to create temp directory: %v", err)
			}

			defer func(path string) {
				err := os.RemoveAll(path)
				if err != nil {
					log.Fatalf("Failed to remove temp directory: %v", err)
				}
			}(tempDir)
			// Call GarbageCollectTemporaryFiles
			deletedFiles, err := GarbageCollectTemporaryFiles(tempDir, "file*", tt.n)
			if err != tt.err {
				t.Fatalf("GarbageCollectTemporaryFiles failed: %v, want: %v", err, tt.err)
			}

			// Check the number of deleted files
			if got := len(deletedFiles); got != tt.want {
				t.Errorf("GarbageCollectTemporaryFiles = %v, want %v", got, tt.want)
			}
		})
	}
}

func initTest(t *testing.T, dir string, pattern string, numFiles int, filePattern string) (string, []string, error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp(dir, pattern)
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	fileNames := []string{}
	// Create temporary files
	for i := 0; i < numFiles; i++ {
		tempFile, err := os.CreateTemp(tempDir, filePattern)
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		_ = tempFile.Close()
		fileNames = append(fileNames, tempFile.Name())
	}
	return tempDir, fileNames, err
}

func TestGarbageCollectTemporaryFilesWithNonExistentDirectory(t *testing.T) {
	tempDir := "/nonexistent"
	n := 2
	_, err := GarbageCollectTemporaryFiles(tempDir, "file*", n)
	if err == nil {
		t.Fatalf("GarbageCollectTemporaryFiles with invalid directory should fail")
	}
}

func TestGarbageCollectTemporaryFilesWithInvalidMask(t *testing.T) {
	tempDir, _, err := initTest(t, "", "temp", 5, "file*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			log.Fatalf("Failed to remove temp directory: %v", err)
		}
	}(tempDir)

	fs, err := GarbageCollectTemporaryFiles(tempDir, "\x00", 2)
	if err != nil {
		t.Fatalf("GarbageCollectTemporaryFiles failed: %v, want: %v", err, os.ErrInvalid)
	}
	if len(fs) != 0 {
		t.Fatalf("GarbageCollectTemporaryFiles failed: %v, want: %v", len(fs), 0)
	}
}

func TestGarbageCollectTemporaryFilesWithUnreadableDirError(t *testing.T) {
	tempDir, _, err := initTest(t, "", "temp", 5, "file*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			log.Fatalf("Failed to remove temp directory: %v", err)
		}
	}(tempDir)

	n := 2
	err = os.Chmod(tempDir, 0000)
	if err != nil {
		log.Fatalf("Failed to change permissions on temp directory: %v", err)
	} // remove all permissions
	_, err = GarbageCollectTemporaryFiles(tempDir, "file*", n)
	if !errors.Is(err, os.ErrPermission) {
		t.Fatalf("GarbageCollectTemporaryFiles failed: %v, want: %v", err, os.ErrPermission)
	}
	err = os.Chmod(tempDir, 0755)
	if err != nil {
		log.Fatalf("Failed to change permissions on temp directory: %v", err)
	} // restore permissions
}
