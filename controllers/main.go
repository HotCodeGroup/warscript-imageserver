package controllers

import "net/http"

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
	w.Write(uploadFormTmpl)
}
