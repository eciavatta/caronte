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
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
	"time"
)

type a struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`
	A  string             `bson:"a,omitempty"`
	B  int                `bson:"b,omitempty"`
	C  time.Time          `bson:"c,omitempty"`
	D  map[string]b       `bson:"d"`
	E  []b                `bson:"e,omitempty"`
}

type b struct {
	A string `bson:"a,omitempty"`
	B int    `bson:"b,omitempty"`
}

func TestOperationOnInvalidCollection(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)

	simpleDoc := UnorderedDocument{"key": "a", "value": 0}
	insertOp := wrapper.Storage.Insert("invalid_collection").Context(wrapper.Context)
	insertedID, err := insertOp.One(simpleDoc)
	assert.Nil(t, insertedID)
	assert.Error(t, err)

	insertedIDs, err := insertOp.Many([]interface{}{simpleDoc})
	assert.Nil(t, insertedIDs)
	assert.Error(t, err)

	updateOp := wrapper.Storage.Update("invalid_collection").Context(wrapper.Context)
	isUpdated, err := updateOp.One(simpleDoc)
	assert.False(t, isUpdated)
	assert.Error(t, err)

	updated, err := updateOp.Many(simpleDoc)
	assert.Zero(t, updated)
	assert.Error(t, err)

	findOp := wrapper.Storage.Find("invalid_collection").Context(wrapper.Context)
	var result interface{}
	err = findOp.First(&result)
	assert.Nil(t, result)
	assert.Error(t, err)

	var results interface{}
	err = findOp.All(&result)
	assert.Nil(t, results)
	assert.Error(t, err)

	wrapper.Destroy(t)
}

func TestSimpleInsertAndFind(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	collectionName := "simple_insert_find"
	wrapper.AddCollection(collectionName)

	insertOp := wrapper.Storage.Insert(collectionName).Context(wrapper.Context)
	simpleDocA := UnorderedDocument{"key": "a"}
	idA, err := insertOp.One(simpleDocA)
	assert.Len(t, idA, 12)
	assert.Nil(t, err)

	simpleDocB := UnorderedDocument{"_id": "idb", "key": "b"}
	idB, err := insertOp.One(simpleDocB)
	assert.Equal(t, "idb", idB)
	assert.Nil(t, err)

	var result UnorderedDocument
	findOp := wrapper.Storage.Find(collectionName).Context(wrapper.Context)
	err = findOp.Filter(OrderedDocument{{"key", "a"}}).First(&result)
	assert.Nil(t, err)
	assert.Equal(t, idA, result["_id"])
	assert.Equal(t, simpleDocA["key"], result["key"])

	err = findOp.Filter(OrderedDocument{{"_id", idB}}).First(&result)
	assert.Nil(t, err)
	assert.Equal(t, idB, result["_id"])
	assert.Equal(t, simpleDocB["key"], result["key"])

	wrapper.Destroy(t)
}

func TestSimpleInsertManyAndFindMany(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	collectionName := "simple_insert_many_find_many"
	wrapper.AddCollection(collectionName)

	insertOp := wrapper.Storage.Insert(collectionName).Context(wrapper.Context)
	simpleDocs := []interface{}{
		UnorderedDocument{"key": "a"},
		UnorderedDocument{"_id": "idb", "key": "b"},
		UnorderedDocument{"key": "c"},
	}
	ids, err := insertOp.Many(simpleDocs)
	assert.Nil(t, err)
	assert.Len(t, ids, 3)
	assert.Equal(t, "idb", ids[1])

	var results []UnorderedDocument
	findOp := wrapper.Storage.Find(collectionName).Context(wrapper.Context)
	err = findOp.Sort("key", false).All(&results) // test sort ascending
	assert.Nil(t, err)
	assert.Len(t, results, 3)
	assert.Equal(t, "c", results[0]["key"])
	assert.Equal(t, "b", results[1]["key"])
	assert.Equal(t, "a", results[2]["key"])

	err = findOp.Sort("key", true).All(&results) // test sort descending
	assert.Nil(t, err)
	assert.Len(t, results, 3)
	assert.Equal(t, "c", results[2]["key"])
	assert.Equal(t, "b", results[1]["key"])
	assert.Equal(t, "a", results[0]["key"])

	err = findOp.Filter(OrderedDocument{{"key", OrderedDocument{{"$gte", "b"}}}}).
		Sort("key", true).All(&results) // test filter
	assert.Nil(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "b", results[0]["key"])
	assert.Equal(t, "c", results[1]["key"])

	wrapper.Destroy(t)
}

func TestSimpleUpdateOneUpdateMany(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	collectionName := "simple_update_one_update_many"
	wrapper.AddCollection(collectionName)

	insertOp := wrapper.Storage.Insert(collectionName).Context(wrapper.Context)
	simpleDocs := []interface{}{
		UnorderedDocument{"_id": "ida", "key": "a"},
		UnorderedDocument{"key": "b"},
		UnorderedDocument{"key": "c"},
	}
	_, err := insertOp.Many(simpleDocs)
	assert.Nil(t, err)

	updateOp := wrapper.Storage.Update(collectionName).Context(wrapper.Context)
	isUpdated, err := updateOp.Filter(OrderedDocument{{"_id", "ida"}}).
		One(OrderedDocument{{"key", "aa"}})
	assert.Nil(t, err)
	assert.True(t, isUpdated)

	updated, err := updateOp.Filter(OrderedDocument{{"key", OrderedDocument{{"$gte", "b"}}}}).
		Many(OrderedDocument{{"key", "bb"}})
	assert.Nil(t, err)
	assert.Equal(t, int64(2), updated)

	var upsertID interface{}
	isUpdated, err = updateOp.Upsert(&upsertID).Filter(OrderedDocument{{"key", "d"}}).
		One(OrderedDocument{{"key", "d"}})
	assert.Nil(t, err)
	assert.False(t, isUpdated)
	assert.NotNil(t, upsertID)

	var results []UnorderedDocument
	findOp := wrapper.Storage.Find(collectionName).Context(wrapper.Context)
	err = findOp.Sort("key", true).All(&results) // test sort ascending
	assert.Nil(t, err)
	assert.Len(t, results, 4)
	assert.Equal(t, "aa", results[0]["key"])
	assert.Equal(t, "bb", results[1]["key"])
	assert.Equal(t, "bb", results[2]["key"])
	assert.Equal(t, "d", results[3]["key"])

	wrapper.Destroy(t)
}

func TestComplexInsertManyFindMany(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	collectionName := "complex_insert_many_find_many"
	wrapper.AddCollection(collectionName)

	testTime := time.Now()
	oid1, err := primitive.ObjectIDFromHex("ffffffffffffffffffffffff")
	assert.Nil(t, err)

	docs := []interface{}{
		a{
			A: "test0",
			B: 0,
			C: testTime,
			D: map[string]b{
				"first":  {A: "0", B: 0},
				"second": {A: "1", B: 1},
			},
			E: []b{
				{A: "0", B: 0}, {A: "1", B: 0},
			},
		},
		a{
			ID: oid1,
			A:  "test1",
			B:  1,
			C:  testTime,
			D:  map[string]b{},
			E:  []b{},
		},
		a{},
	}

	ids, err := wrapper.Storage.Insert(collectionName).Context(wrapper.Context).Many(docs)
	assert.Nil(t, err)
	assert.Len(t, ids, 3)
	assert.Equal(t, ids[1], oid1)

	var results []a
	err = wrapper.Storage.Find(collectionName).Context(wrapper.Context).All(&results)
	assert.Nil(t, err)
	assert.Len(t, results, 3)
	doc0, doc1, doc2 := docs[0].(a), docs[1].(a), docs[2].(a)
	assert.Equal(t, ids[0], results[0].ID)
	assert.Equal(t, doc1.ID, results[1].ID)
	assert.Equal(t, ids[2], results[2].ID)
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

	wrapper.Destroy(t)
}
