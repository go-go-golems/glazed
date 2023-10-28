package files

import (
	"os"
	"path/filepath"
	"sort"
)

type fileInfo struct {
	os.FileInfo
	path string // Full path to the file
}

type byModTime []fileInfo

func (s byModTime) Len() int           { return len(s) }
func (s byModTime) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byModTime) Less(i, j int) bool { return s[i].ModTime().Before(s[j].ModTime()) }

// GarbageCollectTemporaryFiles deletes the oldest files based on the mask, leaving only 'n' files.
func GarbageCollectTemporaryFiles(tempDir string, mask string, n int) ([]string, error) {
	var deletedFiles []string

	// Get files that match the glob pattern in the temporary directory
	matchingFiles, err := filepath.Glob(filepath.Join(tempDir, mask))
	if err != nil {
		return nil, err
	}

	// Check if we have read access to tempDir
	d, err := os.Stat(tempDir)
	if err != nil {
		return nil, err
	}
	if d.Mode().Perm()&0444 == 0 {
		return nil, &os.PathError{Op: "GarbageCollectTemporaryFiles", Path: tempDir, Err: os.ErrPermission}
	}

	// Check if we have write access to tempDir
	if d.Mode().Perm()&0222 == 0 {
		return nil, &os.PathError{Op: "GarbageCollectTemporaryFiles", Path: tempDir, Err: os.ErrPermission}
	}

	var fileInfos []fileInfo
	for _, fileName := range matchingFiles {
		fi, err := os.Stat(fileName)
		if err != nil {
			// If we can't obtain the FileInfo, skip this file
			continue
		}

		// Check if it's a regular file (not a directory)
		if !fi.IsDir() {
			fileInfos = append(fileInfos, fileInfo{FileInfo: fi, path: fileName})
		}
	}

	// Sort the files based on modification time (oldest first)
	sort.Sort(byModTime(fileInfos))

	// Delete the oldest files, stopping when only 'n' remain
	for i, fileInfo := range fileInfos {
		if i >= len(fileInfos)-n {
			break
		}
		err := os.Remove(fileInfo.path)
		if err != nil {
			return deletedFiles, err
		}
		deletedFiles = append(deletedFiles, fileInfo.path)
	}

	return deletedFiles, nil
}
