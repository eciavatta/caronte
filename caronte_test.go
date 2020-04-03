package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"testing"
	"time"
)

var storage Storage
var testContext context.Context

func TestMain(m *testing.M) {
	mongoHost, ok := os.LookupEnv("MONGO_HOST")
	if !ok {
		mongoHost = "localhost"
	}
	mongoPort, ok := os.LookupEnv("MONGO_PORT")
	if !ok {
		mongoPort = "27017"
	}

	uniqueDatabaseName := sha256.Sum256([]byte(time.Now().String()))

	client, err := mongo.NewClient(options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%v", mongoHost, mongoPort)))
	if err != nil {
		panic("failed to create mongo client")
	}

	db := client.Database(fmt.Sprintf("%x", uniqueDatabaseName[:31]))
	mongoStorage := MongoStorage{
		client:      client,
		collections: map[string]*mongo.Collection{testCollection: db.Collection(testCollection)},
	}

	testContext, _ = context.WithTimeout(context.Background(), 10 * time.Second)

	err = mongoStorage.Connect(nil)
	if err != nil {
		panic(err)
	}
	storage = &mongoStorage

	exitCode := m.Run()
	os.Exit(exitCode)
}
