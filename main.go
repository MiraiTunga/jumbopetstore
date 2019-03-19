package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/joho/godotenv/autoload"
	"net/http"
	"os"
	"petstore/handlers"
	"strings"
)

func init() {

}

func main() {

	sess, _ := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-2"),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), "")},
	)
	svc := dynamodb.New(sess)

	//s3:=s

	port := os.Getenv("PORT")

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(middlewareHeadersByRequestURI())
	router.LoadHTMLGlob("static/*.html")
	router.Static("static", "static")

	router.GET("/", func(c *gin.Context) {

		c.HTML(http.StatusOK, "index.html", nil)
	})

	apiRoutes := router.Group("/api")

	apiRoutes.POST("/pet", handlers.AddPet(svc))
	apiRoutes.PUT("/pet", handlers.AddPet(svc))

	/*we need to do this because of gin does not have strong wild card matching yet :(*/
	apiRoutes.GET("/pet/:petId", handlers.ResolveRouteConflict(svc))
	//apiRoutes.GET("/pet/findByStatus",handlers.ResolveRoute(svc))

	apiRoutes.POST("/pet/:petId", handlers.UpdatePetFormData(svc))

	apiRoutes.DELETE("/pet/:petId", handlers.DeletePet(svc))

	apiRoutes.POST("/pet/:petId/uploadImage", handlers.UpLoadImage(svc))

	_ = router.Run(":" + port)
}

func middlewareHeadersByRequestURI() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.RequestURI, "/static/") {
			c.Header("Cache-Control", "no-cache")
		} else if strings.HasPrefix(c.Request.RequestURI, "/images/") {
			c.Header("Cache-Control", "no-cache")
		} else if strings.HasPrefix(c.Request.RequestURI, "/api/") {
			c.Header("Content-Type", "application/json")
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers")
		}
	}
}
