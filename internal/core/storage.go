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
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collections names
const (
	Connections       = "connections"
	ConnectionStreams = "connection_streams"
	ImportingSessions = "importing_sessions"
	Rules             = "rules"
	Searches          = "searches"
	Settings          = "settings"
	Services          = "services"
	Statistics        = "statistics"
)

var ZeroRowID [12]byte

type Storage interface {
	Insert(collectionName string) InsertOperation
	Update(collectionName string) UpdateOperation
	Find(collectionName string) FindOperation
	Delete(collectionName string) DeleteOperation
}

type MongoStorage struct {
	client      *mongo.Client
	collections map[string]*mongo.Collection
	database    *mongo.Database
}

type OrderedDocument = bson.D
type UnorderedDocument = bson.M
type Entry = bson.E
type RowID = primitive.ObjectID

func NewMongoStorage(uri string, port int, database, username, password string) (*MongoStorage, error) {
	ctx := context.Background()
	opt := options.Client()
	opt.ApplyURI(fmt.Sprintf("mongodb://%s:%v", uri, port))

	if username != "" && password != "" {
		opt.SetAuth(options.Credential{
			Username: username,
			Password: password,
		})
	}

	client, err := mongo.NewClient(opt)
	if err != nil {
		return nil, err
	}

	if err := client.Connect(ctx); err != nil {
		return nil, err
	}

	db := client.Database(database)
	collections := map[string]*mongo.Collection{
		Connections:       db.Collection(Connections),
		ConnectionStreams: db.Collection(ConnectionStreams),
		ImportingSessions: db.Collection(ImportingSessions),
		Rules:             db.Collection(Rules),
		Searches:          db.Collection(Searches),
		Settings:          db.Collection(Settings),
		Services:          db.Collection(Services),
		Statistics:        db.Collection(Statistics),
	}

	if _, err := collections[Services].Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return nil, err
	}

	if _, err := collections[ConnectionStreams].Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "connection_id", Value: -1}}, // descending
		},
		{
			Keys: bson.D{{Key: "payload_string", Value: "text"}},
		},
	}); err != nil {
		return nil, err
	}

	return &MongoStorage{
		client:      client,
		collections: collections,
		database:    db,
	}, nil
}

// InsertOne and InsertMany

type InsertOperation interface {
	Context(ctx context.Context) InsertOperation
	StopOnFail(stop bool) InsertOperation
	One(document any) (any, error)
	Many(documents []any) ([]any, error)
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

func (fo MongoInsertOperation) One(document any) (any, error) {
	if fo.err != nil {
		return nil, fo.err
	}

	result, err := fo.collection.InsertOne(fo.ctx, document)
	if err != nil {
		return nil, err
	}

	return result.InsertedID, nil
}

func (fo MongoInsertOperation) Many(documents []any) ([]any, error) {
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
	Upsert(upsertResults *any) UpdateOperation
	One(update any) (bool, error)
	OneComplex(update any) (bool, error)
	Many(update any) (int64, error)
}

type MongoUpdateOperation struct {
	collection   *mongo.Collection
	filter       OrderedDocument
	update       OrderedDocument
	ctx          context.Context
	opt          *options.UpdateOptions
	upsertResult *any
	err          error
}

func (fo MongoUpdateOperation) Context(ctx context.Context) UpdateOperation {
	fo.ctx = ctx
	return fo
}

func (fo MongoUpdateOperation) Filter(filter OrderedDocument) UpdateOperation {
	for _, elem := range filter {
		fo.filter = append(fo.filter, primitive.E{Key: elem.Key, Value: elem.Value})
	}
	return fo
}

func (fo MongoUpdateOperation) Upsert(upsertResults *any) UpdateOperation {
	fo.upsertResult = upsertResults
	fo.opt.SetUpsert(true)
	return fo
}

func (fo MongoUpdateOperation) One(update any) (bool, error) {
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

func (fo MongoUpdateOperation) OneComplex(update any) (bool, error) {
	if fo.err != nil {
		return false, fo.err
	}

	result, err := fo.collection.UpdateOne(fo.ctx, fo.filter, update, fo.opt)
	if err != nil {
		return false, err
	}

	if fo.upsertResult != nil {
		*(fo.upsertResult) = result.UpsertedID
	}
	return result.ModifiedCount == 1, nil
}

func (fo MongoUpdateOperation) Many(update any) (int64, error) {
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
		update:     OrderedDocument{{Key: "$set", Value: nil}},
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
	Projection(filter OrderedDocument) FindOperation
	Sort(field string, ascending bool) FindOperation
	Limit(n int64) FindOperation
	Skip(n int64) FindOperation
	MaxTime(duration time.Duration) FindOperation
	First(result any) error
	All(results any) error
	Traverse(constructor func() any, f func(any) bool) error
}

type MongoFindOperation struct {
	collection *mongo.Collection
	filter     OrderedDocument
	projection OrderedDocument
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
	for _, elem := range filter {
		fo.filter = append(fo.filter, primitive.E{Key: elem.Key, Value: elem.Value})
	}
	return fo
}

func (fo MongoFindOperation) Projection(projection OrderedDocument) FindOperation {
	for _, elem := range projection {
		fo.projection = append(fo.projection, primitive.E{Key: elem.Key, Value: elem.Value})
	}
	fo.optFindOne.SetProjection(fo.projection)
	fo.optFind.SetProjection(fo.projection)
	return fo
}

func (fo MongoFindOperation) Limit(n int64) FindOperation {
	fo.optFind.SetLimit(n)
	return fo
}

func (fo MongoFindOperation) Skip(n int64) FindOperation {
	fo.optFind.SetSkip(n)
	return fo
}

func (fo MongoFindOperation) MaxTime(duration time.Duration) FindOperation {
	fo.optFind.SetMaxTime(duration)
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

func (fo MongoFindOperation) First(result any) error {
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

func (fo MongoFindOperation) All(results any) error {
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

func (fo MongoFindOperation) Traverse(constructor func() any, f func(any) bool) error {
	if fo.err != nil {
		return fo.err
	}

	var limit int64
	if fo.optFind.Limit != nil {
		limit = *fo.optFind.Limit
		fo.optFind.SetLimit(0)
	}
	cursor, err := fo.collection.Find(fo.ctx, fo.filter, fo.optFind)
	if err != nil {
		return err
	}

	var counter int64
	for cursor.Next(fo.ctx) {
		obj := constructor()
		if err := cursor.Decode(obj); err != nil {
			return err
		}
		if matches := f(obj); matches {
			counter++
		}
		if limit > 0 && counter >= limit {
			break
		}
	}

	return cursor.Err()
}

func (storage *MongoStorage) Find(collectionName string) FindOperation {
	collection, ok := storage.collections[collectionName]
	op := MongoFindOperation{
		collection: collection,
		filter:     OrderedDocument{},
		projection: OrderedDocument{},
		optFind:    options.Find(),
		optFindOne: options.FindOne(),
		sorts:      OrderedDocument{},
	}
	if !ok {
		op.err = errors.New("invalid collection: " + collectionName)
	}
	return op
}

// Delete one/many

type DeleteOperation interface {
	Context(ctx context.Context) DeleteOperation
	Filter(filter OrderedDocument) DeleteOperation
	One() error
	Many() error
}

func (storage *MongoStorage) Delete(collectionName string) DeleteOperation {
	collection, ok := storage.collections[collectionName]
	op := MongoDeleteOperation{
		collection: collection,
		opts:       options.Delete(),
	}
	if !ok {
		op.err = errors.New("invalid collection: " + collectionName)
	}
	return op
}

type MongoDeleteOperation struct {
	collection *mongo.Collection
	ctx        context.Context
	opts       *options.DeleteOptions
	filter     OrderedDocument
	err        error
}

func (do MongoDeleteOperation) Context(ctx context.Context) DeleteOperation {
	do.ctx = ctx
	return do
}

func (do MongoDeleteOperation) Filter(filter OrderedDocument) DeleteOperation {
	for _, elem := range filter {
		do.filter = append(do.filter, primitive.E{Key: elem.Key, Value: elem.Value})
	}

	return do
}

func (do MongoDeleteOperation) One() error {
	if do.err != nil {
		return do.err
	}

	result, err := do.collection.DeleteOne(do.ctx, do.filter, do.opts)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("nothing to delete")
	}

	return nil
}

func (do MongoDeleteOperation) Many() error {
	if do.err != nil {
		return do.err
	}

	result, err := do.collection.DeleteMany(do.ctx, do.filter, do.opts)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("nothing to delete")
	}

	return nil
}
