//glazedclilint:file-ignore // want `glazedclilint suppression requires a reason`

package a

import "os"

func invalidFileSuppressionStillReports() string {
	return os.Getenv("FILE_INVALID") // want `use Glazed config/env middleware`
}
