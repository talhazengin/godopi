package server

import (
	"godopi/docs"
	"godopi/internal/app/api/controllers"
	. "godopi/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title       Godopi API
// @version     1.0
// @description A Docker Management API
// @BasePath  	/api/v1

func NewRouter() *gin.Engine {
	Logger().Info("Initializing router..")

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	docs.SwaggerInfo.BasePath = "/api/v1"

	health := controllers.HealthController{}
	router.GET("api/v1/health", health.Status)

	v1 := router.Group("api/v1")
	{
		dockerGroup := v1.Group("docker")
		{
			dockerController := controllers.NewDockerController()
			dockerGroup.GET("/containers", dockerController.GetAllContainers)
			dockerGroup.GET("/containers/:id", dockerController.GetDetailedContainer)
			dockerGroup.POST("/containers", dockerController.CreateContainer)
			dockerGroup.DELETE("/containers/:id", dockerController.DeleteContainer)
		}
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	return router
}
