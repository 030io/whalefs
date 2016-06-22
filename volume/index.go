package volume

type Index interface {
	Has(uint64) bool
	Get(uint64) (*FileInfo, error)
	Set(*FileInfo) error
	Delete(uint64) error
}
