package volume

type Index interface {
	Get(uint64) (*FileInfo, error)
	Set(*FileInfo) error
	Delete(uint64) error
}
