package volume

type Index interface {
	Has(fid uint64) bool
	Get(fid uint64) (*FileInfo, error)
	Set(fi *FileInfo) error
	Delete(fid uint64) error
}
