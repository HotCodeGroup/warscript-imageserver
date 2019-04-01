package controllers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/HotCodeGroup/warscript-imageserver/storage"
	"github.com/HotCodeGroup/warscript-imageserver/utils"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type createResp struct {
	PhotoUUID string `json:"photo_uuid"`
}

const (
	originSuf    = "origin"
	square300Suf = "300x300"
)

type Controller struct {
	store *storage.Storage
}

func Init(awsAccess, awsSecret, awsToken, bucketName string) (*Controller, error) {
	st, err := storage.Init(awsAccess, awsSecret, awsToken, bucketName)
	if err != nil {
		return nil, errors.Wrap(err, "can not open storage")
	}

	return &Controller{
		store: st,
	}, nil
}

// UploadPhoto handler for photo upload to server
func (c *Controller) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		SendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, multipartFileHeader, err := r.FormFile("photo")
	if err != nil {
		SendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	ident, err := c.store.SaveImage(file, multipartFileHeader.Size)
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

// GetPhoto handler for getting photo from server
func (c *Controller) GetPhoto(w http.ResponseWriter, r *http.Request) {
	photoUUID := mux.Vars(r)["photo_uuid"]
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

	fileBody, fileInfo, err := c.store.GetFile(photoUUID)
	if err != nil {
		SendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+photoUUID)
	w.Header().Set("Content-Type", fileInfo.Type)
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size, 10))

	_, err = io.Copy(w, fileBody)
	if err != nil {
		SendError(w, "can't read file", http.StatusInternalServerError)
		return
	}
}
