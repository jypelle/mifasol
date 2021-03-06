package webSrv

import (
	"github.com/jypelle/mifasol/internal/srv/webSrv/model"
	"net/http"
)

func (s *WebServer) GetMainLayout(r *http.Request) *model.MainLayout {

	layout := &model.MainLayout{
		Title:     "Mifasol",
		MenuTitle: "Mifasol",
	}
	return layout
}
