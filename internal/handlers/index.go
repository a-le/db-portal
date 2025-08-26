package handlers

import (
	"bytes"
	"db-portal/internal/jsminifier"
	"db-portal/internal/meta"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
)

type pageData struct {
	AppName    template.JS
	AppVersion template.JS
	JsScripts  template.JS
	JsVersion  template.JS
	CssScript  template.JS
	CssVersion template.JS
}

func (s *Services) IndexHandler(w http.ResponseWriter, r *http.Request) {

	// minify js files when necessary
	jsMini, err := jsminifier.NewJSMinifyStatus(meta.ManifestPath, meta.MinjsPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("error checking if the JS minified version needs updating: %v", err), http.StatusBadRequest)
		return
	}
	if jsMini.Expired {
		jsMini, err = jsMini.Combinify()
		if err != nil {
			http.Error(w, fmt.Sprintf("error while minifying JS: %v", err), http.StatusBadRequest)
			return
		}

		// replace import.js
		content := "// this file is generated automatically for IDE awareness. \n"
		for _, f := range jsMini.SourceFiles {
			path := "./" + f[len(meta.WebFolder):]
			content += fmt.Sprintf("import \"%s\";\n", path)
		}
		os.WriteFile(meta.ImportJSPath, []byte(content), 0644)
		fmt.Printf("JS changes detected. The contents of %s and %s have be updated.\n", meta.MinjsPath, meta.ImportJSPath)
	}

	// prepare template vars
	pageData := pageData{
		AppName:    template.JS(meta.AppName),
		AppVersion: template.JS(meta.AppVersion),
		JsVersion:  template.JS(fmt.Sprintf("%d", jsMini.LatestSourceModTime.Unix())),
		CssScript:  template.JS(meta.CssPath),
	}
	info, err := os.Stat(meta.CssPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("error, can't find css file %v: %v", meta.CssPath, err), http.StatusBadRequest)
		return
	}
	pageData.CssVersion = template.JS(fmt.Sprintf("%d", info.ModTime().Unix()))

	if r.URL.Query().Has("nominify") {
		quoted := make([]string, len(jsMini.SourceFiles))
		for i, file := range jsMini.SourceFiles {
			quoted[i] = `"` + file + `"`
		}
		pageData.JsScripts = template.JS(strings.Join(quoted, ",")) // Format as JS array: "file1.js","file2.js",...
	} else {
		pageData.JsScripts = `"` + template.JS(jsMini.MinifiedPath[1:]) + `"` // remove 1st char which is a dot
	}

	// parse and execute the template
	tmpl, err := template.ParseFiles(meta.IndexPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing template file: %v", err), http.StatusBadRequest)
		return
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, pageData)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to render HTML response: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(buf.Bytes())
}
