package main

import (
	"example/web-go/controllers"
	"example/web-go/templates"
	"example/web-go/views"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	router := chi.NewRouter()

	router.Get("/", controllers.StaticHanlder(views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "home.gohtml"))))
	router.Get("/contact", controllers.StaticHanlder(views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "contact.gohtml"))))
	router.Get("/faq", controllers.FAQ(views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "faq.gohtml"))))

	userC := controllers.User{}
	userC.Templates.New = views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "signup.gohtml"))
	router.Get("/signup", userC.New)
	router.Post("/users", userC.Create)

	router.NotFound(controllers.StaticHanlder(views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "notFound.gohtml"))))

	fmt.Println("The server is listeing on: 3000")

	http.ListenAndServe("localhost:3000", router)

}
