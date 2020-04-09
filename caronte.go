package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func main() {
	// test(); return
	mongoHost := flag.String("mongo-host", "localhost", "address of MongoDB")
	mongoPort := flag.Int("mongo-port", 27017, "port of MongoDB")
	dbName := flag.String("db-name", "caronte", "name of database to use")

	bindAddress := flag.String("bind-address", "0.0.0.0", "address where server is bind")
	bindPort := flag.Int("bind-port", 3333, "port where server is bind")

	flag.Parse()

	storage := NewMongoStorage(*mongoHost, *mongoPort, *dbName)
	err := storage.Connect(nil)
	if err != nil {
		log.WithError(err).Fatal("failed to connect to MongoDB")
	}

	router := gin.Default()
	ApplicationRoutes(router)
	err = router.Run(fmt.Sprintf("%s:%v", *bindAddress, *bindPort))
	if err != nil {
		log.WithError(err).Fatal("failed to create the server")
	}
}
