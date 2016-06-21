package volume

import "os"

type Volume struct {
	Index    *Index
	DataFile *os.File
}

