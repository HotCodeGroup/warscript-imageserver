package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/HotCodeGroup/warscript-imageserver/utils"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type createResp struct {
	PhotoUUID string `json:"photo_uuid"`
}

const (
	originSuf    = "origin"
	square300Suf = "300x300"
	dirpath      = "images"
)

// StorageInit inits file storage
func StorageInit() error {
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		err = os.MkdirAll(dirpath, 0700)
		if err != nil {
			return errors.Wrap(err, "failed to create directoty")
		}
	}
	return nil
}

const stdAvaSize = 300

func checkImageType(file io.ReadSeeker) error {
	buff := make([]byte, 512) // http://golang.org/pkg/net/http/#DetectContentType
	_, err := file.Read(buff)
	if err != nil {
		return errors.Wrap(utils.ErrInternal, "failed to read")
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return errors.Wrap(utils.ErrInternal, "failed to seek")
	}

	filetype := http.DetectContentType(buff)
	switch filetype {
	case "image/jpeg", "image/jpg", "image/gif", "image/png":
		return nil // cuts "image/"
	default:
		return errors.Wrapf(utils.ErrBadType, "%s is not allowed", filetype)
	}
}

func saveImage(file io.ReadSeeker) (string, error) {
	err := checkImageType(file)
	if err != nil {
		return "", errors.Wrap(err, "detecting fyle type failed")
	}

	ident, err := uuid.NewV4()
	if err != nil {
		return "", errors.Wrap(utils.ErrInternal, "failed to create uuid")
	}

	filesetid := ident.String()
	originName := fmt.Sprintf("./%s/%s.%s", dirpath, filesetid, originSuf)     // "./" + dirpath + "/" + filesetid + "." + originSuf + "." + itype
	avasizeName := fmt.Sprintf("./%s/%s.%s", dirpath, filesetid, square300Suf) // "./" + dirpath + "/" + filesetid + "." + square300Suf + "." + itype
	f, err := os.Create(originName)
	if err != nil {
		return "", errors.Wrap(utils.ErrInternal, "can't create origin")
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		return "", errors.Wrap(utils.ErrInternal, "failed to copy original")
	}

	err = utils.ResizeImage(originName, avasizeName, stdAvaSize, stdAvaSize)
	if err != nil {
		respErr := errors.Wrap(err, "resizig failed")

		errg := os.RemoveAll(fmt.Sprintf("./%s/%s", dirpath, filesetid))
		if errg != nil {
			return "", errors.Wrap(respErr, fmt.Sprintf("removing copies failed (%s)", errg.Error()))
		}

		return "", respErr
	}
	return filesetid, nil
}

// UploadPhoto handler for photo upload to server
func UploadPhoto(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		SendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("photo")
	if err != nil {
		SendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	ident, err := saveImage(file)
	if err != nil {
		switch errors.Cause(err) {
		case utils.ErrInternal:
			SendError(w, err.Error(), http.StatusInternalServerError)
		case utils.ErrBadType:
			SendError(w, err.Error(), http.StatusNotAcceptable)
		}
		return
	}

	SendResponse(w, createResp{
		PhotoUUID: ident,
	}, http.StatusOK)

}

func getFile(w http.ResponseWriter, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		SendError(w, "file not opend: "+err.Error(), http.StatusNotFound)
		return
	}
	defer file.Close()
	FileHeader := make([]byte, 512) // http://golang.org/pkg/net/http/#DetectContentType
	_, err = file.Read(FileHeader)
	if err != nil {
		SendError(w, "can't read file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	FileContentType := http.DetectContentType(FileHeader)
	FileStat, err := file.Stat()
	if err != nil {
		SendError(w, "can't read file", http.StatusInternalServerError)
		return
	}
	FileSize := strconv.FormatInt(FileStat.Size(), 10)

	_, err = file.Seek(0, 0)
	if err != nil {
		log.Printf("failed to send %s", filename)
		SendError(w, "can't read file", http.StatusInternalServerError)
		return
	}

	log.Printf("sending %s", filename)
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", FileContentType)
	w.Header().Set("Content-Length", FileSize)

	_, err = io.Copy(w, file)
	if err != nil {
		log.Printf("failed to send %s", filename)
		SendError(w, "can't read file", http.StatusInternalServerError)
		return
	}
}

// GetPhoto handler for getting photo from server
func GetPhoto(w http.ResponseWriter, r *http.Request) {
	PhotoUUID := mux.Vars(r)["photo_uuid"]
	var format string
	keys, ok := r.URL.Query()["format"]

	if !ok || len(keys[0]) < 1 {
		format = originSuf
	} else {
		format = keys[0]
	}

	if format != originSuf && format != square300Suf {
		SendError(w, "bad format "+format, http.StatusBadRequest)
		return
	}

	filepath := fmt.Sprintf("./%s/%s.%s", dirpath, PhotoUUID, format)
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		SendError(w, "bad uuid", http.StatusNotFound)
		return
	}

	getFile(w, filepath)
}
