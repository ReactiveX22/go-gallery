package main

import (
	"example/web-go/controllers"
	"example/web-go/migrations"
	"example/web-go/models"
	"example/web-go/templates"
	"example/web-go/views"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
)

type config struct {
	PSQL models.PostgresConfig
	SMTP models.SMTPConfig
	CSRF struct {
		Key    string
		Secure bool
	}
	Server struct {
		Address string
	}
}

func loadEnvConfig() (config, error) {
	var cfg config

	err := godotenv.Load(".env")
	if err != nil {
		return cfg, fmt.Errorf("load env: %w", err)
	}

	// todo psql
	cfg.PSQL = models.DefaultPostgresConfig()
	// todo smtp
	cfg.SMTP.Host = os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	cfg.SMTP.Port, err = strconv.Atoi(portStr)
	if err != nil {
		return cfg, fmt.Errorf("parse smtp port: %w", err)
	}
	cfg.SMTP.Username = os.Getenv("SMTP_USERNAME")
	cfg.SMTP.Password = os.Getenv("SMTP_PASSWORD")
	// todo csrf
	cfg.CSRF.Key = "8w7nD5Qjv1hN8Kd3gRqZ6L9zB0JvWp2Y"
	cfg.CSRF.Secure = false
	// server
	cfg.Server.Address = "localhost:3000"

	return cfg, nil
}

func main() {
	// load config
	cfg, err := loadEnvConfig()
	if err != nil {
		panic(err)
	}

	// Setup the database
	db, err := models.Open(cfg.PSQL)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = models.MigrateFS(db, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	// Setup services
	userService := &models.UserService{
		DB: db,
	}
	sessionService := &models.SessionService{
		DB: db,
	}
	passwordResetService := &models.PasswordResetService{
		DB: db,
	}
	emailService := models.NewEmailService(cfg.SMTP)

	// Setup middelwares
	umw := controllers.UserMiddleware{
		SessionService: sessionService,
	}
	csrfMw := csrf.Protect([]byte(cfg.CSRF.Key), csrf.Secure(cfg.CSRF.Secure))

	// Setup Controllers
	userC := controllers.User{
		UserService:          userService,
		SessionService:       sessionService,
		PasswordResetService: passwordResetService,
		EmailService:         emailService,
	}
	userC.Templates.New = views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "signup.gohtml"))
	userC.Templates.SignIn = views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "signin.gohtml"))
	userC.Templates.ForgotPassword = views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "forgot-pw.gohtml"))
	userC.Templates.ResetPassword = views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "reset-pw.gohtml"))
	userC.Templates.CheckYourEmail = views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "check-your-email.gohtml"))

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
	r.Get("/forgot-pw", userC.ForgotPassword)
	r.Post("/forgot-pw", userC.ProcessForgotPassword)
	r.Get("/reset-pw", userC.ResetPassword)
	r.Post("/reset-pw", userC.ProcessResetPassword)
	r.NotFound(controllers.StaticHanlder(views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "notFound.gohtml"))))

	// Start the server
	fmt.Printf("The server is listeing on: %s...\n", cfg.Server.Address)
	err = http.ListenAndServe(cfg.Server.Address, r)
	if err != nil {
		panic(err)
	}
}
