package volume

import (
	"testing"
	"io/ioutil"
	"os"
)

func TestIndexLevelDB(t *testing.T) {
	dir, _ := ioutil.TempDir("", "test_index_leveldb_")
	defer os.RemoveAll(dir)

	var (
		index Index
		err error
	)

	index, err = NewLevelDBIndex(dir, 0)
	if err != nil {
		t.Error(err)
	}

	fi, err := index.Get(0)
	if fi != nil || err == nil {
		t.Error(fi, err)
	}

	fi = new(FileInfo)
	fi.Fid = 0
	fi.FileName = "test_file"
	err = index.Set(fi)
	if err != nil {
		t.Error(err)
	}

	fi, err = index.Get(0)
	if fi == nil || err != nil || fi.FileName != "test_file" {
		t.Error(fi, err)
	}

	index.Delete(0)
	fi, err = index.Get(0)
	if fi != nil || err == nil {
		t.Error(fi, err)
	}
}
