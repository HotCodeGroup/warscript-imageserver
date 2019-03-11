package utils

import (
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"

	"github.com/pkg/errors"

	"github.com/nfnt/resize"
)

var (
	// ErrInternal внутренняя ошибка
	ErrInternal = errors.New("internal")
	// ErrBadType некорректный тип
	ErrBadType = errors.New("bad_type")
)

// ResizeImage resizes image by given size
func ResizeImage(imgName, resizeName string, height, width uint) error {
	img, itype, err := openResize(imgName)
	if err != nil {
		return err
	}

	img = resize.Resize(height, width, img, resize.Bicubic)
	err = saveResize(resizeName, itype, img)
	if err != nil {
		return err
	}

	return nil
}

func openResize(imgName string) (image.Image, string, error) {
	imgIn, err := os.Open(imgName)
	if err != nil {
		return nil, "", errors.Wrap(ErrInternal, "can't open origin")
	}
	defer imgIn.Close()

	img, itype, err := image.Decode(imgIn)
	if err != nil {
		return nil, "", errors.Wrap(ErrInternal, "can't decode origin")
	}

	return img, itype, nil
}

func saveResize(resizeName, itype string, img image.Image) error {
	imgOut, err := os.Create(resizeName)
	if err != nil {
		return errors.Wrap(ErrInternal, "can't create image for resize")
	}
	defer imgOut.Close()

	switch itype {
	case "gif":
		err = gif.Encode(imgOut, img, nil)
	case "png":
		err = png.Encode(imgOut, img)
	case "jpg", "jpeg":
		err = jpeg.Encode(imgOut, img, nil)
	default:
		return errors.Wrap(ErrBadType, "can't resize type "+itype)
	}

	if err != nil {
		return errors.Wrap(ErrInternal, "failed to resize")
	}
	return nil
}
