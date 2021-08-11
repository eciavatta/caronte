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
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateApplicationContext(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	wrapper.AddCollection(Settings)

	appContext, err := CreateApplicationContext(wrapper.Storage, "test")
	assert.NoError(t, err)
	assert.False(t, appContext.IsConfigured)
	assert.Zero(t, appContext.Config)
	assert.Len(t, appContext.Accounts, 0)
	assert.Nil(t, appContext.PcapImporter)
	assert.Nil(t, appContext.RulesManager)

	notificationController := NewNotificationController(appContext)
	appContext.SetNotificationController(notificationController)
	assert.Equal(t, notificationController, appContext.NotificationController)

	config := Config{
		ServerAddress: "10.10.10.10",
		FlagRegex:     "FLAG{test}",
		AuthRequired:  true,
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

	checkAppContext, err := CreateApplicationContext(wrapper.Storage, "test")
	assert.NoError(t, err)
	checkAppContext.SetNotificationController(notificationController)
	checkAppContext.Configure()
	assert.True(t, checkAppContext.IsConfigured)
	assert.Equal(t, checkAppContext.Config, config)
	assert.Equal(t, checkAppContext.Accounts, accounts)
	assert.NotNil(t, checkAppContext.PcapImporter)
	assert.NotNil(t, checkAppContext.RulesManager)
	assert.Equal(t, notificationController, appContext.NotificationController)

	wrapper.Destroy(t)
}
