package main

import (
	"fmt"
	"food-delivery/component/appctx"
	"food-delivery/component/uploadprovider"
	"food-delivery/middleware"
	"log"
	"net/http"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Restaurant struct {
	Id   int    `json:"id" gorm:"column:id;"`
	Name string `json:"name" gorm:"column:name;"`
	Addr string `json:"addr" gorm:"column:addr;"`
}

// trả về tên bảng trong DB
func (Restaurant) TableName() string {
	return "restaurants"
}

type RestaurantUpdate struct {
	Name *string `json:"name" gorm:"column:name;"`
	Addr *string `json:"addr" gorm:"column:addr;"`
}

func (RestaurantUpdate) TableName() string {
	return Restaurant{}.TableName()
}

func main() {
	// Lay bien env tu file .env
	envErr := godotenv.Load(".env")
	if envErr != nil {
		fmt.Println("Khong the mo file .env")
		os.Exit(1)
	}

	dsn := os.Getenv("MYSQL_CONN_STRING")

	s3BucketName := os.Getenv("S3_BUCKET_NAME")
	s3Region := os.Getenv("S3_REGION")
	s3APIKey := os.Getenv("S3_API_KEY")
	s3SecretKey := os.Getenv("S3_SECRET")
	s3Domain := os.Getenv("S3_DOMAIN")
	secretKey := os.Getenv("SYSTEM_SECRET")

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}

	db = db.Debug()
	s3Provider := uploadprovider.NewS3Provider(s3BucketName, s3Region, s3APIKey, s3SecretKey, s3Domain)

	appContext := appctx.NewAppContext(db, s3Provider, secretKey)

	r := gin.Default()
	r.Use(middleware.Recover(appContext))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.Static("/static", "./static")

	// POST /v1/restaurants
	v1 := r.Group("/v1")

	setupRoute(appContext, v1)
	setupAdminRoute(appContext, v1)

	r.Run()

}
