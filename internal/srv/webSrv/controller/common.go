package controller

import (
	"github.com/jypelle/mifasol/internal/srv/config"
	"github.com/jypelle/mifasol/internal/srv/webSrv/model"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)

type CommonController struct {
	WebServerer
	controllerId string
	log          *logrus.Entry
}

func NewCommonController(
	webServerer WebServerer,
	controllerId string,
) *CommonController {
	return &CommonController{WebServerer: webServerer, controllerId: controllerId, log: webServerer.Log().WithField("controllerId", controllerId)}
}

func (c *CommonController) Log() *logrus.Entry {
	return c.log
}

func (c *CommonController) ControllerId() string {
	return c.controllerId
}

func (c *CommonController) GetMainLayout(r *http.Request) *model.MainLayout {
	mainLayout := c.WebServerer.GetMainLayout(r)
	mainLayout.ControllerId = c.controllerId
	return mainLayout
}

type WebServerer interface {
	Log() *logrus.Entry
	Config() *config.ServerConfig

	GetMainLayout(r *http.Request) *model.MainLayout
	HtmlWriterRender(w http.ResponseWriter, content interface{}, filenames ...string)

	UrlPathHelper(route string, pairs ...string) *url.URL
}
