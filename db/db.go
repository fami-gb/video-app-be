package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDB() *gorm.DB {
	var dsn string
	// docker-compose.yml で設定した環境変数を利用
	// ホスト名が "db" になっているのがDocker通信のポイントです
	if databaseUrl := os.Getenv("DATABASE_URL"); databaseUrl != "" {
		dsn = databaseUrl
	} else {
		dsn = fmt.Sprintf(
			"host=db user=%s password=%s dbname=%s port=5432 sslmode=disable TimeZone=Asia/Tokyo",
			os.Getenv("POSTGRES_USER"),
			os.Getenv("POSTGRES_PASSWORD"),
			os.Getenv("POSTGRES_DB"),
		)
	}

	// DBへの接続試行
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalln("DB Connection failed:", err)
	}

	fmt.Println("🚀 Connected to the database successfully!")

	return db
}

// CloseDB はDB接続を閉じるためのヘルパー関数です（必要に応じて使用）
func CloseDB(db *gorm.DB) {
	sqlDB, _ := db.DB()
	if err := sqlDB.Close(); err != nil {
		log.Fatalln("Error closing database connection:", err)
	}
}
