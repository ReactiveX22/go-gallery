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
	// Setup the database
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

	// Setup services
	userService := models.UserService{
		DB: db,
	}
	sessionService := models.SessionService{
		DB: db,
	}

	// Setup middelwares
	umw := controllers.UserMiddleware{
		SessionService: &sessionService,
	}
	csrfKey := "8w7nD5Qjv1hN8Kd3gRqZ6L9zB0JvWp2Y"
	csrfMw := csrf.Protect([]byte(csrfKey), csrf.Secure(false)) // TODO: make this true for production

	// Setup Controllers
	userC := controllers.User{
		UserService:    &userService,
		SessionService: &sessionService,
	}
	userC.Templates.New = views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "signup.gohtml"))
	userC.Templates.SignIn = views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "signin.gohtml"))

	// Setup r and routes
	r := chi.NewRouter()
	r.Use(csrfMw)
	r.Use(umw.SetUser)
	r.Get("/", controllers.StaticHanlder(views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "home.gohtml"))))
	r.Get("/contact", controllers.StaticHanlder(views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "contact.gohtml"))))
	r.Get("/faq", controllers.FAQ(views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "faq.gohtml"))))
	r.Get("/signup", userC.New)
	r.Post("/users", userC.Create)
	r.Get("/signin", userC.SignIn)
	r.Post("/signin", userC.ProcessSignIn)
	r.Route("/users/me", func(r chi.Router) {
		r.Use(umw.RequireUser)
		r.Get("/", userC.CurrentUser)
	})
	r.Route("/app", func(r chi.Router) {

	})
	r.Post("/signout", userC.ProcessSignOut)
	r.NotFound(controllers.StaticHanlder(views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "notFound.gohtml"))))

	// Start the server
	fmt.Println("The server is listeing on: 3000")
	http.ListenAndServe("localhost:3000", r)
}
