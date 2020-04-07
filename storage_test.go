package main

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
	"time"
)

type a struct {
	Id primitive.ObjectID `bson:"_id,omitempty"`
	A string `bson:"a,omitempty"`
	B int `bson:"b,omitempty"`
	C time.Time `bson:"c,omitempty"`
	D map[string]b `bson:"d"`
	E []b `bson:"e,omitempty"`
}

type b struct {
	A string `bson:"a,omitempty"`
	B int `bson:"b,omitempty"`
}

func testInsert(t *testing.T) {
	// insert a document in an invalid connection
	insertedId, err := storage.InsertOne(testContext, "invalid_collection",
		OrderedDocument{{"key", "invalid"}})
	if insertedId != nil || err == nil {
		t.Fatal("inserting documents in invalid collections must fail")
	}

	// insert ordered document
	beatriceId, err := storage.InsertOne(testContext, testCollection,
		OrderedDocument{{"name", "Beatrice"}, {"description", "blablabla"}})
	if err != nil {
		t.Fatal(err)
	}
	if beatriceId == nil {
		t.Fatal("failed to insert an ordered document")
	}

	// insert unordered document
	virgilioId, err := storage.InsertOne(testContext, testCollection,
		UnorderedDocument{"name": "Virgilio", "description": "blablabla"})
	if err != nil {
		t.Fatal(err)
	}
	if virgilioId == nil {
		t.Fatal("failed to insert an unordered document")
	}

	// insert document with custom id
	danteId := "000000"
	insertedId, err = storage.InsertOne(testContext, testCollection,
		UnorderedDocument{"_id": danteId, "name": "Dante Alighieri", "description": "blablabla"})
	if err != nil {
		t.Fatal(err)
	}
	if insertedId != danteId {
		t.Fatal("returned id doesn't match")
	}

	// insert duplicate document
	insertedId, err = storage.InsertOne(testContext, testCollection,
		UnorderedDocument{"_id": danteId, "name": "Dante Alighieri", "description": "blablabla"})
	if insertedId != nil || err == nil {
		t.Fatal("inserting duplicate id must fail")
	}
}

func testFindOne(t *testing.T) {
	// find a document in an invalid connection
	result, err := storage.FindOne(testContext, "invalid_collection",
		OrderedDocument{{"key", "invalid"}})
	if result != nil || err == nil {
		t.Fatal("find a document in an invalid collections must fail")
	}

	// find an existing document
	result, err = storage.FindOne(testContext, testCollection, OrderedDocument{{"_id", "000000"}})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("FindOne cannot find the valid document")
	}
	name, ok := result["name"]
	if !ok || name != "Dante Alighieri" {
		t.Fatal("document retrieved with FindOne is invalid")
	}

	// find an existing document
	result, err = storage.FindOne(testContext, testCollection, OrderedDocument{{"_id", "invalid_id"}})
	if err != nil {
		t.Fatal(err)
	}
	if result != nil {
		t.Fatal("FindOne cannot find an invalid document")
	}
}

func TestBasicOperations(t *testing.T) {
	t.Run("testInsert", testInsert)
	t.Run("testFindOne", testFindOne)
}

func TestInsertManyFindDocuments(t *testing.T) {
	testTime := time.Now()
	oid1, err := primitive.ObjectIDFromHex("ffffffffffffffffffffffff")
	assert.Nil(t, err)

	docs := []interface{}{
		a{
			A: "test0",
			B: 0,
			C: testTime,
			D: map[string]b{
				"first": {A: "0", B: 0},
				"second": {A: "1", B: 1},
			},
			E: []b{
				{A: "0", B: 0}, {A: "1", B: 0},
			},
		},
		a{
			Id: oid1,
			A: "test1",
			B: 1,
			C: testTime,
			D: map[string]b{},
			E: []b{},
		},
		a{},
	}

	ids, err := storage.InsertMany(testContext, testInsertManyFindCollection, docs)
	assert.Nil(t, err)
	assert.Len(t, ids, 3)
	assert.Equal(t, ids[1], oid1)

	var results []a
	err = storage.Find(testContext, testInsertManyFindCollection, NoFilters, &results)
	assert.Nil(t, err)
	assert.Len(t, results, 3)
	doc0, doc1, doc2 := docs[0].(a), docs[1].(a), docs[2].(a)
	assert.Equal(t, ids[0], results[0].Id)
	assert.Equal(t, doc1.Id, results[1].Id)
	assert.Equal(t, ids[2], results[2].Id)
	assert.Equal(t, doc0.A, results[0].A)
	assert.Equal(t, doc1.A, results[1].A)
	assert.Equal(t, doc2.A, results[2].A)
	assert.Equal(t, doc0.B, results[0].B)
	assert.Equal(t, doc1.B, results[1].B)
	assert.Equal(t, doc2.B, results[2].B)
	assert.Equal(t, doc0.C.Unix(), results[0].C.Unix())
	assert.Equal(t, doc1.C.Unix(), results[1].C.Unix())
	assert.Equal(t, doc2.C.Unix(), results[2].C.Unix())
	assert.Equal(t, doc0.D, results[0].D)
	assert.Equal(t, doc1.D, results[1].D)
	assert.Equal(t, doc2.D, results[2].D)
	assert.Equal(t, doc0.E, results[0].E)
	assert.Nil(t, results[1].E)
	assert.Nil(t, results[2].E)
}

type testStorage struct {
	insertFunc func(ctx context.Context, collectionName string, document interface{}) (interface{}, error)
	insertManyFunc func(ctx context.Context, collectionName string, document []interface{}) ([]interface{}, error)
	updateOne func(ctx context.Context, collectionName string, filter interface{}, update interface {}, upsert bool) (interface{}, error)
	updateMany func(ctx context.Context, collectionName string, filter interface{}, update interface {}, upsert bool) (interface{}, error)
	findOne func(ctx context.Context, collectionName string, filter interface{}) (UnorderedDocument, error)
	find func(ctx context.Context, collectionName string, filter interface{}, results interface{}) error
}

func (ts testStorage) InsertOne(ctx context.Context, collectionName string, document interface{}) (interface{}, error) {
	if ts.insertFunc != nil {
		return ts.insertFunc(ctx, collectionName, document)
	}
	return nil, errors.New("not implemented")
}

func (ts testStorage) InsertMany(ctx context.Context, collectionName string, document []interface{}) ([]interface{}, error) {
	if ts.insertFunc != nil {
		return ts.insertManyFunc(ctx, collectionName, document)
	}
	return nil, errors.New("not implemented")
}

func (ts testStorage) UpdateOne(ctx context.Context, collectionName string, filter interface{}, update interface {},
	upsert bool) (interface{}, error) {
	if ts.updateOne != nil {
		return ts.updateOne(ctx, collectionName, filter, update, upsert)
	}
	return nil, errors.New("not implemented")
}

func (ts testStorage) UpdateMany(ctx context.Context, collectionName string, filter interface{}, update interface {},
	upsert bool) (interface{}, error) {
	if ts.updateOne != nil {
		return ts.updateMany(ctx, collectionName, filter, update, upsert)
	}
	return nil, errors.New("not implemented")
}

func (ts testStorage) FindOne(ctx context.Context, collectionName string, filter interface{}) (UnorderedDocument, error) {
	if ts.findOne != nil {
		return ts.findOne(ctx, collectionName, filter)
	}
	return nil, errors.New("not implemented")
}

func (ts testStorage) Find(ctx context.Context, collectionName string, filter interface{}, results interface{}) error {
	if ts.find != nil {
		return ts.find(ctx, collectionName, filter, results)
	}
	return errors.New("not implemented")
}
