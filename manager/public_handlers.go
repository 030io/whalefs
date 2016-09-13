package manager

import (
	"net/http"
	"regexp"
	"strconv"
	"io"
	"strings"
	"fmt"
)

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
		if publicUrlRegex.MatchString(r.URL.Path) {
			vm.publicReadFile(w, r)
		}else {
			http.NotFound(w, r)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (vm *VolumeManager)publicReadFile(w http.ResponseWriter, r *http.Request) {
	match := publicUrlRegex.FindStringSubmatch(r.URL.Path)

	vid, _ := strconv.ParseUint(match[1], 10, 64)
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

	if r.Header.Get("Range") != "" {
		ranges := strings.Split(r.Header.Get("Range")[6:], "-")
		start, err := strconv.ParseUint(ranges[0], 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var length uint64
		if start > file.Info.Size {
			start = file.Info.Size
			length = 0
		}else if ranges[1] != "" {
			end, err := strconv.ParseUint(ranges[1], 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if end > file.Info.Size - 1 {
				end = file.Info.Size - 1
			}

			length = end - start + 1
		}else {
			length = file.Info.Size - start
		}

		w.Header().Set("Content-Length", strconv.FormatUint(length, 10))
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, start + length - 1, file.Info.Size))
		w.WriteHeader(http.StatusPartialContent)

		if length != 0 {
			file.Seek(int64(start), 0)
			io.CopyN(w, file, int64(length))
		}
	}else {
		io.Copy(w, file)
	}
}
