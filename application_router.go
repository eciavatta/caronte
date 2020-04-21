package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	log "github.com/sirupsen/logrus"
)

func CreateApplicationRouter(applicationContext *ApplicationContext) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// engine.Static("/", "./frontend/build")

	router.POST("/setup", func(c *gin.Context) {
		var settings struct {
			Config   Config       `json:"config"`
			Accounts gin.Accounts `json:"accounts"`
		}

		if err := c.ShouldBindJSON(&settings); err != nil {
			badRequest(c, err)
			return
		}

		applicationContext.SetConfig(settings.Config)
		applicationContext.SetAccounts(settings.Accounts)

		c.JSON(http.StatusAccepted, gin.H{})
	})

	api := router.Group("/api")
	api.Use(SetupRequiredMiddleware(applicationContext))
	api.Use(AuthRequiredMiddleware(applicationContext))
	{
		api.GET("/rules", func(c *gin.Context) {
			success(c, applicationContext.RulesManager.GetRules())
		})

		api.POST("/rules", func(c *gin.Context) {
			var rule Rule

			if err := c.ShouldBindJSON(&rule); err != nil {
				badRequest(c, err)
				return
			}

			if id, err := applicationContext.RulesManager.AddRule(c, rule); err != nil {
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
			rule, found := applicationContext.RulesManager.GetRule(id)
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

			updated, err := applicationContext.RulesManager.UpdateRule(c, id, rule)
			if err != nil {
				badRequest(c, err)
			} else if !updated {
				notFound(c, UnorderedDocument{"id": id})
			} else {
				success(c, rule)
			}
		})
	}

	return router
}

func SetupRequiredMiddleware(applicationContext *ApplicationContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Error("aaaaaaaaaaaaaa")
		if !applicationContext.IsConfigured {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "setup required",
				"url":   c.Request.Host + "/setup",
			})
		} else {
			c.Next()
		}
	}
}

func AuthRequiredMiddleware(applicationContext *ApplicationContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !applicationContext.Config.AuthRequired {
			c.Next()
			return
		}

		gin.BasicAuth(applicationContext.Accounts)(c)
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
	c.JSON(http.StatusUnprocessableEntity, UnorderedDocument{"result": "error", "error": err.Error()})
}

func notFound(c *gin.Context, obj interface{}) {
	c.JSON(http.StatusNotFound, obj)
}
