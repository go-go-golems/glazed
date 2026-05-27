package a

import "os"

//glazedclilint:file-ignore-old legacy typo in test fixture

func malformedFileSuppressionDoesNotSuppress() string {
	return os.Getenv("MALFORMED_FILE") // want `use Glazed config/env middleware`
}
