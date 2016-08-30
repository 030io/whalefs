package check

import (
	"os"
	"github.com/syndtr/goleveldb/leveldb"
)

type Checker struct {
	dataFile *os.File
	index    *leveldb.DB
	status   *leveldb.DB
}

func NewChecker(volumePath string) (*Checker, error) {
	checker := new(Checker)

	var err error
	checker.dataFile, err = os.Open(volumePath)
	if err != nil {
		return checker, err
	}

	checker.index, err = leveldb.Open(volumePath + ".index", nil)
	if err != nil {
		return checker, err
	}

	checker.status, err = leveldb.Open(volumePath + ".status", nil)
	return checker, err
}

func (c *Checker)Check() {
	//TODO: 检查volume index status
}
