package home

import (
	"github.com/jypelle/mifasol/internal/srv/webSrv/controller"
	"github.com/jypelle/mifasol/internal/srv/webSrv/model"
	"net/http"
)

type Controller struct {
	*controller.CommonController
}

func NewController(
	webServerer controller.WebServerer,
) *Controller {
	return &Controller{CommonController: controller.NewCommonController(webServerer, "home")}
}

type IndexView struct {
	*model.MainLayout
}

func (c *Controller) IndexAction(w http.ResponseWriter, r *http.Request) {

	view := &IndexView{
		MainLayout: c.GetMainLayout(r),
	}
	view.MenuTitle = "Welcome to Mifasol"

	c.HtmlWriterRender(w, view, "layout/main.html", "controller/home/index.html")

}
