package main

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// Collections names
const Connections = "connections"
const ConnectionStreams = "connection_streams"
const ImportedPcaps = "imported_pcaps"
const Rules = "rules"

const defaultConnectionTimeout = 10 * time.Second

var ZeroRowID [12]byte

type Storage interface {
	Insert(collectionName string) InsertOperation
	Update(collectionName string) UpdateOperation
	Find(collectionName string) FindOperation
	NewCustomRowID(payload uint64, timestamp time.Time) RowID
	NewRowID() RowID
}

type MongoStorage struct {
	client      *mongo.Client
	collections map[string]*mongo.Collection
}

type OrderedDocument = bson.D
type UnorderedDocument = bson.M
type Entry = bson.E
type RowID = primitive.ObjectID

func NewMongoStorage(uri string, port int, database string) *MongoStorage {
	opt := options.Client()
	opt.ApplyURI(fmt.Sprintf("mongodb://%s:%v", uri, port))
	client, err := mongo.NewClient(opt)
	if err != nil {
		log.WithError(err).Panic("failed to create mongo client")
	}

	db := client.Database(database)
	colls := map[string]*mongo.Collection{
		Connections:   db.Collection(Connections),
		ImportedPcaps: db.Collection(ImportedPcaps),
		Rules:         db.Collection(Rules),
	}

	return &MongoStorage{
		client:      client,
		collections: colls,
	}
}

func (storage *MongoStorage) Connect(ctx context.Context) error {
	return storage.client.Connect(ctx)
}

func (storage *MongoStorage) NewCustomRowID(payload uint64, timestamp time.Time) RowID {
	var key [12]byte
	binary.BigEndian.PutUint32(key[0:4], uint32(timestamp.Unix()))
	binary.BigEndian.PutUint64(key[4:12], payload)

	if oid, err := primitive.ObjectIDFromHex(hex.EncodeToString(key[:])); err == nil {
		return oid
	} else {
		log.WithError(err).Warn("failed to create object id")
		return primitive.NewObjectID()
	}
}

func (storage *MongoStorage) NewRowID() RowID {
	return primitive.NewObjectID()
}

// InsertOne and InsertMany

type InsertOperation interface {
	Context(ctx context.Context) InsertOperation
	StopOnFail(stop bool) InsertOperation
	One(document interface{}) (interface{}, error)
	Many(documents []interface{}) ([]interface{}, error)
}

type MongoInsertOperation struct {
	collection    *mongo.Collection
	ctx           context.Context
	optInsertMany *options.InsertManyOptions
	err           error
}

func (fo MongoInsertOperation) Context(ctx context.Context) InsertOperation {
	fo.ctx = ctx
	return fo
}

func (fo MongoInsertOperation) StopOnFail(stop bool) InsertOperation {
	fo.optInsertMany.SetOrdered(stop)
	return fo
}

func (fo MongoInsertOperation) One(document interface{}) (interface{}, error) {
	if fo.err != nil {
		return nil, fo.err
	}

	result, err := fo.collection.InsertOne(fo.ctx, document)
	if err != nil {
		return nil, err
	}

	return result.InsertedID, nil
}

func (fo MongoInsertOperation) Many(documents []interface{}) ([]interface{}, error) {
	if fo.err != nil {
		return nil, fo.err
	}

	results, err := fo.collection.InsertMany(fo.ctx, documents, fo.optInsertMany)
	if err != nil {
		return nil, err
	}

	return results.InsertedIDs, nil
}

func (storage *MongoStorage) Insert(collectionName string) InsertOperation {
	collection, ok := storage.collections[collectionName]
	op := MongoInsertOperation{
		collection:    collection,
		optInsertMany: options.InsertMany(),
	}
	if !ok {
		op.err = errors.New("invalid collection: " + collectionName)
	}
	return op
}

// UpdateOne and UpdateMany

type UpdateOperation interface {
	Context(ctx context.Context) UpdateOperation
	Filter(filter OrderedDocument) UpdateOperation
	Upsert(upsertResults *interface{}) UpdateOperation
	One(update interface{}) (bool, error)
	Many(update interface{}) (int64, error)
}

type MongoUpdateOperation struct {
	collection   *mongo.Collection
	filter       OrderedDocument
	update       OrderedDocument
	ctx          context.Context
	opt          *options.UpdateOptions
	upsertResult *interface{}
	err          error
}

func (fo MongoUpdateOperation) Context(ctx context.Context) UpdateOperation {
	fo.ctx = ctx
	return fo
}

func (fo MongoUpdateOperation) Filter(filter OrderedDocument) UpdateOperation {
	fo.filter = filter
	return fo
}

func (fo MongoUpdateOperation) Upsert(upsertResults *interface{}) UpdateOperation {
	fo.upsertResult = upsertResults
	fo.opt.SetUpsert(true)
	return fo
}

func (fo MongoUpdateOperation) One(update interface{}) (bool, error) {
	if fo.err != nil {
		return false, fo.err
	}

	for i := range fo.update {
		fo.update[i].Value = update
	}
	result, err := fo.collection.UpdateOne(fo.ctx, fo.filter, fo.update, fo.opt)
	if err != nil {
		return false, err
	}

	if fo.upsertResult != nil {
		*(fo.upsertResult) = result.UpsertedID
	}
	return result.ModifiedCount == 1, nil
}

func (fo MongoUpdateOperation) Many(update interface{}) (int64, error) {
	if fo.err != nil {
		return 0, fo.err
	}

	for i := range fo.update {
		fo.update[i].Value = update
	}
	result, err := fo.collection.UpdateMany(fo.ctx, fo.filter, fo.update, fo.opt)
	if err != nil {
		return 0, err
	}

	if fo.upsertResult != nil {
		*(fo.upsertResult) = result.UpsertedID
	}
	return result.ModifiedCount, nil
}

func (storage *MongoStorage) Update(collectionName string) UpdateOperation {
	collection, ok := storage.collections[collectionName]
	op := MongoUpdateOperation{
		collection: collection,
		filter:     OrderedDocument{},
		update:     OrderedDocument{{"$set", nil}},
		opt:        options.Update(),
	}
	if !ok {
		op.err = errors.New("invalid collection: " + collectionName)
	}
	return op
}

// Find and FindOne

type FindOperation interface {
	Context(ctx context.Context) FindOperation
	Filter(filter OrderedDocument) FindOperation
	Sort(field string, ascending bool) FindOperation
	Limit(n int64) FindOperation
	First(result interface{}) error
	All(results interface{}) error
}

type MongoFindOperation struct {
	collection *mongo.Collection
	filter     OrderedDocument
	ctx        context.Context
	optFind    *options.FindOptions
	optFindOne *options.FindOneOptions
	sorts      []Entry
	err        error
}

func (fo MongoFindOperation) Context(ctx context.Context) FindOperation {
	fo.ctx = ctx
	return fo
}

func (fo MongoFindOperation) Filter(filter OrderedDocument) FindOperation {
	fo.filter = filter
	return fo
}

func (fo MongoFindOperation) Limit(n int64) FindOperation {
	fo.optFind.SetLimit(n)
	return fo
}

func (fo MongoFindOperation) Sort(field string, ascending bool) FindOperation {
	var sort int
	if ascending {
		sort = 1
	} else {
		sort = -1
	}
	fo.sorts = append(fo.sorts, primitive.E{Key: field, Value: sort})
	fo.optFind.SetSort(fo.sorts)
	fo.optFindOne.SetSort(fo.sorts)
	return fo
}

func (fo MongoFindOperation) First(result interface{}) error {
	if fo.err != nil {
		return fo.err
	}

	err := fo.collection.FindOne(fo.ctx, fo.filter, fo.optFindOne).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			result = nil
			return nil
		}

		return err
	}
	return nil
}

func (fo MongoFindOperation) All(results interface{}) error {
	if fo.err != nil {
		return fo.err
	}
	cursor, err := fo.collection.Find(fo.ctx, fo.filter, fo.optFind)
	if err != nil {
		return err
	}
	err = cursor.All(fo.ctx, results)
	if err != nil {
		return err
	}
	return nil
}

func (storage *MongoStorage) Find(collectionName string) FindOperation {
	collection, ok := storage.collections[collectionName]
	op := MongoFindOperation{
		collection: collection,
		filter:     OrderedDocument{},
		optFind:    options.Find(),
		optFindOne: options.FindOne(),
		sorts:      OrderedDocument{},
	}
	if !ok {
		op.err = errors.New("invalid collection: " + collectionName)
	}
	return op
}
