package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/auth"
	"github.com/Dimes2212/5th_semester_course_work_backend/internal/listing"
	"github.com/Dimes2212/5th_semester_course_work_backend/internal/lookup"
	"github.com/Dimes2212/5th_semester_course_work_backend/internal/s3"
	"github.com/Dimes2212/5th_semester_course_work_backend/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	//бд
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://campuslost:campuslost@localhost:5436/campuslost?sslmode=disable"
	}

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("create PostgreSQL pool: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("connect to PostgreSQL: %v", err)
	}

	log.Println("connected to PostgreSQL")

	// настройка s3
	s3UseSSL, err := strconv.ParseBool(
		os.Getenv("S3_USE_SSL"),
	)
	if err != nil {
		log.Fatalf("parse S3_USE_SSL: %v", err)
	}

	objectStorage, err := s3.NewClient(s3.Config{
		Endpoint:        os.Getenv("S3_ENDPOINT"),
		Region:          os.Getenv("S3_REGION"),
		Bucket:          os.Getenv("S3_BUCKET"),
		AccessKeyID:     os.Getenv("S3_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("S3_SECRET_ACCESS_KEY"),
		UseSSL:          s3UseSSL,
	})
	if err != nil {
		log.Fatalf("create S3 client: %v", err)
	}

	storageCheckCtx, cancelStorageCheck := context.WithTimeout(ctx, 10*time.Second)

	err = objectStorage.CheckBucket(storageCheckCtx)
	cancelStorageCheck()

	if err != nil {
		log.Fatalf("connect to S3 bucket: %v", err)
	}

	log.Println("connected to S3 bucket")
	// настройка секреток
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "local-development-secret-change-me-1234567890"
	}

	jwtIssuer := os.Getenv("JWT_ISSUER")
	if jwtIssuer == "" {
		jwtIssuer = "campuslost"
	}

	// репа авторизации
	authRepository := auth.NewRepository(db)

	//сервис авторизации
	authService := auth.NewService(
		authRepository,
		jwtSecret,
		jwtIssuer,
		15*time.Minute,
		30*24*time.Hour,
	)

	// ручка и мв авторизации
	authHandler := auth.NewHandler(authService)
	authMiddleware := auth.NewMiddleware(authService)

	// профиль пользователя
	userRepository := user.NewRepository(db)
	userService := user.NewService(userRepository)
	userHandler := user.NewHandler(userService)

	//лукапы для категорий
	lookupRepository := lookup.NewRepository(db)
	lookupService := lookup.NewService(lookupRepository)
	lookupHandler := lookup.NewHandler(lookupService)

	//листинги
	listingRepository := listing.NewRepository(db)
	listingService := listing.NewService(listingRepository, objectStorage)
	listingHandler := listing.NewHandler(listingService)

	//роутер
	router := chi.NewRouter()

	// встроенные миддлвары
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	//маршруты
	router.Route("/api/v1/auth", func(router chi.Router) {
		router.Post("/register", authHandler.Register)
		router.Post("/login", authHandler.Login)
		router.Post("/refresh", authHandler.Refresh)

		router.With(authMiddleware.Authenticate).
			Post("/logout", authHandler.Logout)
	})

	router.Get("/api/v1/categories", lookupHandler.GetCategories)
	router.Get("/api/v1/locations", lookupHandler.GetLocations)

	router.Get("/api/v1/listings", listingHandler.GetListingsFeed)

	router.Get("/api/v1/listings/{listingID}", listingHandler.GetListingByID)

	router.Group(func(router chi.Router) {
		router.Use(authMiddleware.Authenticate)

		router.Get("/api/v1/me", userHandler.GetMe)
		router.Patch("/api/v1/me", userHandler.UpdateName)
		router.Delete("/api/v1/me", userHandler.DeleteMe)
		router.Post("/api/v1/listings", listingHandler.CreateListing)
		router.Put("/api/v1/listings/{listingID}", listingHandler.UpdateListing)
		router.Post("/api/v1/listings/{listingID}/resolve", listingHandler.ResolveListing)
		router.Delete("/api/v1/listings/{listingID}", listingHandler.DeleteListing)
		router.Get("/api/v1/me/listings", listingHandler.GetMyListings)
		router.Post("/api/v1/listings/{listingID}/photos", listingHandler.AddListingPhoto)
		router.Delete("/api/v1/listings/{listingID}/photos/{photoID}", listingHandler.DeleteListingPhoto)
	})

	// запуск сервера
	httpAddress := os.Getenv("HTTP_ADDRESS")
	if httpAddress == "" {
		httpAddress = ":8080"
	}

	server := &http.Server{
		Addr:              httpAddress,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)

	// graceful shutdown
	go func() {
		log.Printf("HTTP server is listening on %s", httpAddress)

		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	select {
	case err := <-serverErrors:
		log.Printf("run HTTP server: %v", err)
		return
	case <-ctx.Done():
		log.Println("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf(
			"shutdown HTTP server: %v",
			err,
		)
	}

}
