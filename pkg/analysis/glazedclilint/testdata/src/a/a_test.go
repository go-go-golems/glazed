package a

import "os"

func testEnvIsSkipped() string {
	return os.Getenv("TEST")
}
