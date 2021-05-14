package router

import (
	"github.com/gin-contrib/cors"
	"net/http"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	app := gin.New()

	app.Use(gin.Logger())
	// app.Use(gin.Recovery())
	app.Use(cors.New(cors.Config{
		// AllowOrigins: []string{ /*"http://localhost", "http://localhost:3000", "http://localhost:4000",*/ "*"},
		AllowMethods: []string{"POST", "GET", "PATCH", "DELETE", "HEAD", "PUT", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Content-Length", "Origin", "Accept", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Cache-Control"},
		// ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		AllowCredentials: true,
		AllowWebSockets:  true,
		AllowAllOrigins:  true,
		MaxAge: 12 * time.Hour,
	}))
	app.Use(gzip.Gzip(gzip.DefaultCompression))
	app.Use(middleware.Context())
	app.Use(middleware.RateLimit())
	app.LoadHTMLGlob("public/static/**/*")
	app.Static("/css", config.App.PublicDir+"css")
	app.Static("/images", config.App.UploadDir+"images")
	app.Static("/videos", config.App.UploadDir+"videos")

	app.GET("/", controller.Index)
	app.GET("/api/version", controller.Version)

	app.POST("/api/auth/signup", auth.Signup)
	app.POST("/api/auth/login", auth.Login)
	app.POST("/api/auth/google", auth.GoogleLogin)
	app.POST("/api/auth/facebook", auth.FacebookLogin)
	app.POST("/api/auth/forgot", auth.RestorePassword)
	app.POST("/api/auth/forgot/verify", auth.VerifyRestorePassword)
	app.POST("/api/auth/refresh", middleware.JWTExpired(), auth.Refresh)
	app.POST("/api/auth/logout", middleware.JWT(), auth.Logout)

	app.POST("/api/pushnotification", middleware.JWT(), pushnotification.CreatePushNotification)

	
	app.GET("/api/users/:id", middleware.JWT(), user.GetUser)
	app.PATCH("/api/users/:id", middleware.JWT(), user.UpdateUser)
	app.DELETE("/api/users/:id", middleware.JWT(), user.DeleteUser)

	app.POST("/api/users/images", middleware.JWT(), middleware.MaxSizeBodyAllowed(config.App.MaxMultipartMemory<<20), user.UploadUserImage)
	app.POST("/api/users/username", middleware.JWT(), user.UpdateUserName)

	app.POST("/api/payment", middleware.JWT(), payment.CreatePayment)
	app.POST("/api/payment/do", middleware.JWT(), payment.DoPayment)

	app.POST("/api/comments", middleware.JWT(), comment.CreateComment)
	app.GET("/api/comments", middleware.JWT(), comment.ListComments)
	app.GET("/api/comments/:id", middleware.JWT(), comment.GetComments)
	app.DELETE("/api/comments/:id", middleware.JWT(), comment.DeleteComment)

	app.POST("/api/likes", middleware.JWT(), like.CreateLike)
	app.GET("/api/likes", middleware.JWT(), like.ListLikes)
	app.DELETE("/api/likes/:id", middleware.JWT(), like.DeleteLike)

	app.POST("/api/chats", middleware.JWT(), chat.CreateChat)
	app.GET("/api/chats", middleware.JWT(), chat.ListChat)
	app.GET("/api/chats/:id", middleware.JWT(), chat.GetChat)

	app.GET("/api/search", middleware.JWT(), search.ListSearch)

	node := webchat.GetNode()

	// still need to run gin web server.
	go func() {
		if err := node.Run(); err != nil {
			logger.Fatal(err)
		}
	}()

	websocketHandler := chat.NewWebsocketHandler(node, chat.WebsocketConfig{
		ReadBufferSize:     1024,
		UseWriteBufferPool: true,
		MessageSizeLimit:   config.App.WebsocketMessageSizeLimit << 20, // 1 MiB
	})
	app.GET("/ws", gin.WrapH(middleware.AuthWebChat(websocketHandler)))

	app.GET("/favicon.ico", func(c *gin.Context) {
		c.String(http.StatusNoContent, "")
	})

	app.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{})
	})

	return app
}
