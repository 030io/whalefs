package api

import (
	"net/http"
	"regexp"
	"strconv"
	"gopkg.in/bufio.v1"
)

const readBufferSize = 512 * 1024

var publicUrlRegex *regexp.Regexp

func init() {
	var err error
	publicUrlRegex, err = regexp.Compile("/([0-9]*)/([0-9]*)/(.*)")
	if err != nil {
		panic(err)
	}
}

func (vm *VolumeManager)publicEntry(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		switch r.URL.Path {
		case "/favicon.ico":
			w.WriteHeader(http.StatusMethodNotAllowed)
		default:
			if publicUrlRegex.MatchString(r.URL.Path) {
				vm.publicReadFile(w, r)
			}else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

//TODO: public read file
func (vm *VolumeManager)publicReadFile(w http.ResponseWriter, r *http.Request) {
	match := publicUrlRegex.FindStringSubmatch(r.URL.Path)
	if len(match) != 4 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(r.URL.String() + " can't match"))
		return
	}

	vid, _ := strconv.Atoi(match[1])
	volume := vm.Volumes[vid]
	if volume == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fid, _ := strconv.ParseUint(match[2], 10, 64)
	file, err := volume.Get(fid)
	if err != nil || file.Info.FileName != match[3] {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	reader := bufio.NewReaderSize(file, readBufferSize)
	for {
		data, err := reader.Peek(4096)
		w.Write(data)
		if err != nil {
			break
		}
	}
}
