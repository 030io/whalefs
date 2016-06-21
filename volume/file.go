package volume

import (
	"time"
	"encoding/binary"
	"errors"
)

type FileInfo struct {
	Fid      uint64
	Offset   uint64
	Size     uint64
	Ctime    time.Time
	Mtime    time.Time
	Atime    time.Time
	FileName string
}

func (iv *FileInfo)MarshalBinary() []byte {
	data := make([]byte, 48 + len(iv.FileName))
	binary.LittleEndian.PutUint64(data[0:8], iv.Fid)
	binary.LittleEndian.PutUint64(data[8:16], iv.Offset)
	binary.LittleEndian.PutUint64(data[16:24], iv.Size)
	binary.LittleEndian.PutUint64(data[24:32], uint64(iv.Ctime.Unix()))
	binary.LittleEndian.PutUint64(data[32:40], uint64(iv.Mtime.Unix()))
	binary.LittleEndian.PutUint64(data[40:48], uint64(iv.Atime.Unix()))
	copy(data[48:], []byte(iv.FileName))
	return data
}

func (iv *FileInfo)UnMarshalBinary(data []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	iv.Fid = binary.LittleEndian.Uint64(data[0:8])
	iv.Offset = binary.LittleEndian.Uint64(data[8:16])
	iv.Size = binary.LittleEndian.Uint64(data[16:24])
	iv.Ctime = time.Unix(int64(binary.LittleEndian.Uint64(data[24:32])), 0)
	iv.Mtime = time.Unix(int64(binary.LittleEndian.Uint64(data[32:40])), 0)
	iv.Atime = time.Unix(int64(binary.LittleEndian.Uint64(data[40:48])), 0)
	iv.FileName = string(data[48:])
	return err
}

type File struct {
	volume *Volume
	Info   *FileInfo
	offset uint64
}

func (f *File)Read(b []byte) (n int, err error) {
	start := f.Info.Offset + f.offset
	end := f.Info.Offset + f.Info.Size
	length := end - start
	if len(b) > int(length) {
		b = b[:length]
	}
	n, err = f.volume.dataFile.ReadAt(b, int64(start))
	f.offset += uint64(n)
	return
}

func (f *File)Write(b []byte) (n int, err error) {
	start := f.Info.Offset + f.offset
	end := f.Info.Offset + f.Info.Size
	length := end - start
	if len(b) > int(length) {
		//b = b[:length]
		return 0, errors.New("you should create a new File to write")
	}else {
		n, err = f.volume.dataFile.WriteAt(b, int64(start))
		return
	}
}

func (f *File)Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0:
		f.offset = uint64(offset)
	case 1:
		f.offset += uint64(offset)
	case 2:
		f.offset = f.Info.Size - uint64(offset)
	}
	if f.offset > f.Info.Size {
		return int64(f.offset), errors.New("offset > file.size")
	}else {
		return int64(f.offset), nil
	}
}
