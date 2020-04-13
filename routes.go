package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/go-playground/validator/v10"
	"net/http"
)

func ApplicationRoutes(engine *gin.Engine) {
	engine.Static("/", "./frontend/build")

	api := engine.Group("/api")
	{
		api.POST("/rule", func(c *gin.Context) {
			var rule Rule

			//data, _ := c.GetRawData()
			//
			//var json = jsoniter.ConfigCompatibleWithStandardLibrary
			//err := json.Unmarshal(data, &filter)
			//
			//if err != nil {
			//	log.WithError(err).Error("failed to unmarshal")
			//	c.String(500, "failed to unmarshal")
			//}
			//
			//err = validator.New().Struct(filter)
			//log.WithError(err).WithField("filter", filter).Error("aaaa")
			//c.String(200, "ok")

			if err := c.ShouldBindJSON(&rule); err != nil {
				validationErrors, ok := err.(validator.ValidationErrors)
				if !ok {
					log.WithError(err).WithField("rule", rule).Error("oops")
					c.JSON(http.StatusBadRequest, gin.H{})
					return
				}

				for _, fieldErr := range validationErrors {
					log.Println(fieldErr)
					c.JSON(http.StatusBadRequest, gin.H{
						"error": fmt.Sprintf("field '%v' does not respect the %v(%v) rule",
							fieldErr.Field(), fieldErr.Tag(), fieldErr.Param()),
					})
					log.WithError(err).WithField("rule", rule).Error("oops")
					return // exit on first error
				}
			}

			c.JSON(200, rule)
		})
	}
}
