package stats

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func captureLogging(action func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	action()
	return buf.String()
}

func TestPanicOn(t *testing.T) {
	if assert.NotPanics(t, func() { PanicOn(nil) }) {
		assert.Panics(t, func() { PanicOn(io.EOF) })
	}
}

func TestLoggers(t *testing.T) {
	testCases := []struct {
		name      string
		function  func(string, ...interface{})
		verbosity int
		want      bool
	}{
		{"Info", Info, 0, false},
		{"Info", Info, 1, true},
		{"Info", Info, 2, true},
		{"Info", Info, 2, true},
		{"Info", Info, 99999, true},

		{"ExtraInfo", ExtraInfo, 0, false},
		{"ExtraInfo", ExtraInfo, 1, false},
		{"ExtraInfo", ExtraInfo, 2, true},
		{"ExtraInfo", ExtraInfo, 2, true},
		{"ExtraInfo", ExtraInfo, 99999, true},
	}
	verbosity := *Verbose
	defer func() { *Verbose = verbosity }()
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s/%d", tc.name, tc.verbosity), func(t *testing.T) {
			*Verbose = tc.verbosity
			output := captureLogging(func() { tc.function("hic%s", "cup") })
			if tc.want {
				assert.Contains(t, output, "hiccup\n")
			} else {
				assert.Equal(t, "", output)
			}
		})
	}
}
