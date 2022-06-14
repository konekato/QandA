package main

import (
	"fmt"
	"io"
	"kone-server/db"
	"kone-server/handler"
	"log"
	"os"
	"text/template"

	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/lib/pq"
)

func main() {
	if err := godotenv.Load(fmt.Sprintf("%s.env", os.Getenv("GO_ENV"))); err != nil {
		log.Fatal(err)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Renderer = initView()

	db.Init()

	handler.InitRouting(e)

	e.Start(":" + os.Getenv("PORT"))
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func initView() *Template {
	return &Template{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}
