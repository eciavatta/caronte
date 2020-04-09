package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func ApplicationRoutes(engine *gin.Engine) {
	engine.Static("/", "./frontend/build")

	api := engine.Group("/api")
	{
		api.POST("/rules", func(c *gin.Context) {
			var rule Rule

			if err := c.ShouldBindJSON(&rule); err != nil {
				for _, fieldErr := range err.(validator.ValidationErrors) {
					log.Println(fieldErr)
					c.JSON(http.StatusBadRequest, gin.H{
						"error": fmt.Sprintf("field '%v' does not respect the %v(%v) rule",
							fieldErr.Field(), fieldErr.Tag(), fieldErr.Param()),
					})
					log.WithError(err).WithField("rule", rule).Panic("oops")
					return // exit on first error
				}
			}

			c.JSON(200, rule)
		})
	}
}
