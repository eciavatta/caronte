package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func ApplicationRoutes(engine *gin.Engine, rulesManager RulesManager) {
	// engine.Static("/", "./frontend/build")

	api := engine.Group("/api")
	{
		api.GET("/rules", func(c *gin.Context) {
			success(c, rulesManager.GetRules())
		})

		api.POST("/rules", func(c *gin.Context) {
			var rule Rule

			if err := c.ShouldBindJSON(&rule); err != nil {
				badRequest(c, err)
				return
			}

			if id, err := rulesManager.AddRule(c, rule); err != nil {
				unprocessableEntity(c, err)
			} else {
				success(c, UnorderedDocument{"id": id})
			}
		})

		api.GET("/rules/:id", func(c *gin.Context) {
			hex := c.Param("id")
			id, err := RowIDFromHex(hex)
			if err != nil {
				badRequest(c, err)
				return
			}
			rule, found := rulesManager.GetRule(id)
			if !found {
				notFound(c, UnorderedDocument{"id": id})
			} else {
				success(c, rule)
			}
		})

		api.PUT("/rules/:id", func(c *gin.Context) {
			hex := c.Param("id")
			id, err := RowIDFromHex(hex)
			if err != nil {
				badRequest(c, err)
				return
			}
			var rule Rule
			if err := c.ShouldBindJSON(&rule); err != nil {
				badRequest(c, err)
				return
			}

			updated := rulesManager.UpdateRule(c, id, rule)
			if !updated {
				notFound(c, UnorderedDocument{"id": id})
			} else {
				success(c, rule)
			}
		})
	}
}

func success(c *gin.Context, obj interface{}) {
	c.JSON(http.StatusOK, obj)
}

func badRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, UnorderedDocument{"result": "error", "error": err.Error()})

	//validationErrors, ok := err.(validator.ValidationErrors)
	//if !ok {
	//	log.WithError(err).WithField("rule", rule).Error("oops")
	//	c.JSON(http.StatusBadRequest, gin.H{})
	//	return
	//}
	//
	//for _, fieldErr := range validationErrors {
	//	log.Println(fieldErr)
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"error": fmt.Sprintf("field '%v' does not respect the %v(%v) rule",
	//			fieldErr.Field(), fieldErr.Tag(), fieldErr.Param()),
	//	})
	//	log.WithError(err).WithField("rule", rule).Error("oops")
	//	return // exit on first error
	//}
}

func unprocessableEntity(c *gin.Context, err error) {
	c.JSON(http.StatusOK, UnorderedDocument{"result": "error", "error": err.Error()})
}

func notFound(c *gin.Context, obj interface{}) {
	c.JSON(http.StatusNotFound, obj)
}
