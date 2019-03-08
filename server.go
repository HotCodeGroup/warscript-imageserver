package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	uuid "github.com/satori/go.uuid"
)

// imgError codes
const (
	internal = iota
	badType
)

type imgError struct {
	code int
	msg  string
}

type createResp struct {
	PhotoUUID string `json:"photo_uuid"`
}

func detectImageType(file io.ReadSeeker) (string, *imgError) {
	buff := make([]byte, 512) // http://golang.org/pkg/net/http/#DetectContentType
	_, err := file.Read(buff)
	if err != nil {
		return "", &imgError{
			code: internal,
			msg:  "failed to read",
		}
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return "", &imgError{
			code: internal,
			msg:  "failed to seek",
		}
	}
	filetype := http.DetectContentType(buff)

	fmt.Println(filetype)

	switch filetype {
	case "image/jpeg", "image/jpg", "image/gif", "image/png":
		return strings.TrimPrefix(filetype, "image/"), nil
	default:
		return "", &imgError{
			code: badType,
			msg:  filetype + " is not allowed",
		}
	}
}

func saveImage(file io.ReadSeeker) (string, *imgError) {
	itype, imgErr := detectImageType(file)
	if imgErr != nil {
		return "", imgErr
	}
	dirpath := "images"
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		err = os.MkdirAll(dirpath, 0777)
		if err != nil {
			return "", &imgError{
				code: internal,
				msg:  "failed to create directoty",
			}
		}
	}
	ident, err := uuid.NewV4()
	if err != nil {
		return "", &imgError{
			code: internal,
			msg:  "failed to create uuid",
		}
	}
	filesetid := ident.String()
	originName := "./" + dirpath + "/" + filesetid + "." + "origin" + "." + itype
	avasizeName := "./" + dirpath + "/" + filesetid + "." + "300x300" + "." + itype
	f, err := os.Create(originName)
	if err != nil {
		return "", &imgError{
			code: internal,
			msg:  "can't create origin",
		}
	}

	defer f.Close()
	_, err = io.Copy(f, file)
	if err != nil {
		return "", &imgError{
			code: internal,
			msg:  "failed to resize copying",
		}
	}

	imgErr = resizeImage(originName, avasizeName, itype, 300, 300)
	if imgErr != nil {
		return "", imgErr
	}
	return filesetid, nil
}

func upload(w http.ResponseWriter, r *http.Request) {
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
	ident, imgErr := saveImage(file)
	if imgErr != nil {
		switch imgErr.code {
		case internal:
			SendError(w, imgErr.msg, http.StatusInternalServerError)
		case badType:
			SendError(w, imgErr.msg, http.StatusNotAcceptable)
		}
		return
	}

	SendResponse(w, createResp{
		PhotoUUID: ident,
	}, http.StatusOK)

}

func sendFile(w http.ResponseWriter, filename string) {
	Openfile, err := os.Open(filename)
	if err != nil {
		SendError(w, "file not opend: " + err.Error(), http.StatusNotFound)
		return
	}
	defer Openfile.Close()
	FileHeader := make([]byte, 512) // http://golang.org/pkg/net/http/#DetectContentType
	_, err = Openfile.Read(FileHeader)
	if err != nil {
		SendError(w, "can't read file: " + err.Error(), http.StatusInternalServerError)
		return
	}
	FileContentType := http.DetectContentType(FileHeader)
	FileStat, err := Openfile.Stat()
	if err != nil {
		SendError(w, "can't read file", http.StatusInternalServerError)
		return
	}
	FileSize := strconv.FormatInt(FileStat.Size(), 10)

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", FileContentType)
	w.Header().Set("Content-Length", FileSize)

	_, err = Openfile.Seek(0, 0)
	if err != nil {
		SendError(w, "can't read file", http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(w, Openfile)
	if err != nil {
		SendError(w, "can't read file", http.StatusInternalServerError)
		return
	}
}

func sendPhoto(w http.ResponseWriter, r *http.Request) {
	PhotoUUID := mux.Vars(r)["photo_uuid"]
	var format string
	keys, ok := r.URL.Query()["format"]

	if !ok || len(keys[0]) < 1 {
		format = "origin"
	} else {
		format = keys[0]
	}
	fmt.Println("format: ", format)

	if format != "origin" && format != "300x300" {
		SendError(w, "bad format "+format, http.StatusBadRequest)
		return
	}
	matches, err := filepath.Glob("./images/" + PhotoUUID + "." + format + ".*")
	if err != nil {
		SendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(matches) != 1 {
		SendError(w, "bad uuid", http.StatusNotFound)
		return
	}

	sendFile(w, matches[0])
}

var (
	log = logging.MustGetLogger("auth")

	logFormat = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
)

var uploadFormTmpl = []byte(`
<html>
	<body>
	<form action="/photos" method="post" enctype="multipart/form-data">
		Image: <input type="file" name="photo">
		<input type="submit" value="Upload">
	</form>
	</body>
</html>
`)

func mainPage(w http.ResponseWriter, r *http.Request) {
	w.Write(uploadFormTmpl)
}

func main() {
	// setting logs format
	backendLog := logging.NewLogBackend(os.Stderr, "", 0)
	logging.SetBackend(logging.NewBackendFormatter(backendLog, logFormat))
	fmt.Println(time.Now().Format("January"))

	r := mux.NewRouter()
	r.HandleFunc("/photos", upload).Methods("POST")
	r.HandleFunc("/photos/{photo_uuid}", sendPhoto).Methods("GET")
	r.HandleFunc("/", mainPage).Methods("GET")
	port := os.Getenv("PORT")
	log.Noticef("MainService successfully started at port %s", port)
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Criticalf("cant start main server. err: %s", err.Error())
		return
	}
}
