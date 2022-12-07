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

package core

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	ServerAddress string `json:"server_address" binding:"omitempty,ip|cidr" bson:"server_address"`
	FlagRegex     string `json:"flag_regex" binding:"required,min=8" bson:"flag_regex"`
	AuthRequired  bool   `json:"auth_required" bson:"auth_required"`
}

type Status struct {
	PcapsAnalyzed int    `json:"pcaps_analyzed"`
	LiveCapture   string `json:"live_capture"`
	Version       string `json:"version"`
}

type ApplicationContext struct {
	Storage                     Storage
	Config                      Config
	Accounts                    gin.Accounts
	RulesManager                RulesManager
	PcapImporter                *PcapImporter
	ConnectionsController       ConnectionsController
	ServicesController          *ServicesController
	ConnectionStreamsController ConnectionStreamsController
	SearchController            *SearchController
	StatisticsController        StatisticsController
	NotificationController      NotificationController
	IsConfigured                bool
	Version                     string
}

func CreateApplicationContext(storage Storage, version string) (*ApplicationContext, error) {
	var configWrapper struct {
		Config Config
	}
	if err := storage.Find(Settings).Filter(OrderedDocument{{Key: "_id", Value: "config"}}).
		First(&configWrapper); err != nil {
		return nil, err
	}
	var accountsWrapper struct {
		Accounts gin.Accounts
	}

	if err := storage.Find(Settings).Filter(OrderedDocument{{Key: "_id", Value: "accounts"}}).
		First(&accountsWrapper); err != nil {
		return nil, err
	}
	if accountsWrapper.Accounts == nil {
		accountsWrapper.Accounts = make(gin.Accounts)
	}

	applicationContext := &ApplicationContext{
		Storage:  storage,
		Config:   configWrapper.Config,
		Accounts: accountsWrapper.Accounts,
		Version:  version,
	}

	return applicationContext, nil
}

func (sm *ApplicationContext) SetConfig(config Config) {
	sm.Config = config
	sm.Configure()
	var upsertResults any
	if _, err := sm.Storage.Update(Settings).Upsert(&upsertResults).
		Filter(OrderedDocument{{Key: "_id", Value: "config"}}).One(UnorderedDocument{"config": config}); err != nil {
		log.WithError(err).WithField("config", config).Error("failed to update config")
	}
}

func (sm *ApplicationContext) SetAccounts(accounts gin.Accounts) {
	sm.Accounts = accounts
	var upsertResults any
	if _, err := sm.Storage.Update(Settings).Upsert(&upsertResults).
		Filter(OrderedDocument{{Key: "_id", Value: "accounts"}}).One(UnorderedDocument{"accounts": accounts}); err != nil {
		log.WithError(err).Error("failed to update accounts")
	}
}

func (sm *ApplicationContext) SetNotificationController(notificationController NotificationController) {
	sm.NotificationController = notificationController
}

func (sm *ApplicationContext) Configure() {
	if sm.IsConfigured {
		return
	}
	if sm.Config.FlagRegex == "" {
		return
	}
	serverNet := ParseIPNet(sm.Config.ServerAddress)

	rulesManager, err := LoadRulesManager(sm.Storage, sm.Config.FlagRegex)
	if err != nil {
		log.WithError(err).Panic("failed to create a RulesManager")
	}
	sm.RulesManager = rulesManager
	sm.PcapImporter = NewPcapImporter(sm.Storage, serverNet, sm.RulesManager, sm.NotificationController)
	sm.ServicesController = NewServicesController(sm.Storage)
	sm.SearchController = NewSearchController(sm.Storage)
	sm.ConnectionsController = NewConnectionsController(sm.Storage, sm.SearchController, sm.ServicesController)
	sm.ConnectionStreamsController = NewConnectionStreamsController(sm.Storage)
	sm.StatisticsController = NewStatisticsController(sm.Storage)
	sm.IsConfigured = true
}

func (sm *ApplicationContext) Status() Status {
	return Status{
		PcapsAnalyzed: len(sm.PcapImporter.GetSessions()),
		LiveCapture:   sm.PcapImporter.GetLiveCaptureStatus(),
		Version:       sm.Version,
	}
}
