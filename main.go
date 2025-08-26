package main

import (
	"bind9-api/config"
	"bind9-api/controllers"
	"flag"

	"github.com/gin-gonic/gin"
)

func main() {
	flagConfig := flag.String("config", ".", "Path to the configuration file")
	flag.Parse()
	cfg, err := config.LoadConfig(*flagConfig)
	if err != nil {
		panic(err)
	}

	// Initialize Gin
	r := gin.Default()

	// Authentication
	accounts := gin.Accounts{}
	for _, user := range cfg.ApiAccess {
		accounts[user.Username] = user.Password
	}
	r.Use(gin.BasicAuth(accounts))

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
