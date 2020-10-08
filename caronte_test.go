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
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"strconv"
	"testing"
	"time"
)

type TestStorageWrapper struct {
	DbName  string
	Storage *MongoStorage
	Context context.Context
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

	dbName := fmt.Sprintf("%x", time.Now().UnixNano())
	log.WithField("dbName", dbName).Info("creating new storage")

	storage, err := NewMongoStorage(mongoHost, port, dbName)
	require.NoError(t, err)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	return &TestStorageWrapper{
		DbName:  dbName,
		Storage: storage,
		Context: ctx,
	}
}

func (tsw TestStorageWrapper) AddCollection(collectionName string) {
	tsw.Storage.collections[collectionName] = tsw.Storage.client.Database(tsw.DbName).Collection(collectionName)
}

func (tsw TestStorageWrapper) Destroy(t *testing.T) {
	err := tsw.Storage.client.Disconnect(tsw.Context)
	require.NoError(t, err, "failed to disconnect to database")
}
