package handlers

import (
	"db-portal/internal/jsminifier"
	"db-portal/internal/meta"
	"db-portal/internal/security"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/csrf"
)

func (s *Services) IndexHandler(w http.ResponseWriter, r *http.Request) {
	// check if min.js needs update
	jsInfos, err := jsminifier.GetInfos(meta.JsPath, meta.MinjsPath)
	if err != nil {
		fmt.Println("error checking if the JS minified version needs updating", err)
	}
	if jsInfos.Expired {
		if err = jsminifier.Combinify(meta.JsPath, meta.MinjsPath); err != nil {
			fmt.Println("error while minifying JS", err)
		} else {
			fmt.Println(meta.MinjsPath + " has been updated; a new minified version has been created")
		}
	}

	// prepare html (some js is injected)
	var html string
	if data, err := os.ReadFile("./web/index.html"); err != nil {
		fmt.Println("error while reading index.html")
	} else {
		cssInfo, _ := os.Stat("./web/style.css")
		jsCode := `<script>const versionInfo = { js: '%d', css: '%d', server: '%s', appName: '%s' };const username = '%s';</script>`
		js := fmt.Sprintf(jsCode, jsInfos.ModTime().Unix(), cssInfo.ModTime().Unix(), meta.Version, meta.AppName, r.Context().Value(security.UserContextKey).(string))
		html = strings.Replace(string(data), "{{.js}}", js, 1)

		// Add CSRF token
		html = strings.Replace(html, "{{.csrfToken}}", csrf.Token(r), 1)
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
