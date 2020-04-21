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
)

func TestSetupApplication(t *testing.T) {
	toolkit := NewRouterTestToolkit(t, false)

	settings := make(map[string]interface{})
	assert.Equal(t, http.StatusServiceUnavailable, toolkit.MakeRequest("GET", "/api/rules", nil).Code)
	assert.Equal(t, http.StatusBadRequest, toolkit.MakeRequest("POST", "/setup", settings).Code)
	settings["config"] = Config{ServerIP: "1.2.3.4", FlagRegex: "FLAG{test}", AuthRequired: true}
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

type RouterTestToolkit struct {
	appContext *ApplicationContext
	wrapper    *TestStorageWrapper
	router     *gin.Engine
	t          *testing.T
}

func NewRouterTestToolkit(t *testing.T, withSetup bool) *RouterTestToolkit {
	wrapper := NewTestStorageWrapper(t)
	wrapper.AddCollection(Settings)

	appContext, err := CreateApplicationContext(wrapper.Storage)
	require.NoError(t, err)
	gin.SetMode(gin.ReleaseMode)
	router := CreateApplicationRouter(appContext)

	toolkit := RouterTestToolkit{
		appContext: appContext,
		wrapper:    wrapper,
		router:     router,
		t:          t,
	}

	if withSetup {
		settings := gin.H{
			"config":   Config{ServerIP: "1.2.3.4", FlagRegex: "FLAG{test}", AuthRequired: false},
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
