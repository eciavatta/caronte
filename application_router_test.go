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
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSetupApplication(t *testing.T) {
	toolkit := NewRouterTestToolkit(t, false)

	settings := make(map[string]interface{})
	assert.Equal(t, http.StatusServiceUnavailable, toolkit.MakeRequest("GET", "/api/rules", nil).Code)
	assert.Equal(t, http.StatusBadRequest, toolkit.MakeRequest("POST", "/setup", settings).Code)
	settings["config"] = Config{ServerAddress: "1.2.3.4", FlagRegex: "FLAG{test}", AuthRequired: true}
	assert.Equal(t, http.StatusBadRequest, toolkit.MakeRequest("POST", "/setup", settings).Code)
	settings["accounts"] = gin.Accounts{"username": "password"}
	assert.Equal(t, http.StatusAccepted, toolkit.MakeRequest("POST", "/setup", settings).Code)
	assert.Equal(t, http.StatusNotFound, toolkit.MakeRequest("POST", "/setup", settings).Code)

	toolkit.wrapper.Destroy(t)
}

func TestAuthRequired(t *testing.T) {
	toolkit := NewRouterTestToolkit(t, true)

	assert.Equal(t, http.StatusOK, toolkit.MakeRequest("GET", "/api/rules", nil).Code)
	config := toolkit.appContext.Config
	config.AuthRequired = true
	toolkit.appContext.SetConfig(config)
	toolkit.appContext.SetAccounts(gin.Accounts{"username": "password"})
	assert.Equal(t, http.StatusUnauthorized, toolkit.MakeRequest("GET", "/api/rules", nil).Code)

	toolkit.wrapper.Destroy(t)
}

func TestRulesApi(t *testing.T) {
	toolkit := NewRouterTestToolkit(t, true)

	// AddRule
	assert.Equal(t, http.StatusBadRequest, toolkit.MakeRequest("POST", "/api/rules", Rule{}).Code)
	assert.Equal(t, http.StatusBadRequest, toolkit.MakeRequest("POST", "/api/rules",
		Rule{Name: "testRule"}).Code)
	assert.Equal(t, http.StatusBadRequest, toolkit.MakeRequest("POST", "/api/rules",
		Rule{Name: "testRule", Color: "invalidColor"}).Code)
	w := toolkit.MakeRequest("POST", "/api/rules", Rule{Name: "testRule", Color: "#fff"})
	var testRuleID struct{ ID string }
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &testRuleID))
	assert.Equal(t, http.StatusUnprocessableEntity, toolkit.MakeRequest("POST", "/api/rules",
		Rule{Name: "testRule", Color: "#fff"}).Code) // same name

	// UpdateRule
	assert.Equal(t, http.StatusBadRequest, toolkit.MakeRequest("PUT", "/api/rules/invalidID",
		Rule{Name: "invalidRule", Color: "#000"}).Code)
	assert.Equal(t, http.StatusNotFound, toolkit.MakeRequest("PUT", "/api/rules/000000000000000000000000",
		Rule{Name: "invalidRule", Color: "#000"}).Code)
	assert.Equal(t, http.StatusBadRequest, toolkit.MakeRequest("PUT", "/api/rules/"+testRuleID.ID, Rule{}).Code)
	assert.Equal(t, http.StatusBadRequest, toolkit.MakeRequest("PUT", "/api/rules/"+testRuleID.ID,
		Rule{Name: "invalidRule", Color: "invalidColor"}).Code)
	w = toolkit.MakeRequest("POST", "/api/rules", Rule{Name: "testRule2", Color: "#eee"})
	var testRule2ID struct{ ID string }
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &testRule2ID))
	assert.Equal(t, http.StatusBadRequest, toolkit.MakeRequest("PUT", "/api/rules/"+testRule2ID.ID,
		Rule{Name: "testRule", Color: "#fff"}).Code) // duplicate
	w = toolkit.MakeRequest("PUT", "/api/rules/"+testRuleID.ID, Rule{Name: "newRule1", Color: "#ddd"})
	var testRule Rule
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &testRule))
	assert.Equal(t, "newRule1", testRule.Name)
	assert.Equal(t, "#ddd", testRule.Color)

	// GetRule
	assert.Equal(t, http.StatusBadRequest, toolkit.MakeRequest("GET", "/api/rules/invalidID", nil).Code)
	assert.Equal(t, http.StatusNotFound, toolkit.MakeRequest("GET", "/api/rules/000000000000000000000000", nil).Code)
	w = toolkit.MakeRequest("GET", "/api/rules/"+testRuleID.ID, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &testRule))
	assert.Equal(t, testRuleID.ID, testRule.ID.Hex())
	assert.Equal(t, "newRule1", testRule.Name)
	assert.Equal(t, "#ddd", testRule.Color)

	// GetRules
	w = toolkit.MakeRequest("GET", "/api/rules", nil)
	var rules []Rule
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &rules))
	assert.Len(t, rules, 4)

	toolkit.wrapper.Destroy(t)
}

func TestPcapImporterApi(t *testing.T) {
	toolkit := NewRouterTestToolkit(t, true)

	// Import pcap
	assert.Equal(t, http.StatusBadRequest, toolkit.MakeRequest("POST", "/api/pcap/file", nil).Code)
	assert.Equal(t, http.StatusBadRequest, toolkit.MakeRequest("POST", "/api/pcap/file",
		gin.H{"file": "invalidPath"}).Code)
	w := toolkit.MakeRequest("POST", "/api/pcap/file", gin.H{"file": "test_data/ping_pong_10000.pcap"})
	var sessionID struct{ Session string }
	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &sessionID))
	assert.Equal(t, "369ef4b6abb6214b4ee2e0c81ecb93c49e275c26c85e30493b37727d408cf280", sessionID.Session)
	assert.Equal(t, http.StatusUnprocessableEntity, toolkit.MakeRequest("POST", "/api/pcap/file",
		gin.H{"file": "test_data/ping_pong_10000.pcap"}).Code) // duplicate

	// Get sessions
	var sessions []ImportingSession
	w = toolkit.MakeRequest("GET", "/api/pcap/sessions", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &sessions))
	assert.Len(t, sessions, 1)
	assert.Equal(t, sessionID.Session, sessions[0].ID)

	// Get session
	var session ImportingSession
	w = toolkit.MakeRequest("GET", "/api/pcap/sessions/"+sessionID.Session, nil)
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &session))
	assert.Equal(t, sessionID.Session, session.ID)

	// Cancel session
	assert.Equal(t, http.StatusNotFound, toolkit.MakeRequest("DELETE", "/api/pcap/sessions/invalidSession",
		nil).Code)
	assert.Equal(t, http.StatusAccepted, toolkit.MakeRequest("DELETE", "/api/pcap/sessions/"+sessionID.Session,
		nil).Code)

	time.Sleep(1 * time.Second) // wait for termination

	toolkit.wrapper.Destroy(t)
}

type RouterTestToolkit struct {
	appContext *ApplicationContext
	wrapper    *TestStorageWrapper
	router     *gin.Engine
	t          *testing.T
}

func NewRouterTestToolkit(t *testing.T, withSetup bool) *RouterTestToolkit {
	wrapper := NewTestStorageWrapper(t)
	wrapper.AddCollection(Settings)

	appContext, err := CreateApplicationContext(wrapper.Storage, "test")
	require.NoError(t, err)
	gin.SetMode(gin.ReleaseMode)
	notificationController := NewNotificationController(appContext)
	go notificationController.Run()
	resourcesController := NewResourcesController(notificationController)
	router := CreateApplicationRouter(appContext, notificationController, resourcesController)

	toolkit := RouterTestToolkit{
		appContext: appContext,
		wrapper:    wrapper,
		router:     router,
		t:          t,
	}

	if withSetup {
		settings := gin.H{
			"config":   Config{ServerAddress: "1.2.3.4", FlagRegex: "FLAG{test}", AuthRequired: false},
			"accounts": gin.Accounts{},
		}
		toolkit.MakeRequest("POST", "/setup", settings)
	}

	return &toolkit
}

func (rtt *RouterTestToolkit) MakeRequest(method string, url string, body interface{}) *httptest.ResponseRecorder {
	var r io.Reader

	if body != nil {
		buf, err := json.Marshal(body)
		require.NoError(rtt.t, err)
		r = bytes.NewBuffer(buf)
	}

	w := httptest.NewRecorder()
	req, err := http.NewRequest(method, url, r)
	require.NoError(rtt.t, err)
	rtt.router.ServeHTTP(w, req)

	return w
}
