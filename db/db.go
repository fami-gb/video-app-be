package db

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDB() *gorm.DB {
	var dsn string
	databaseUrl := os.Getenv("DATABASE_URL")

	if databaseUrl != "" {
		// 1. URLを解析
		u, err := url.Parse(databaseUrl)
		if err != nil {
			log.Fatalln("Invalid DATABASE_URL:", err)
		}

		host := u.Hostname()
		fmt.Printf("🔍 Resolving host: %s\n", host)

		// 2. Google Public DNS (8.8.8.8) を使って強制的にIPv4を解決する
		// (RenderのDNSがIPv6を優先するのを防ぐため)
		resolver := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Millisecond * time.Duration(10000),
				}
				return d.DialContext(ctx, "udp", "8.8.8.8:53")
			},
		}

		// IPv4のみ ("ip4") を要求する
		ips, err := resolver.LookupIP(context.Background(), "ip4", host)
		if err != nil {
			log.Printf("⚠️ DNS Lookup failed: %v. Using original host.\n", err)
		} else {
			// IPv4が見つかった場合
			if len(ips) > 0 {
				ipv4 := ips[0]
				fmt.Printf("✅ Found IPv4: %s (Replacing hostname)\n", ipv4.String())

				// URLのホスト部分をIPアドレスに書き換え
				if u.Port() != "" {
					u.Host = fmt.Sprintf("%s:%s", ipv4.String(), u.Port())
				} else {
					u.Host = ipv4.String()
				}
			} else {
				fmt.Println("⚠️ No IPv4 address found.")
			}
		}

		dsn = u.String()

	} else {
		// ローカル開発用
		dsn = fmt.Sprintf(
			"host=db user=%s password=%s dbname=%s port=5432 sslmode=disable TimeZone=Asia/Tokyo",
			os.Getenv("POSTGRES_USER"),
			os.Getenv("POSTGRES_PASSWORD"),
			os.Getenv("POSTGRES_DB"),
		)
	}

	// 3. 接続
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// 詳細なエラーを出す
		log.Fatalln("❌ DB Connection failed:", err)
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
