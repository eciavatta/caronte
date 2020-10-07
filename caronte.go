package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

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

	versionBytes, err := ioutil.ReadFile("VERSION")
	if err != nil {
		log.WithError(err).Fatal("failed to load version file")
	}

	applicationContext, err := CreateApplicationContext(storage, string(versionBytes))
	if err != nil {
		log.WithError(err).WithFields(logFields).Fatal("failed to create application context")
	}

	notificationController := NewNotificationController(applicationContext)
	go notificationController.Run()
	applicationRouter := CreateApplicationRouter(applicationContext, notificationController)
	if applicationRouter.Run(fmt.Sprintf("%s:%v", *bindAddress, *bindPort)) != nil {
		log.WithError(err).WithFields(logFields).Fatal("failed to create the server")
	}
}
