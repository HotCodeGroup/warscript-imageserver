package utils

import (
	"io"
	"net/http"

	"github.com/pkg/errors"
)

var (
	// ErrInternal внутренняя ошибка
	ErrInternal = errors.New("internal")
	// ErrBadType некорректный тип
	ErrBadType = errors.New("bad_type")
)

var validImageTypes = map[string]interface{}{
	"image/jpeg": struct{}{},
	"image/jpg":  struct{}{},
	"image/gif":  struct{}{},
	"image/png":  struct{}{},
}

func GetImageType(file io.ReadSeeker) (string, error) {
	buff := make([]byte, 512) // http://golang.org/pkg/net/http/#DetectContentType
	_, err := file.Read(buff)
	if err != nil {
		return "", errors.Wrap(ErrInternal, "failed to read")
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return "", errors.Wrap(ErrInternal, "failed to seek")
	}

	filetype := http.DetectContentType(buff)
	if _, ok := validImageTypes[filetype]; !ok {
		return "", errors.Wrapf(ErrBadType, "%s is not allowed", filetype)
	}

	return filetype, nil
}
