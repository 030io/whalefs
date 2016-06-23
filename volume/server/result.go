package server

type PostFileResult struct {
	Vid      int `json:"vid"`
	Fid      uint64 `json:"fid"`
	FileName string `json:"filename"`
}