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
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
)

var Version string

func main() {
	mongoHost := flag.String("mongo-host", "localhost", "address of MongoDB")
	mongoPort := flag.Int("mongo-port", 27017, "port of MongoDB")
	dbName := flag.String("db-name", "caronte", "name of database to use")

	bindAddress := flag.String("bind-address", "0.0.0.0", "address where server is bind")
	bindPort := flag.Int("bind-port", 3333, "port where server is bind")

	flag.Parse()

	logFields := log.Fields{"host": *mongoHost, "port": *mongoPort, "dbName": *dbName}
	storage, err := NewMongoStorage(*mongoHost, *mongoPort, *dbName)
	if err != nil {
		log.WithError(err).WithFields(logFields).Fatal("failed to connect to MongoDB")
	}

	if Version == "" {
		Version = "undefined"
	}
	applicationContext, err := CreateApplicationContext(storage, Version)
	if err != nil {
		log.WithError(err).WithFields(logFields).Fatal("failed to create application context")
	}

	notificationController := NewNotificationController(applicationContext)
	go notificationController.Run()
	applicationContext.SetNotificationController(notificationController)

	resourcesController := NewResourcesController(notificationController)
	go resourcesController.Run()

	applicationContext.Configure()
	applicationRouter := CreateApplicationRouter(applicationContext, notificationController, resourcesController)
	if applicationRouter.Run(fmt.Sprintf("%s:%v", *bindAddress, *bindPort)) != nil {
		log.WithError(err).WithFields(logFields).Fatal("failed to create the server")
	}
}
