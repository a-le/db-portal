package handlers

import (
	"net/http"
)

func (s *Services) StaticFileHandler() http.Handler {
	fs := http.FileServer(http.Dir("./web"))
	return http.StripPrefix("/web/", fs)
}
