package webSrv

import (
	"bytes"
	"html/template"
	"net/url"
)

func (d *WebServer) UrlPathHelper(route string, pairs ...string) *url.URL {
	url, e := d.router.Get(route).URLPath(pairs...)
	if e != nil {
		d.Log().Errorf("Unable to create URL Path: %v", e)
		return nil
	}
	return url
}

func (d *WebServer) UrlPathQueryHelper(route string, rawQuery string, pairs ...string) *url.URL {
	u, e := d.router.Get(route).URLPath(pairs...)
	if e != nil {
		d.Log().Errorf("Unable to create URL Path with query: %v", e)
		return nil
	}
	u.RawQuery = rawQuery
	return u
}

func (d *WebServer) QueryParamHelper(pairs ...string) string {
	v := url.Values{}

	ind := 0
	key := ""

	for _, elmt := range pairs {
		if ind == 0 {
			key = elmt
			ind = 1
		} else {
			v.Add(key, elmt)
			ind = 0
		}
	}

	return v.Encode()
}

func (d *WebServer) PartialHelper(filename string, content interface{}) template.HTML {
	var w bytes.Buffer

	d.rawHtmlRender(&w, content, filename)

	return template.HTML(w.String())
}
