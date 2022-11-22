package main

import (
	_ "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	_ "github.com/pkg/errors"
	"github.com/rs/zerolog"
	_ "github.com/rs/zerolog/pkgerrors"
	"glazed/cmd/dd-cli/cmds"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	_ = cmds.RootCmd.Execute()
}
