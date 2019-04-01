package controllers

import (
	"net/http"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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

func MainPage(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write(uploadFormTmpl); err != nil {
		log.Error(errors.Wrap(err, "unable to send upload form"))
	}
}
