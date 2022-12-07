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
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type TestStorageWrapper struct {
	DbName     string
	Storage    *MongoStorage
	Context    context.Context
	CancelFunc context.CancelFunc
}

func TestMain(m *testing.M) {
	if err := os.MkdirAll(ProcessingPcapsBasePath, 0755); err != nil {
		panic(err)
	}

	if err := os.MkdirAll(ConnectionPcapsBasePath, 0755); err != nil {
		panic(err)
	}

	code := m.Run()
	os.Exit(code)
}

func NewTestStorageWrapper(t *testing.T) *TestStorageWrapper {
	mongoHost, ok := os.LookupEnv("MONGO_HOST")
	if !ok {
		mongoHost = "localhost"
	}
	mongoPort, ok := os.LookupEnv("MONGO_PORT")
	if !ok {
		mongoPort = "27017"
	}
	port, err := strconv.Atoi(mongoPort)
	require.NoError(t, err, "invalid port")

	dbName := "caronte_test"

	storage, err := NewMongoStorage(mongoHost, port, dbName, "", "")
	require.NoError(t, err)
	ctx, cancelFunc := context.WithTimeout(context.Background(), 30*time.Second)

	require.NoError(t, storage.database.Drop(ctx))

	return &TestStorageWrapper{
		DbName:     dbName,
		Storage:    storage,
		Context:    ctx,
		CancelFunc: cancelFunc,
	}
}

func (tsw TestStorageWrapper) AddCollection(collectionName string) {
	tsw.Storage.collections[collectionName] = tsw.Storage.client.Database(tsw.DbName).Collection(collectionName)
}

func (tsw TestStorageWrapper) Destroy(t *testing.T) {
	err := tsw.Storage.client.Disconnect(tsw.Context)
	tsw.CancelFunc()
	require.NoError(t, err, "failed to disconnect to database")
}

func testEnvironmentHost() string {
	host, ok := os.LookupEnv("TEST_ENVIRONMENT_HOST")
	if !ok {
		host = "localhost"
	}

	return host
}

func testEnvironmentSshPort() uint16 {
	env, ok := os.LookupEnv("TEST_ENVIRONMENT_SSH_PORT")
	if !ok {
		env = "default"
	}
	if port, err := strconv.Atoi(env); err != nil {
		return 2222
	} else {
		return uint16(port)
	}
}

func testEnvironmentHttpPort() uint16 {
	env, ok := os.LookupEnv("TEST_ENVIRONMENT_SSH_PORT")
	if !ok {
		env = "default"
	}
	if http, err := strconv.Atoi(env); err != nil {
		return 8080
	} else {
		return uint16(http)
	}
}

func testContainerAddress(t *testing.T) string {
	cmd := exec.Command("docker", "inspect", "caronte-test-environment", "-f", "{{.NetworkSettings.IPAddress}}")
	output, err := cmd.Output()

	require.NoError(t, err)

	return strings.TrimSpace(string(output))
}
