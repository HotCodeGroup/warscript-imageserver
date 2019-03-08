package main

import (
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"

	"github.com/nfnt/resize"
)

func resizeImage(imgName, resizeName, itype string, height, width uint) *imgError {
	imgJpg, imgErr := openResize(imgName, itype)
	if imgErr != nil {
		return imgErr
	}
	imgJpg = resize.Resize(height, width, imgJpg, resize.Bicubic)
	imgErr = saveResize(resizeName, itype, imgJpg)
	if imgErr != nil {
		return imgErr
	}

	return nil
}

func openResize(imgName, itype string) (image.Image, *imgError) {
	imgIn, err := os.Open(imgName)
	if err != nil {
		return nil, &imgError{
			code: internal,
			msg:  "can't open origin",
		}
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
		return nil, &imgError{
			code: badType,
			msg:  "can't resize type " + itype,
		}
	}

	if err != nil {
		return nil, &imgError{
			code: internal,
			msg:  "can't decode origin",
		}
	}
	return imgJpg, nil
}

func saveResize(resizeName, itype string, imgJpg image.Image) *imgError {
	imgOut, err := os.Create(resizeName)
	if err != nil {
		return &imgError{
			code: internal,
			msg:  "can't create avasize",
		}
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
		return &imgError{
			code: badType,
			msg:  "can't resize type " + itype,
		}
	}

	if err != nil {
		return &imgError{
			code: internal,
			msg:  "failed to resize",
		}
	}
	return nil
}
