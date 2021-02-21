// logging/output management functions.
package stats

import (
	"log"

	"github.com/spf13/pflag"
)

// PanicOn will panic if err is not nil.
func PanicOn(err error) {
	if err != nil {
		panic(err)
	}
}

// ExtraInfo will log a message at verbosity level 2 or higher.
func ExtraInfo(msg string, args ...interface{}) {
	if *Verbose > 1 {
		log.Printf(msg+"\n", args...)
	}
}

// Info will log a message at verbosity level 1 or higher.
func Info(msg string, args ...interface{}) {
	if *Verbose > 0 {
		log.Printf(msg+"\n", args...)
	}
}

// Verbose enables additional output.
var Verbose = pflag.CountP("verbose", "v", "enable additional output")
