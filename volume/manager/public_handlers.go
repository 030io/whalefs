package manager

import (
	"net/http"
	"regexp"
	"strconv"
)

const readBufferSize = 64 * 1024

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
			http.NotFound(w, r)
		default:
			if publicUrlRegex.MatchString(r.URL.Path) {
				vm.publicReadFile(w, r)
			}else {
				http.NotFound(w, r)
			}
		}
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
	}
}

func (vm *VolumeManager)publicReadFile(w http.ResponseWriter, r *http.Request) {
	match := publicUrlRegex.FindStringSubmatch(r.URL.Path)

	vid, _ := strconv.Atoi(match[1])
	volume := vm.Volumes[vid]
	if volume == nil {
		http.Error(w, "can't find volume", http.StatusNotFound)
		return
	}

	fid, _ := strconv.ParseUint(match[2], 10, 64)
	file, err := volume.Get(fid)
	if err != nil || file.Info.FileName != match[3] {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Accept-Ranges", "bytes")
	if r.Method == http.MethodHead {
		w.Header().Set("Content-Length", strconv.FormatUint(file.Info.Size, 10))
		return
	}
	data := make([]byte, readBufferSize)
	for {
		n, err := file.Read(data)
		w.Write(data[:n])
		if err != nil {
			break
		}
	}
}
