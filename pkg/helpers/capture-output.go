package helpers

import (
	"bytes"
	"io"
	"os"
)

// CaptureOutput executes a provided function and captures any data
// the function writes to standard output. It returns the captured output
// as a string and any error that occurred during execution or capturing.
// If multiple errors occur, only the last one will be returned.
func CaptureOutput(fn func() error) (string, error) {
	var buf bytes.Buffer
	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}

	os.Stdout = w

	// run in a goroutine so that we don't deadlock
	captureDone := make(chan struct{})
	go func() {
		_, errCopy := io.Copy(&buf, r)
		if errCopy != nil {
			err = errCopy
		}
		close(captureDone)
	}()

	errRun := fn()
	if errRun != nil {
		err = errRun
	}

	if errClose := w.Close(); errClose != nil {
		err = errClose
	}

	<-captureDone
	os.Stdout = originalStdout
	return buf.String(), err
}
