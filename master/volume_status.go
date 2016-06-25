package master

type VolumeStatus struct {
	Id           int
	DataFileSize uint64
	//FreeSpace    uint64

	//Writable     bool

	vmStatus     *VolumeManagerStatus `json:"-"`
}
