package main

import "time"

type ConnectionStream struct {
	ID               RowID                   `json:"id" bson:"_id"`
	ConnectionID     RowID                   `json:"connection_id" bson:"connection_id"`
	DocumentIndex    int                     `json:"document_index" bson:"document_index"`
	Payload          []byte                  `json:"payload" bson:"payload"`
	BlocksIndexes    []int                   `json:"blocks_indexes" bson:"blocks_indexes"`
	BlocksTimestamps []time.Time             `json:"blocks_timestamps" bson:"blocks_timestamps"`
	BlocksLoss       []bool                  `json:"blocks_loss" bson:"blocks_loss"`
	PatternMatches   map[uint][]PatternSlice `json:"pattern_matches" bson:"pattern_matches"`
}

type PatternSlice [2]uint64
