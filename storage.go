package main

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const Connections = "connections"
const ImportedPcaps = "imported_pcaps"
const Rules = "rules"

var NoFilters = UnorderedDocument{}

const defaultConnectionTimeout = 10*time.Second
const defaultOperationTimeout = 3*time.Second

type Storage interface {
	InsertOne(ctx context.Context, collectionName string, document interface{}) (interface{}, error)
	InsertMany(ctx context.Context, collectionName string, documents []interface{}) ([]interface{}, error)
	UpdateOne(ctx context.Context, collectionName string, filter interface{}, update interface {}, upsert bool) (interface{}, error)
	UpdateMany(ctx context.Context, collectionName string, filter interface{}, update interface {}, upsert bool) (interface{}, error)
	FindOne(ctx context.Context, collectionName string, filter interface{}) (UnorderedDocument, error)
	Find(ctx context.Context, collectionName string, filter interface{}, results interface{}) error
}

type MongoStorage struct {
	client *mongo.Client
	collections map[string]*mongo.Collection
}

type OrderedDocument = bson.D
type UnorderedDocument = bson.M

func UniqueKey(timestamp time.Time, payload uint32) string {
	var key [8]byte
	binary.BigEndian.PutUint32(key[0:4], uint32(timestamp.Unix()))
	binary.BigEndian.PutUint32(key[4:8], payload)

	return hex.EncodeToString(key[:])
}

func NewMongoStorage(uri string, port int, database string) *MongoStorage {
	opt := options.Client()
	opt.ApplyURI(fmt.Sprintf("mongodb://%s:%v", uri, port))
	client, err := mongo.NewClient(opt)
	if err != nil {
		panic("Failed to create mongo client")
	}

	db := client.Database(database)
	colls := map[string]*mongo.Collection{
		Connections: db.Collection(Connections),
		ImportedPcaps: db.Collection(ImportedPcaps),
		Rules: db.Collection(Rules),
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

func (storage *MongoStorage) InsertMany(ctx context.Context, collectionName string,
	documents []interface{}) ([]interface{}, error) {

	collection, ok := storage.collections[collectionName]
	if !ok {
		return nil, errors.New("invalid collection: " + collectionName)
	}

	if ctx == nil {
		ctx, _ = context.WithTimeout(context.Background(), defaultOperationTimeout)
	}

	result, err := collection.InsertMany(ctx, documents)
	if err != nil {
		return nil, err
	}

	return result.InsertedIDs, nil
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

type FindOperation struct {
	options options.FindOptions
}



func (storage *MongoStorage) Find(ctx context.Context, collectionName string,
	filter interface{}, results interface{}) error {

	collection, ok := storage.collections[collectionName]
	if !ok {
		return errors.New("invalid collection: " + collectionName)
	}

	if ctx == nil {
		ctx, _ = context.WithTimeout(context.Background(), defaultOperationTimeout)
	}

	options.FindOptions{
		AllowDiskUse:        nil,
		AllowPartialResults: nil,
		BatchSize:           nil,
		Collation:           nil,
		Comment:             nil,
		CursorType:          nil,
		Hint:                nil,
		Limit:               nil,
		Max:                 nil,
		MaxAwaitTime:        nil,
		MaxTime:             nil,
		Min:                 nil,
		NoCursorTimeout:     nil,
		OplogReplay:         nil,
		Projection:          nil,
		ReturnKey:           nil,
		ShowRecordID:        nil,
		Skip:                nil,
		Snapshot:            nil,
		Sort:                nil,
	}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return err
	}
	err = cursor.All(ctx, results)
	if err != nil {
		return err
	}

	return nil
}
