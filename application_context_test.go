package main

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateApplicationContext(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	wrapper.AddCollection(Settings)

	appContext, err := CreateApplicationContext(wrapper.Storage)
	assert.NoError(t, err)
	assert.False(t, appContext.IsConfigured)
	assert.Zero(t, appContext.Config)
	assert.Len(t, appContext.Accounts, 0)
	assert.Nil(t, appContext.PcapImporter)
	assert.Nil(t, appContext.RulesManager)

	config := Config{
		ServerIP:     "10.10.10.10",
		FlagRegex:    "FLAG{test}",
		AuthRequired: true,
	}
	accounts := gin.Accounts{
		"username": "password",
	}
	appContext.SetConfig(config)
	appContext.SetAccounts(accounts)
	assert.Equal(t, appContext.Config, config)
	assert.Equal(t, appContext.Accounts, accounts)
	assert.NotNil(t, appContext.PcapImporter)
	assert.NotNil(t, appContext.RulesManager)
	assert.True(t, appContext.IsConfigured)

	config.FlagRegex = "FLAG{test2}"
	accounts["username"] = "password2"
	appContext.SetConfig(config)
	appContext.SetAccounts(accounts)

	checkAppContext, err := CreateApplicationContext(wrapper.Storage)
	assert.NoError(t, err)
	assert.True(t, checkAppContext.IsConfigured)
	assert.Equal(t, checkAppContext.Config, config)
	assert.Equal(t, checkAppContext.Accounts, accounts)
	assert.NotNil(t, checkAppContext.PcapImporter)
	assert.NotNil(t, checkAppContext.RulesManager)

	wrapper.Destroy(t)
}
