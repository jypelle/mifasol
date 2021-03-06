package webSrv

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
)

func (d *WebServer) rawHtmlRender(w io.Writer, content interface{}, filenames ...string) {

	// Parsing template files
	t := template.New("").Funcs(d.templateHelpers)

	for _, filename := range filenames {

		file, e := d.TemplatesFs.Open(filename)
		if e != nil {
			d.log.Panicf("Unable to open template file %s: %v\n", filename, e)
		}
		defer file.Close()

		buf := bytes.NewBuffer(nil)
		_, e = io.Copy(buf, file)
		if e != nil {
			d.log.Panicf("Unable to read template file %s: %v\n", filename, e)
		}

		t, e = t.Parse(string(buf.Bytes()))
		if e != nil {
			d.log.Panicf("Unable to interpret template file %s: %v\n", filename, e)
		}
	}

	e := t.Execute(w, content)
	if e != nil {
		d.log.Panicf("Unable to execute template files : %v\n", e)
	}

}

func (d *WebServer) HtmlWriterRender(w http.ResponseWriter, content interface{}, filenames ...string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	d.rawHtmlRender(w, content, filenames...)
}
