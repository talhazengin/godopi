package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthController struct{}

// Status godoc
// @Summary Checks API Status
// @Tags Health
// @Accept 	json
// @Produce json
// @Success 200 {string} Status
// @Router /health [get]
func (h HealthController) Status(c *gin.Context) {
	c.String(http.StatusOK, "Healthy!")
}
