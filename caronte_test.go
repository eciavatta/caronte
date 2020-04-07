package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"testing"
	"time"
)

var storage Storage
var testContext context.Context

const testInsertManyFindCollection = "testFi"
const testCollection = "characters"

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

	dbName := fmt.Sprintf("%x", uniqueDatabaseName[:31])
	db := client.Database(dbName)
	log.Println("using database", dbName)
	mongoStorage := MongoStorage{
		client:      client,
		collections: map[string]*mongo.Collection{
			testInsertManyFindCollection: db.Collection(testInsertManyFindCollection),
			testCollection: db.Collection(testCollection),
		},
	}

	testContext, _ = context.WithTimeout(context.Background(), 10 * time.Second)

	err = mongoStorage.Connect(testContext)
	if err != nil {
		panic(err)
	}
	storage = &mongoStorage

	exitCode := m.Run()
	os.Exit(exitCode)
}
