package main

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net"
)

type Config struct {
	ServerIP     string `json:"server_ip" binding:"required,ip" bson:"server_ip"`
	FlagRegex    string `json:"flag_regex" binding:"required,min=8" bson:"flag_regex"`
	AuthRequired bool   `json:"auth_required" bson:"auth_required"`
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
	IsConfigured                bool
}

func CreateApplicationContext(storage Storage) (*ApplicationContext, error) {
	var configWrapper struct {
		Config Config
	}
	if err := storage.Find(Settings).Filter(OrderedDocument{{"_id", "config"}}).
		First(&configWrapper); err != nil {
		return nil, err
	}
	var accountsWrapper struct {
		Accounts gin.Accounts
	}

	if err := storage.Find(Settings).Filter(OrderedDocument{{"_id", "accounts"}}).
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
	}

	applicationContext.configure()
	return applicationContext, nil
}

func (sm *ApplicationContext) SetConfig(config Config) {
	sm.Config = config
	sm.configure()
	var upsertResults interface{}
	if _, err := sm.Storage.Update(Settings).Upsert(&upsertResults).
		Filter(OrderedDocument{{"_id", "config"}}).One(UnorderedDocument{"config": config}); err != nil {
		log.WithError(err).WithField("config", config).Error("failed to update config")
	}
}

func (sm *ApplicationContext) SetAccounts(accounts gin.Accounts) {
	sm.Accounts = accounts
	var upsertResults interface{}
	if _, err := sm.Storage.Update(Settings).Upsert(&upsertResults).
		Filter(OrderedDocument{{"_id", "accounts"}}).One(UnorderedDocument{"accounts": accounts}); err != nil {
		log.WithError(err).Error("failed to update accounts")
	}
}

func (sm *ApplicationContext) configure() {
	if sm.IsConfigured {
		return
	}
	if sm.Config.ServerIP == "" || sm.Config.FlagRegex == "" {
		return
	}
	serverIP := net.ParseIP(sm.Config.ServerIP)
	if serverIP == nil {
		return
	}

	rulesManager, err := LoadRulesManager(sm.Storage, sm.Config.FlagRegex)
	if err != nil {
		log.WithError(err).Panic("failed to create a RulesManager")
	}
	sm.RulesManager = rulesManager
	sm.PcapImporter = NewPcapImporter(sm.Storage, serverIP, sm.RulesManager)
	sm.ServicesController = NewServicesController(sm.Storage)
	sm.ConnectionsController = NewConnectionsController(sm.Storage, sm.ServicesController)
	sm.ConnectionStreamsController = NewConnectionStreamsController(sm.Storage)
	sm.IsConfigured = true
}
