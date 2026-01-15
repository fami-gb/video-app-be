package db

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDB() *gorm.DB {
	var dsn string

	// 本番環境 (Render) かどうかを環境変数で判断
	if databaseUrl := os.Getenv("DATABASE_URL"); databaseUrl != "" {
		// URLを解析
		u, err := url.Parse(databaseUrl)
		if err != nil {
			log.Fatalln("Invalid DATABASE_URL:", err)
		}

		// ホスト名からIPv4アドレスを強制的に解決する
		// (RenderはIPv6のアウトバウンド通信に弱いため、明示的にIPv4を使う)
		ips, err := net.LookupIP(u.Hostname())
		if err != nil {
			log.Printf("Failed to lookup IP for host %s: %v", u.Hostname(), err)
		} else {
			for _, ip := range ips {
				if ipv4 := ip.To4(); ipv4 != nil {
					// IPv4が見つかったら、URLのホスト部分をそのIPアドレスに書き換える
					fmt.Printf("Force resolving host %s to IPv4: %s\n", u.Hostname(), ipv4.String())
					if u.Port() != "" {
						u.Host = fmt.Sprintf("%s:%s", ipv4.String(), u.Port())
					} else {
						u.Host = ipv4.String()
					}
					break
				}
			}
		}
		// 書き換えた（または元の）URLを使用
		dsn = u.String()

	} else {
		// ローカル開発環境 (Docker Compose)
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

func CloseDB(db *gorm.DB) {
	sqlDB, _ := db.DB()
	if err := sqlDB.Close(); err != nil {
		log.Fatalln("Error closing database connection:", err)
	}
}
