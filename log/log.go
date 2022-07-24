package log

import (
	"github.com/eycorsican/go-tun2socks/common/log"
	"strings"
)

func SetLogLevel(logLevel string) {
	// Set log level.
	switch strings.ToLower(logLevel) {
	case "debug":
		log.SetLevel(log.DEBUG)
	case "info":
		log.SetLevel(log.INFO)
	case "warn":
		log.SetLevel(log.WARN)
	case "error":
		log.SetLevel(log.ERROR)
	case "none":
		log.SetLevel(log.NONE)
	default:
		panic("unsupported logging level")
	}
}
