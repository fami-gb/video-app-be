package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fami-gb/video-app-be/db"
	"github.com/fami-gb/video-app-be/storage"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
)

// ビデオの付加情報としてタグなども追加できるようにしたい
type Video struct {
	gorm.Model
	Title    string `json:"title"`
	URL      string `json:"url"`
	VideoKey string `json:"video_key"`
}

// フロントからの動画登録リクエスト用構造体
type CreateVideoRequest struct {
	Title    string `json:"title"`
	VideoKey string `json:"video_key"`
}

func main() {
	// Video構造体を元にテーブルを自動生成(php artisan migrate的な)
	// Video{} : Video構造体のインスタンスを作成
	// &Video{} : Video構造体のインスタンスのアドレスを渡す(gormがポインタで受け取るため)
	database := db.NewDB()
	database.AutoMigrate(&Video{})

	r2Manager, err := storage.NewR2Manager()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize R2 Manager: %v", err))
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// リクエストごとにdbとr2をコンテキストにセットする(すべてのエンドポイントのリクエスト前に実行される)
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("db", database)
			c.Set("r2", r2Manager)
			return next(c)
		}
	})

	// ヘルスチェックみたいなapi
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Video App Backend is running!",
		})
	})

	// とりま/apiでまとめておく
	api := e.Group("/api")

	// 全ビデオ取得
	api.GET("/videos", func(c echo.Context) error {
		db := c.Get("db").(*gorm.DB)
		var videos []Video
		db.Find(&videos) // SELECT * FROM videos;
		return c.JSON(http.StatusOK, videos)
	})

	// 動画情報の登録(アップロード完了後に呼び出し)
	// フロントから動画のタイトルなどを受け取って保存する
	api.POST("/videos", func(c echo.Context) error {
		var input CreateVideoRequest
		if err := c.Bind(&input); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid input",
			})
		}

		db := c.Get("db").(*gorm.DB)

		// 環境変数からPublic Domainを取得
		publicDomain := os.Getenv("PUBLIC_DOMAIN")
		if publicDomain == "" {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Public domain configuration missing",
			})
		}

		// 保存データ作成
		video := Video{
			Title:    input.Title,
			URL:      fmt.Sprintf("%s/%s", publicDomain, input.VideoKey),
			VideoKey: input.VideoKey,
		}

		// dbにCreateあったっけ？
		if err := db.Create(&video).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to save video",
			})
		}

		return c.JSON(http.StatusCreated, video)
	})

	api.POST("/upload-url", func(c echo.Context) error {
		var input struct {
			Filename string `json:"filename"`
		}
		if err := c.Bind(&input); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid input",
			})
		}

		objectKey := fmt.Sprintf("%d-%s", time.Now().Unix(), input.Filename)

		r2 := c.Get("r2").(*storage.R2Manager)

		url, err := r2.GenerateUploadURL(objectKey)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to generate upload URL",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"uploadUrl": url,
			"key":       objectKey,
		})
	})

	// 動画削除
	api.DELETE("/videos/:id", func(c echo.Context) error {
		id := c.Param("id")
		db := c.Get("db").(*gorm.DB)
		r2 := c.Get("r2").(*storage.R2Manager)

		var video Video
		if err := db.First(&video, id).Error; err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Video not found",
			})
		}

		if err := r2.DeleteFile(video.VideoKey); err != nil {
			fmt.Printf("Failed to delete from R2: %v\n", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to delete video from storage",
			})
		}

		if err := db.Unscoped().Delete(&video).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to delete video from database",
			})
		}
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Video deleted successfully",
		})
	})

	e.Logger.Fatal(e.Start(":8080"))
}
