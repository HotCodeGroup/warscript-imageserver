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
func ResizeImage(imgName, resizeName, itype string, height, width uint) error {
	imgJpg, err := openResize(imgName, itype)
	if err != nil {
		return err
	}
	imgJpg = resize.Resize(height, width, imgJpg, resize.Bicubic)
	err = saveResize(resizeName, itype, imgJpg)
	if err != nil {
		return err
	}

	return nil
}

func openResize(imgName, itype string) (image.Image, error) {
	imgIn, err := os.Open(imgName)

	if err != nil {
		return nil, errors.Wrap(ErrInternal, "can't open origin")
	}
	defer imgIn.Close()

	var imgJpg image.Image

	switch itype {
	case "gif":
		imgJpg, err = gif.Decode(imgIn)
	case "png":
		imgJpg, err = png.Decode(imgIn)
	case "jpg", "jpeg":
		imgJpg, err = jpeg.Decode(imgIn)
	default:
		return nil, errors.Wrap(ErrBadType, "can't resize type "+itype)
	}

	if err != nil {
		return nil, errors.Wrap(ErrInternal, "can't decode origin")
	}
	return imgJpg, nil
}

func saveResize(resizeName, itype string, imgJpg image.Image) error {
	imgOut, err := os.Create(resizeName)
	if err != nil {
		return errors.Wrap(ErrInternal, "can't create image for resize")
	}
	defer imgOut.Close()

	switch itype {
	case "gif":
		err = gif.Encode(imgOut, imgJpg, nil)
	case "png":
		err = png.Encode(imgOut, imgJpg)
	case "jpg", "jpeg":
		err = jpeg.Encode(imgOut, imgJpg, nil)
	default:
		return errors.Wrap(ErrBadType, "can't resize type "+itype)
	}

	if err != nil {
		return errors.Wrap(ErrInternal, "failed to resize")
	}
	return nil
}
