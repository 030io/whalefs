package manager

import (
	"mime"
	"path"
)

func init() {
	if mime.TypeByExtension(".txt") == "" {
		panic("mime.types not found")
	}
}

func get_content_type(filepath string) string {
	content_type := "application/octet-stream"
	ext := path.Ext(filepath)
	if ext != "" && ext != "." {
		content_type_ := mime.TypeByExtension(ext)
		if content_type_ != "" {
			content_type = content_type_
		}
	}
	return content_type
}
