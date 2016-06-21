package volume

import (
	"testing"
	"io"
)

func TestFile(t *testing.T) {
	var f io.ReadWriteSeeker = new(File)
	if f == nil {
		t.Error(f)
	}
}
