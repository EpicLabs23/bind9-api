package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"bind9-api/config"
	"bind9-api/models"
	"bind9-api/utils"

	"github.com/gin-gonic/gin"
)

type DNSController struct {
	Config *config.Config
}

func NewDNSController(cfg *config.Config) *DNSController {
	return &DNSController{Config: cfg}
}

func (dc *DNSController) CreateZone(c *gin.Context) {
	var zone models.Zone
	if err := c.ShouldBindJSON(&zone); err != nil {
		fmt.Print(err)
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Write zone file
	if err := utils.WriteZoneFile(zone, dc.Config); err != nil {
		fmt.Print(err)
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Update Bind9 config
	zoneFilePath := filepath.Join(dc.Config.Bind9.ZoneFileDir, zone.Name+".zone")
	if err := utils.UpdateZoneConfig(dc.Config.Bind9.ConfigFile, zone.Name, zoneFilePath); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Reload Bind9 to apply changes
	success, stdOut, err := utils.CheckAndReload(zone.Name, dc.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: success,
			Error:   "Failed to reload Bind9: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.Response{
		Success: success,
		Message: "Zone created successfully",
		Details: stdOut,
	})
}

func (dc *DNSController) GetZone(c *gin.Context) {
	zoneName := c.Param("zone")
	zoneFilePath := filepath.Join(dc.Config.Bind9.ZoneFileDir, zoneName+".zone")

	zoneData, err := utils.ParseZoneFile(zoneFilePath, zoneName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, zoneData)
}

func (dc *DNSController) UpdateZone(c *gin.Context) {
	zoneName := c.Param("zone")

	zoneText := c.PostForm("zone_text")

	zoneFileName := filepath.Join(dc.Config.Bind9.ZoneFileDir, zoneName+".zone")
	oldContent, err := os.ReadFile(zoneFileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	err = os.WriteFile(zoneFileName, []byte(zoneText), 0644)
	if err != nil {
		if len(oldContent) == 0 {
			os.Remove(zoneFileName)
		} else {
			os.WriteFile(zoneFileName, oldContent, 0644) // Revert to old content
		}
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Reload Bind9 to apply changes
	_, stdOut, err := utils.CheckAndReload(zoneName, dc.Config)
	if err != nil {
		os.WriteFile(zoneFileName, oldContent, 0644) // Revert to old content
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Failed to reload Bind9: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Zone updated successfully",
		Details: stdOut,
	})
}

func (dc *DNSController) DeleteZone(c *gin.Context) {
	zoneName := c.Param("zone")

	// Delete zone file
	if err := utils.DeleteZoneFile(dc.Config.Bind9.ZoneFileDir, zoneName); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Remove zone from config
	if err := utils.RemoveZoneFromConfig(dc.Config.Bind9.ConfigFile, zoneName); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Reload Bind9 to apply changes
	_, stdOut, err := utils.CheckCofigSyntax(dc.Config.Bind9.ConfigFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Failed to reload Bind9: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Zone deleted successfully",
		Details: stdOut,
	})
}

func (dc *DNSController) AddRecord(c *gin.Context) {
	zoneName := c.Param("zone")
	var record models.ResourceRecord

	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	rrStr := utils.MakeRRString(record)
	if rrStr == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Error:   "Invalid record format",
		})
		return
	}

	// Get existing zone
	zoneFilePath := filepath.Join(dc.Config.Bind9.ZoneFileDir, zoneName+".zone")
	oldContent, _ := os.ReadFile(zoneFilePath)
	parsedZone, err := utils.ParseZoneFile(zoneFilePath, zoneName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	parsedZone = append(parsedZone, record)
	updatedZone := models.Zone{
		Name:    zoneName,
		Records: parsedZone,
	}
	utils.WriteZoneFile(updatedZone, dc.Config)

	// Reload Bind9 to apply changes
	_, stdOut, err := utils.CheckAndReload(zoneName, dc.Config)
	if err != nil {
		os.WriteFile(zoneFilePath, oldContent, 0644)
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Failed to reload Bind9: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.Response{
		Success: true,
		Message: "Record added successfully",
		Details: stdOut,
	})
}

func (dc *DNSController) DeleteRecord(c *gin.Context) {
	zoneName := c.Param("zone")
	recordName := c.Param("record")
	recordType := c.Param("type")

	// Get existing zone
	zoneFilePath := filepath.Join(dc.Config.Bind9.ZoneFileDir, zoneName+".zone")
	oldContent, _ := os.ReadFile(zoneFilePath)
	parsedZone, err := utils.ParseZoneFile(zoneFilePath, zoneName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Filter out the record to delete
	var updatedZone []models.ResourceRecord
	for _, r := range parsedZone {
		if !strings.HasSuffix(recordName, ".") {
			recordName += "."
		}

		if r.Name == recordName && r.Type == recordType {
			continue
		}
		updatedZone = append(updatedZone, r)
	}

	if len(updatedZone) == len(parsedZone) {
		c.JSON(http.StatusNotFound, models.Response{
			Success: false,
			Error:   "Record not found",
		})
		return
	}

	parsedZone = updatedZone
	utils.WriteZoneFile(models.Zone{
		Name:    zoneName,
		Records: parsedZone,
	}, dc.Config)

	// Reload Bind9 to apply changes
	_, stdOut, err := utils.CheckAndReload(zoneName, dc.Config)
	if err != nil {
		os.WriteFile(zoneFilePath, oldContent, 0644)
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Failed to reload Bind9: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Record deleted successfully",
		Details: stdOut,
	})
}

func (dc *DNSController) ListZones(c *gin.Context) {
	zones, err := utils.ListZones(dc.Config.Bind9.ZoneFileDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"zones":   zones,
		"count":   len(zones),
	})
}
