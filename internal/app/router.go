package router


import (
	"net/http"
	"time"


	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
	"diary/internal/handler"
	"diary/internal/middleware"
	"diary/config"
	"diary/internal/repository/mysql"
	"diary/internal/service"
)


func SetupRouter(db *gorm.DB, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	// middlewares
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(60 * time.Second))

	userRepo := mysql.NewUserRepository(db)
	
	userService := service.NewUserService(userRepo, cfg)
	
	userHandler := handler.NewUserHandler(userService)

	// public
	r.Post("/api/register", userHandler.Register)
	r.Post("/api/login", userHandler.Login)


	// static
	fs := http.FileServer(http.Dir(cfg.UploadDir))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", fs))


	// protected
	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg, db))

		r.Route("/user", func(r chi.Router) {
			r.Get("/profile", userHandler.GetProfile)        // 获取当前用户信息
			r.Put("/username", userHandler.UpdateUsername)   // 更新用户名
			r.Put("/password", userHandler.UpdatePassword)   // 更新密码
			r.Delete("/", userHandler.DeleteUser)            // 删除账户
		})


		// diary
		// r.Route("/diaries", func(r chi.Router) {
		// 	r.Get("/", ListDiariesHandler(db, cfg))
		// 	r.Post("/", CreateDiaryHandler(db, cfg))
		// 	r.Get("/tags", ListTagsHandler(db, cfg))
		// 	r.Get("/{id}", GetDiaryHandler(db, cfg))
		// 	r.Put("/{id}", UpdateDiaryHandler(db, cfg))
		// 	r.Delete("/{id}", DeleteDiaryHandler(db, cfg))
		// })


		// todos
		// r.Get("/todos", ListTodosHandler(db, cfg))
		// r.Post("/todos", CreateTodoHandler(db, cfg))
		// todo detail by pattern /api/todos/{id}
		// r.Route("/todos/{id}", func(r chi.Router) {
		// 	r.Get("/", TodoDetailHandler(db, cfg))
		// 	r.Put("/", TodoDetailHandler(db, cfg))
		// 	r.Delete("/", TodoDetailHandler(db, cfg))
		// })


		// upload
		// r.Post("/upload", handler.UploadHandler(db, cfg))
	})


	return r
}