package web

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, controller *WebController) {
	mux.HandleFunc("GET /", controller.GetMainPage)
	mux.HandleFunc("GET /assets/", controller.ServeAssets)
}
