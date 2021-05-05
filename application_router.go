/*
 * This file is part of caronte (https://github.com/eciavatta/caronte).
 * Copyright (c) 2020 Emiliano Ciavatta.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, version 3.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func CreateApplicationRouter(applicationContext *ApplicationContext,
	notificationController *NotificationController, resourcesController *ResourcesController) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.MaxMultipartMemory = 8 << 30

	router.Use(static.Serve("/", static.LocalFile("./frontend/build", true)))

	for _, path := range []string{"/connections/:id", "/pcaps", "/rules", "/services", "/stats", "/searches"} {
		router.GET(path, func(c *gin.Context) {
			c.File("./frontend/build/index.html")
		})
	}

	router.POST("/setup", func(c *gin.Context) {
		if applicationContext.IsConfigured {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		var settings struct {
			Config   Config       `json:"config" binding:"required"`
			Accounts gin.Accounts `json:"accounts" binding:"required"`
		}

		if err := c.ShouldBindJSON(&settings); err != nil {
			badRequest(c, err)
			return
		}

		applicationContext.SetConfig(settings.Config)
		applicationContext.SetAccounts(settings.Accounts)

		c.JSON(http.StatusAccepted, gin.H{})
		notificationController.Notify("setup", gin.H{})
	})

	router.GET("/ws", func(c *gin.Context) {
		if err := notificationController.NotificationHandler(c.Writer, c.Request); err != nil {
			serverError(c, err)
		}
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
				response := UnorderedDocument{"id": id}
				success(c, response)
				notificationController.Notify("rules.new", response)
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

			isPresent, err := applicationContext.RulesManager.UpdateRule(c, id, rule)
			if err != nil {
				badRequest(c, err)
			} else if !isPresent {
				notFound(c, UnorderedDocument{"id": id})
			} else {
				success(c, rule)
				notificationController.Notify("rules.edit", rule)
			}
		})

		api.POST("/pcap/upload", func(c *gin.Context) {
			fileHeader, err := c.FormFile("file")
			if err != nil {
				badRequest(c, err)
				return
			}
			flushAllValue, isPresent := c.GetPostForm("flush_all")
			flushAll := isPresent && strings.ToLower(flushAllValue) == "true"
			fileName := fmt.Sprintf("%v-%s", time.Now().UnixNano(), fileHeader.Filename)
			if err := c.SaveUploadedFile(fileHeader, ProcessingPcapsBasePath+fileName); err != nil {
				log.WithError(err).Panic("failed to save uploaded file")
			}

			if sessionID, err := applicationContext.PcapImporter.ImportPcap(fileName, flushAll); err != nil {
				unprocessableEntity(c, err)
			} else {
				response := gin.H{"session": sessionID}
				c.JSON(http.StatusAccepted, response)
				notificationController.Notify("pcap.upload", response)
			}
		})

		api.POST("/pcap/file", func(c *gin.Context) {
			var request struct {
				File               string `json:"file"`
				FlushAll           bool   `json:"flush_all"`
				DeleteOriginalFile bool   `json:"delete_original_file"`
			}

			if err := c.ShouldBindJSON(&request); err != nil {
				badRequest(c, err)
				return
			}
			if !FileExists(request.File) {
				badRequest(c, errors.New("file not exists"))
				return
			}

			fileName := fmt.Sprintf("%v-%s", time.Now().UnixNano(), filepath.Base(request.File))
			if err := CopyFile(ProcessingPcapsBasePath+fileName, request.File); err != nil {
				log.WithError(err).Panic("failed to copy pcap file")
			}
			if sessionID, err := applicationContext.PcapImporter.ImportPcap(fileName, request.FlushAll); err != nil {
				if request.DeleteOriginalFile {
					if err := os.Remove(request.File); err != nil {
						log.WithError(err).Panic("failed to remove processed file")
					}
				}
				unprocessableEntity(c, err)
			} else {
				response := gin.H{"session": sessionID}
				c.JSON(http.StatusAccepted, response)
				notificationController.Notify("pcap.file", response)
			}
		})

		api.GET("/pcap/sessions", func(c *gin.Context) {
			success(c, applicationContext.PcapImporter.GetSessions())
		})

		api.GET("/pcap/sessions/:id", func(c *gin.Context) {
			sessionID := c.Param("id")
			if session, isPresent := applicationContext.PcapImporter.GetSession(sessionID); isPresent {
				success(c, session)
			} else {
				notFound(c, gin.H{"session": sessionID})
			}
		})

		api.GET("/pcap/sessions/:id/download", func(c *gin.Context) {
			sessionID := c.Param("id")
			if _, isPresent := applicationContext.PcapImporter.GetSession(sessionID); isPresent {
				if FileExists(PcapsBasePath + sessionID + ".pcap") {
					c.FileAttachment(PcapsBasePath+sessionID+".pcap", sessionID[:16]+".pcap")
				} else if FileExists(PcapsBasePath + sessionID + ".pcapng") {
					c.FileAttachment(PcapsBasePath+sessionID+".pcapng", sessionID[:16]+".pcapng")
				} else {
					log.WithField("sessionID", sessionID).Panic("pcap file not exists")
				}
			} else {
				notFound(c, gin.H{"session": sessionID})
			}
		})

		api.DELETE("/pcap/sessions/:id", func(c *gin.Context) {
			sessionID := c.Param("id")
			session := gin.H{"session": sessionID}
			if cancelled := applicationContext.PcapImporter.CancelSession(sessionID); cancelled {
				c.JSON(http.StatusAccepted, session)
				notificationController.Notify("sessions.delete", session)
			} else {
				notFound(c, session)
			}
		})

		api.GET("/connections", func(c *gin.Context) {
			var filter ConnectionsFilter
			if err := c.ShouldBindQuery(&filter); err != nil {
				badRequest(c, err)
				return
			}
			success(c, applicationContext.ConnectionsController.GetConnections(c, filter))
		})

		api.GET("/connections/:id", func(c *gin.Context) {
			if id, err := RowIDFromHex(c.Param("id")); err != nil {
				badRequest(c, err)
			} else {
				if connection, isPresent := applicationContext.ConnectionsController.GetConnection(c, id); isPresent {
					success(c, connection)
				} else {
					notFound(c, gin.H{"connection": id})
				}
			}
		})

		api.POST("/connections/:id/:action", func(c *gin.Context) {
			id, err := RowIDFromHex(c.Param("id"))
			if err != nil {
				badRequest(c, err)
				return
			}

			var result bool
			switch action := c.Param("action"); action {
			case "hide":
				result = applicationContext.ConnectionsController.SetHidden(c, id, true)
			case "show":
				result = applicationContext.ConnectionsController.SetHidden(c, id, false)
			case "mark":
				result = applicationContext.ConnectionsController.SetMarked(c, id, true)
			case "unmark":
				result = applicationContext.ConnectionsController.SetMarked(c, id, false)
			case "comment":
				var comment struct {
					Comment string `json:"comment"`
				}
				if err := c.ShouldBindJSON(&comment); err != nil {
					badRequest(c, err)
					return
				}
				result = applicationContext.ConnectionsController.SetComment(c, id, comment.Comment)
			default:
				badRequest(c, errors.New("invalid action"))
				return
			}

			if result {
				response := gin.H{"connection_id": c.Param("id"), "action": c.Param("action")}
				success(c, response)
				notificationController.Notify("connections.action", response)
			} else {
				notFound(c, gin.H{"connection": id})
			}
		})

		api.GET("/searches", func(c *gin.Context) {
			success(c, applicationContext.SearchController.GetPerformedSearches())
		})

		api.POST("/searches/perform", func(c *gin.Context) {
			var options SearchOptions

			if err := c.ShouldBindJSON(&options); err != nil {
				badRequest(c, err)
				return
			}

			// stupid checks because validator library is a shit
			var badContentError error
			if options.TextSearch.isZero() == options.RegexSearch.isZero() {
				badContentError = errors.New("specify either 'text_search' or 'regex_search'")
			}
			if !options.TextSearch.isZero() {
				if (options.TextSearch.Terms == nil) == (options.TextSearch.ExactPhrase == "") {
					badContentError = errors.New("specify either 'terms' or 'exact_phrase'")
				}
				if (options.TextSearch.Terms == nil) && (options.TextSearch.ExcludedTerms != nil) {
					badContentError = errors.New("'excluded_terms' must be specified only with 'terms'")
				}
			}
			if !options.RegexSearch.isZero() {
				if (options.RegexSearch.Pattern == "") == (options.RegexSearch.NotPattern == "") {
					badContentError = errors.New("specify either 'pattern' or 'not_pattern'")
				}
			}

			if badContentError != nil {
				badRequest(c, badContentError)
				return
			}

			success(c, applicationContext.SearchController.PerformSearch(c, options))
		})

		api.GET("/streams/:id", func(c *gin.Context) {
			id, err := RowIDFromHex(c.Param("id"))
			if err != nil {
				badRequest(c, err)
				return
			}
			var format GetMessageFormat
			if err := c.ShouldBindQuery(&format); err != nil {
				badRequest(c, err)
				return
			}

			if messages, found := applicationContext.ConnectionStreamsController.GetConnectionMessages(c, id, format); !found {
				notFound(c, gin.H{"connection": id})
			} else {
				success(c, messages)
			}
		})

		api.GET("/streams/:id/download", func(c *gin.Context) {
			id, err := RowIDFromHex(c.Param("id"))
			if err != nil {
				badRequest(c, err)
				return
			}
			var format DownloadMessageFormat
			if err := c.ShouldBindQuery(&format); err != nil {
				badRequest(c, err)
				return
			}

			if blob, found := applicationContext.ConnectionStreamsController.DownloadConnectionMessages(c, id, format); !found {
				notFound(c, gin.H{"connection": id})
			} else {
				c.String(http.StatusOK, blob)
			}
		})

		api.GET("/services", func(c *gin.Context) {
			success(c, applicationContext.ServicesController.GetServices())
		})

		api.PUT("/services", func(c *gin.Context) {
			var service Service
			if err := c.ShouldBindJSON(&service); err != nil {
				badRequest(c, err)
				return
			}
			if err := applicationContext.ServicesController.SetService(c, service); err == nil {
				success(c, service)
				notificationController.Notify("services.edit", service)
			} else {
				unprocessableEntity(c, err)
			}
		})

		api.DELETE("/services", func(c *gin.Context) {
			var service Service
			if err := c.ShouldBindJSON(&service); err != nil {
				badRequest(c, err)
				return
			}
			if err := applicationContext.ServicesController.DeleteService(c, service); err == nil {
				success(c, service)
				notificationController.Notify("services.edit", service)
			} else {
				unprocessableEntity(c, err)
			}
		})

		api.GET("/statistics", func(c *gin.Context) {
			var filter StatisticsFilter
			if err := c.ShouldBindQuery(&filter); err != nil {
				badRequest(c, err)
				return
			}

			success(c, applicationContext.StatisticsController.GetStatistics(c, filter))
		})

		api.GET("/statistics/totals", func(c *gin.Context) {
			var filter StatisticsFilter
			if err := c.ShouldBindQuery(&filter); err != nil {
				badRequest(c, err)
				return
			}

			success(c, applicationContext.StatisticsController.GetTotalStatistics(c, filter))
		})

		api.GET("/resources/system", func(c *gin.Context) {
			success(c, resourcesController.GetSystemStats(c))
		})

		api.GET("/resources/process", func(c *gin.Context) {
			success(c, resourcesController.GetProcessStats(c))
		})
	}

	return router
}

func SetupRequiredMiddleware(applicationContext *ApplicationContext) gin.HandlerFunc {
	return func(c *gin.Context) {
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

func serverError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, UnorderedDocument{"result": "error", "error": err.Error()})
}
