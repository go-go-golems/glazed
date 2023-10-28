package helpers

import (
	"bytes"
	"io"
	"os"
)

// CaptureOutput executes a provided function and captures any data
// the function writes to standard output.
func CaptureOutput(fn func() error) (string, error) {
	var buf bytes.Buffer

	// Save the original stdout to restore it later
	originalStdout := os.Stdout

	// Use a temporary pipe to capture output
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}

	// This is where you'll capture the output
	os.Stdout = w

	// Capture stdout in a separate goroutine to prevent deadlocks
	captureDone := make(chan struct{})
	go func() {
		// Copy the data written to the PipeReader to the buffer
		_, errCopy := io.Copy(&buf, r)
		if errCopy != nil {
			// This error will be handled later
			err = errCopy
		}

		// Signal that capturing is complete
		close(captureDone)
	}()

	// Run the function that writes to stdout
	errRun := fn()
	if errRun != nil {
		err = errRun // Store this error to return later after cleaning up
	}

	// You must close the writer or the copy in the goroutine will never complete
	if errClose := w.Close(); errClose != nil {
		err = errClose // Store this error to return later after cleaning up
	}

	// Wait for capturing to finish
	<-captureDone

	// Restore original stdout
	os.Stdout = originalStdout

	// Return the captured output and any error that occurred
	return buf.String(), err
}
