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

	cfg.PSQL = models.PostgresConfig{
		Host:     os.Getenv("PSQL_HOST"),
		Port:     os.Getenv("PSQL_PORT"),
		User:     os.Getenv("PSQL_USER"),
		Password: os.Getenv("PSQL_PASSWORD"),
		DBName:   os.Getenv("PSQL_DBNAME"),
		SSLMode:  os.Getenv("PSQL_SSLMODE"),
	}
	if cfg.PSQL.Host == "" && cfg.PSQL.Port == "" {
		return cfg, fmt.Errorf("no PSQL config provided.")
	}

	cfg.SMTP.Host = os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	cfg.SMTP.Port, err = strconv.Atoi(portStr)
	if err != nil {
		return cfg, fmt.Errorf("parse smtp port: %w", err)
	}
	cfg.SMTP.Username = os.Getenv("SMTP_USERNAME")
	cfg.SMTP.Password = os.Getenv("SMTP_PASSWORD")

	cfg.CSRF.Key = os.Getenv("CSRF_KEY")
	cfg.CSRF.Secure = os.Getenv("CSRF_SECURE") == "true"
	cfg.Server.Address = os.Getenv("SERVER_ADDRESS")

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
	galleryService := &models.GalleryService{
		DB: db,
	}

	// Setup middelwares
	umw := controllers.UserMiddleware{
		SessionService: sessionService,
	}
	csrfMw := csrf.Protect([]byte(cfg.CSRF.Key), csrf.Secure(cfg.CSRF.Secure), csrf.Path("/"))

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
	userC.Templates.CheckYourEmail = views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "check-your-email.gohtml"))

	galleriesC := controllers.Galleries{
		GalleryService: galleryService,
	}
	galleriesC.Templates.Index = views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "galleries/index.gohtml"))
	galleriesC.Templates.Show = views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "galleries/show.gohtml"))
	galleriesC.Templates.New = views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "galleries/new.gohtml"))
	galleriesC.Templates.Edit = views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "galleries/edit.gohtml"))

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
	r.Route("/galleries", func(r chi.Router) {
		r.Get("/{id}", galleriesC.Show)
		r.Get("/{id}/images/{filename}", galleriesC.Image)
		r.Group(func(r chi.Router) {
			r.Use(umw.RequireUser)
			r.Get("/new", galleriesC.New)
			r.Get("/{id}/edit", galleriesC.Edit)
			r.Post("/{id}/delete", galleriesC.Delete)
			r.Get("/", galleriesC.Index)
			r.Get("/{id}", galleriesC.Index)
			r.Post("/", galleriesC.Create)
			r.Post("/{id}", galleriesC.Update)
			r.Post("/{id}/images", galleriesC.UploadImage)
			r.Post("/{id}/images/{filename}/delete", galleriesC.DeleteImage)
		})
		r.Get("/{id}", galleriesC.Show)
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
