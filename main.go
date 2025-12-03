package main


import (
	"log"
	"net/http"
	"os"


	"diary/config"
	"diary/internal"
	"diary/internal/app"
	"diary/internal/database"
	"github.com/joho/godotenv"
)

func EnsureUploadDir(dir string, cfg *config.Config) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

func main() {
	_ = godotenv.Load()


	cfg := config.LoadConfig()
	db := mysql.InitDB(cfg)
	defer mysql.CloseDB(db)


	if err := models.AutoMigrate(db); err != nil {
		log.Fatalf("migrate failed: %v", err)
	}


	err := EnsureUploadDir(cfg.UploadDir, cfg)
	if err != nil {
		log.Fatalf("create upload dir failed: %v", err)
	}


	r := router.SetupRouter(db, cfg)
	addr := ":" + cfg.Port
	log.Printf("server running at %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server err: %v", err)
	}
}