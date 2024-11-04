package main

import (
	"example/web-go/controllers"
	"example/web-go/migrations"
	"example/web-go/models"
	"example/web-go/templates"
	"example/web-go/views"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
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

	err = models.MigrateFS(db, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	userService := models.UserService{
		DB: db,
	}

	sessionService := models.SessionService{
		DB: db,
	}

	userC := controllers.User{
		UserService:    &userService,
		SessionService: &sessionService,
	}
	userC.Templates.New = views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "signup.gohtml"))
	userC.Templates.SignIn = views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "signin.gohtml"))

	router.Get("/signup", userC.New)
	router.Post("/users", userC.Create)
	router.Get("/signin", userC.SignIn)
	router.Post("/signin", userC.ProcessSignIn)
	router.Get("/users/me", userC.CurrentUser)
	router.Post("/signout", userC.ProcessSignOut)

	router.NotFound(controllers.StaticHanlder(views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "notFound.gohtml"))))

	fmt.Println("The server is listeing on: 3000")

	csrfKey := "8w7nD5Qjv1hN8Kd3gRqZ6L9zB0JvWp2Y"
	csrfMw := csrf.Protect(
		[]byte(csrfKey),
		// TODO: make this true for production
		csrf.Secure(false))
	http.ListenAndServe("localhost:3000", csrfMw(router))

}
