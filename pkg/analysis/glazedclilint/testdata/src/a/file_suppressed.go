//glazedclilint:file-ignore legacy file scoped suppression in test fixture

package a

import (
	"flag"
	"os"
)

func fileSuppressedEnv() string {
	return os.Getenv("FILE")
}

func fileSuppressedFlag() {
	_ = flag.String("file", "", "file")
}
