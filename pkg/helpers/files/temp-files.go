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

// GarbageCollectTemporaryFiles deletes the oldest files in a directory based on a given mask,
// leaving only 'n' newest files. It returns a slice of paths of the deleted files.
// It requires read and write permissions on the directory. If it encounters a file that it cannot delete,
// it will stop and return the files it has deleted so far along with the error.
func GarbageCollectTemporaryFiles(tempDir string, mask string, n int) ([]string, error) {
	var deletedFiles []string

	matchingFiles, err := filepath.Glob(filepath.Join(tempDir, mask))
	if err != nil {
		return nil, err
	}

	d, err := os.Stat(tempDir)
	if err != nil {
		return nil, err
	}
	if d.Mode().Perm()&0444 == 0 || d.Mode().Perm()&0222 == 0 {
		return nil, &os.PathError{Op: "GarbageCollectTemporaryFiles", Path: tempDir, Err: os.ErrPermission}
	}

	var fileInfos []fileInfo
	for _, fileName := range matchingFiles {
		fi, err := os.Stat(fileName)
		if err != nil {
			continue
		}

		if !fi.IsDir() {
			fileInfos = append(fileInfos, fileInfo{FileInfo: fi, path: fileName})
		}
	}

	sort.Sort(byModTime(fileInfos))

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
