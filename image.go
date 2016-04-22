package wordepress

import (
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"path"
)

type Image struct {
	Filename  string
	Extension string
	MimeType  string
	Hash      string
	Content   []byte
}

func ReadImage(filename string) (*Image, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	sum := sha1.Sum(content)

	return &Image{
		Filename:  filename,
		Extension: path.Ext(filename),
		MimeType:  http.DetectContentType(content),
		Hash:      hex.EncodeToString(sum[:]),
		Content:   content}, nil
}
