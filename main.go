package main

import (
	"example/web-go/controllers"
	"example/web-go/models"
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

	cfg := models.DefaultPostgresConfig()
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	userService := models.UserService{
		DB: db,
	}

	userC := controllers.User{
		UserService: &userService,
	}
	userC.Templates.New = views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "signup.gohtml"))
	userC.Templates.SignIn = views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "signin.gohtml"))
	router.Get("/signup", userC.New)
	router.Post("/users", userC.Create)
	router.Get("/signin", userC.SignIn)
	router.Post("/signin", userC.ProcessSignIn)
	router.Get("/users/me", userC.CurrentUser)

	router.NotFound(controllers.StaticHanlder(views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "notFound.gohtml"))))

	fmt.Println("The server is listeing on: 3000")

	http.ListenAndServe("localhost:3000", router)

}
