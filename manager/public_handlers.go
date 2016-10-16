package manager

import (
	"net/http"
	"regexp"
	"strconv"
	"io"
	"strings"
	"fmt"
	"time"
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
		} else {
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
	w.Header().Set("ETag", fmt.Sprintf("\"%d\"", fid))
	//暂时不使用Last-Modified,用ETag即可
	//w.Header().Set("Last-Modified", file.Info.Mtime.Format(http.TimeFormat))
	w.Header().Set("Expires", time.Now().Add(DefaultExpires).Format(http.TimeFormat))

	etagMatch := false
	if r.Header.Get("If-None-Match") != "" {
		s := r.Header.Get("If-None-Match")
		if etag, err := strconv.ParseUint(s[1:len(s) - 1], 10, 64); err == nil && etag == fid {
			etagMatch = true
		}
	}

	if r.Header.Get("Range") != "" {
		ranges := strings.Split(r.Header.Get("Range")[6:], "-")
		start, err := strconv.ParseUint(ranges[0], 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		length := uint64(0)
		if start > file.Info.Size {
			start = file.Info.Size
		} else if ranges[1] != "" {
			end, err := strconv.ParseUint(ranges[1], 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			end += 1

			if end > file.Info.Size {
				end = file.Info.Size
			}

			length = end - start
		} else {
			length = file.Info.Size - start
		}

		w.Header().Set("Content-Length", strconv.FormatUint(length, 10))

		if length == 0 {
			w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
			return
		}

		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, start + length - 1, file.Info.Size))

		if etagMatch {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.WriteHeader(http.StatusPartialContent)

		if r.Method != http.MethodHead {
			file.Seek(int64(start), 0)
			io.CopyN(w, file, int64(length))
		}
	} else {
		w.Header().Set("Content-Length", strconv.FormatUint(file.Info.Size, 10))
		if etagMatch {
			w.WriteHeader(http.StatusNotModified)
		} else if r.Method != http.MethodHead {
			io.Copy(w, file)
		}
	}
}
