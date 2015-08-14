package render

import (
	"bytes"
	"html/template"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	ContentType = "Content-Type"
	ContentHTML = "text/html; charset=UTF-8"
)

var templates *template.Template

type Renderer struct {
	http.ResponseWriter
	req *http.Request
}

func init() {
	loadTemplates()
}

func loadTemplates() {
	templates = template.New("templates")
	funcs := template.FuncMap{
		"yield": func() (template.HTML, error) {
			return "", nil
		},
	}
	template.Must(templates.Funcs(funcs).ParseGlob("templates/*"))
}

func New(context *gin.Context) *Renderer {
	if gin.IsDebugging() {
		loadTemplates()
	}

	return &Renderer{
		ResponseWriter: context.Writer,
		req:            context.Request,
	}
}

func (r *Renderer) HTML(status int, name string, binding interface{}) {
	addYield(name, binding)

	out, err := execute("layout.tmpl", binding)
	if err != nil {
		http.Error(r, err.Error(), http.StatusInternalServerError)
		return
	}

	// template rendered fine, write out the result
	r.Header().Set(ContentType, ContentHTML)
	r.WriteHeader(status)
	io.Copy(r, out)
}

func execute(name string, binding interface{}) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	return buf, templates.ExecuteTemplate(buf, name, binding)
}

func addYield(name string, binding interface{}) {
	funcs := template.FuncMap{
		"yield": func() (template.HTML, error) {
			buf, err := execute(name, binding)
			return template.HTML(buf.String()), err
		},
	}
	templates.Funcs(funcs)
}
