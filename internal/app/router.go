package app

import (
	"net/http"
	"time"

	"diary/config"
	"diary/internal/handler"
	"diary/internal/middleware"
	"diary/internal/repository/mysql"
	"diary/internal/service"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	// middlewares
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(60 * time.Second))
	r.Use(middleware.CORSMiddleware)

	// Repositories
	userRepo := mysql.NewUserRepository(db)
	tagRepo := mysql.NewTagRepository(db)
	todoRepo := mysql.NewTodoRepository(db)
	imageRepo := mysql.NewImageRepository(db)
	diaryRepo := mysql.NewDiaryRepository(db)

	// Services
	userService := service.NewUserService(userRepo, cfg)
	tagService := service.NewTagService(tagRepo)
	todoService := service.NewTodoService(todoRepo)
	imageService := service.NewImageService(imageRepo, cfg)
	diaryService := service.NewDiaryService(diaryRepo, tagRepo, imageRepo, cfg)

	// Handlers
	userHandler := handler.NewUserHandler(userService)
	tagHandler := handler.NewTagHandler(tagService)
	todoHandler := handler.NewTodoHandler(todoService)
	imageHandler := handler.NewImageHandler(imageService)
	diaryHandler := handler.NewDiaryHandler(diaryService)
	statsHandler := handler.NewStatsHandler(diaryRepo, todoRepo)
	exportHandler := handler.NewExportHandler(diaryService)

	// public
	// if cfg.EnableRegistration {
	r.Post("/api/register", userHandler.Register)
	// }
	r.Post("/api/login", userHandler.Login)
	r.Get("/api/diaries/public", diaryHandler.ListPublic)

	// static
	fs := http.FileServer(http.Dir(cfg.UploadDir))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", fs))

	// protected
	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg, db))

		// User
		r.Route("/user", func(r chi.Router) {
			r.Get("/profile", userHandler.GetProfile)
			r.Put("/username", userHandler.UpdateUsername)
			r.Put("/password", userHandler.UpdatePassword)
			r.Delete("/", userHandler.DeleteUser)
		})

		// Stats
		r.Get("/stats/dashboard", statsHandler.GetDashboardStats)

		// Export
		r.Get("/diaries/{id}/export", exportHandler.ExportSingle)
		r.Post("/diaries/export", exportHandler.ExportBatch)

		// Tags
		r.Route("/tags", func(r chi.Router) {
			r.Post("/", tagHandler.Create)
			r.Get("/", tagHandler.List)
			r.Get("/popular", tagHandler.GetPopular)
			r.Put("/{id}", tagHandler.Update)
			r.Delete("/{id}", tagHandler.Delete)
		})

		// Todos
		r.Route("/todos", func(r chi.Router) {
			r.Post("/", todoHandler.Create)
			r.Get("/", todoHandler.List)
			r.Get("/stats", todoHandler.GetStats)
			r.Route("/{id}", func(r chi.Router) {
				r.Put("/", todoHandler.Update)
				r.Delete("/", todoHandler.Delete)
				r.Patch("/done", todoHandler.MarkAsDone)
				r.Patch("/undone", todoHandler.MarkAsUndone)
			})
		})

		// Images
		r.Route("/images", func(r chi.Router) {
			r.Post("/upload", imageHandler.Upload)
			r.Get("/", imageHandler.List)
			r.Route("/{id}", func(r chi.Router) {
				r.Delete("/", imageHandler.Delete)
				r.Post("/attach", imageHandler.AttachToDiary)
			})
		})

		// Diaries
		r.Route("/diaries", func(r chi.Router) {
			r.Post("/", diaryHandler.Create)
			r.Get("/", diaryHandler.List)
			r.Get("/search", diaryHandler.Search)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", diaryHandler.GetByID)
				r.Put("/", diaryHandler.Update)
				r.Delete("/", diaryHandler.Delete)
				r.Post("/pin", diaryHandler.TogglePin)
			})
		})
	})

	return r
}
