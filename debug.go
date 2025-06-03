package usbtmc

import (
	"io"
	"log"
	"os"
)

const (
	debugEnv    = "USBTMC_DEBUG"
	debugPrefix = "[usbtmc] "
)

var (
	debug *log.Logger
)

func init() {
	if os.Getenv(debugEnv) != "" {
		debug = log.New(os.Stderr, debugPrefix, log.LstdFlags)
	} else {
		debug = log.New(io.Discard, "", 0)
	}
}
