package main

import (
	"bind9-api/config"
	"bind9-api/controllers"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	// Initialize Gin
	r := gin.Default()

	// Initialize controller
	dnsController := controllers.NewDNSController(cfg)

	// Routes
	api := r.Group("/api/v1/zones")
	{
		api.GET("/", dnsController.ListZones)
		api.POST("/", dnsController.CreateZone)
		api.GET("/:zone", dnsController.GetZone)
		api.PUT("/:zone", dnsController.UpdateZone)
		api.DELETE("/:zone", dnsController.DeleteZone)

		// Record operations
		api.POST("/:zone/records", dnsController.AddRecord)
		api.DELETE("/:zone/records/:record/:type", dnsController.DeleteRecord)
	}

	// Start server
	r.Run(":" + cfg.Server.Port)
}
