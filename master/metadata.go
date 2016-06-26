package master

type Metadata interface {
	Get(filePath string) (vid int, fid uint64, fileName string, err error)
	Set(filePath string, vid int, fid uint64, fileName string) error
	Delete(filePath string) error
	Has(filePath string) bool
	setConfig(key string, value string) error
	getConfig(key string) (value string, err error)
}
