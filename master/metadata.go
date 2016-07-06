package master

import "io"

type Metadata interface {
	Get(filePath string) (vid uint64, fid uint64, fileName string, err error)
	Set(filePath string, vid uint64, fid uint64, fileName string) error
	Delete(filePath string) error
	Has(filePath string) bool
	io.Closer
}
