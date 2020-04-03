package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const defaultConnectionTimeout = 10*time.Second
const defaultOperationTimeout = 3*time.Second

type Storage interface {
	InsertOne(ctx context.Context, collectionName string, document interface{}) (interface{}, error)
	UpdateOne(ctx context.Context, collectionName string, filter interface{}, update interface {}, upsert bool) (interface{}, error)
	FindOne(ctx context.Context, collectionName string, filter interface{}) (UnorderedDocument, error)
}

type MongoStorage struct {
	client *mongo.Client
	collections map[string]*mongo.Collection
}

type OrderedDocument = bson.D
type UnorderedDocument = bson.M

func NewMongoStorage(uri string, port int, database string) *MongoStorage {
	opt := options.Client()
	opt.ApplyURI(fmt.Sprintf("mongodb://%s:%v", uri, port))
	client, err := mongo.NewClient(opt)
	if err != nil {
		panic("Failed to create mongo client")
	}

	db := client.Database(database)
	colls := map[string]*mongo.Collection{
		"imported_pcaps": db.Collection("imported_pcaps"),
		"connections": db.Collection("connections"),
	}

	return &MongoStorage{
		client:      client,
		collections: colls,
	}
}

func (storage *MongoStorage) Connect(ctx context.Context) error {
	if ctx == nil {
		ctx, _ = context.WithTimeout(context.Background(), defaultConnectionTimeout)
	}

	return storage.client.Connect(ctx)
}

func (storage *MongoStorage) InsertOne(ctx context.Context, collectionName string,
	document interface{}) (interface{}, error) {

	collection, ok := storage.collections[collectionName]
	if !ok {
		return nil, errors.New("invalid collection: " + collectionName)
	}

	if ctx == nil {
		ctx, _ = context.WithTimeout(context.Background(), defaultOperationTimeout)
	}

	result, err := collection.InsertOne(ctx, document)
	if err != nil {
		return nil, err
	}

	return result.InsertedID, nil
}

func (storage *MongoStorage) UpdateOne(ctx context.Context, collectionName string,
	filter interface{}, update interface {}, upsert bool) (interface{}, error) {

	collection, ok := storage.collections[collectionName]
	if !ok {
		return nil, errors.New("invalid collection: " + collectionName)
	}

	if ctx == nil {
		ctx, _ = context.WithTimeout(context.Background(), defaultOperationTimeout)
	}

	opts := options.Update().SetUpsert(upsert)
	update = bson.D{{"$set", update}}

	result, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, err
	}

	if upsert {
		return result.UpsertedID, nil
	}

	return result.ModifiedCount == 1, nil
}

func (storage *MongoStorage) UpdateMany(ctx context.Context, collectionName string,
	filter interface{}, update interface {}, upsert bool) (interface{}, error) {

	collection, ok := storage.collections[collectionName]
	if !ok {
		return nil, errors.New("invalid collection: " + collectionName)
	}

	if ctx == nil {
		ctx, _ = context.WithTimeout(context.Background(), defaultOperationTimeout)
	}

	opts := options.Update().SetUpsert(upsert)
	update = bson.D{{"$set", update}}

	result, err := collection.UpdateMany(ctx, filter, update, opts)
	if err != nil {
		return nil, err
	}

	return result.ModifiedCount, nil
}

func (storage *MongoStorage) FindOne(ctx context.Context, collectionName string,
	filter interface{}) (UnorderedDocument, error) {

	collection, ok := storage.collections[collectionName]
	if !ok {
		return nil, errors.New("invalid collection: " + collectionName)
	}

	if ctx == nil {
		ctx, _ = context.WithTimeout(context.Background(), defaultOperationTimeout)
	}

	var result bson.M
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}

		return nil, err
	}

	return result, nil
}
