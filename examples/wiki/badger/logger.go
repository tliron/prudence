package badger

import (
	"strings"
)

//
// Logger
//

type Logger struct{}

// ([badger.Logger] interface)
func (self Logger) Errorf(format string, args ...any) {
	format = strings.TrimRight(format, "\n")
	log.Errorf(format, args...)
}

// ([badger.Logger] interface)
func (self Logger) Warningf(format string, args ...any) {
	format = strings.TrimRight(format, "\n")
	log.Warningf(format, args...)
}

// ([badger.Logger] interface)
func (self Logger) Infof(format string, args ...any) {
	format = strings.TrimRight(format, "\n")
	log.Infof(format, args...)
}

// ([badger.Logger] interface)
func (self Logger) Debugf(format string, args ...any) {
	format = strings.TrimRight(format, "\n")
	log.Debugf(format, args...)
}
