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
	"context"
	"fmt"
	"github.com/eciavatta/caronte/similarity"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	DefaultQueryLimit               = 50
	MaxQueryLimit                   = 200
	SimilarToBoth                   = 0
	SimilarToClient                 = 1
	SimilarToServer                 = 2
	MinSimilaritySizeRatioThreshold = 0.9
)

type Connection struct {
	ID                      RowID     `json:"id" bson:"_id"`
	SourceIP                string    `json:"ip_src" bson:"ip_src"`
	DestinationIP           string    `json:"ip_dst" bson:"ip_dst"`
	SourcePort              uint16    `json:"port_src" bson:"port_src"`
	DestinationPort         uint16    `json:"port_dst" bson:"port_dst"`
	StartedAt               time.Time `json:"started_at" bson:"started_at"`
	ClosedAt                time.Time `json:"closed_at" bson:"closed_at"`
	ClientBytes             int       `json:"client_bytes" bson:"client_bytes"`
	ServerBytes             int       `json:"server_bytes" bson:"server_bytes"`
	ClientDocuments         int       `json:"client_documents" bson:"client_documents"`
	ServerDocuments         int       `json:"server_documents" bson:"server_documents"`
	ProcessedAt             time.Time `json:"processed_at" bson:"processed_at"`
	MatchedRules            []RowID   `json:"matched_rules" bson:"matched_rules"`
	Hidden                  bool      `json:"hidden" bson:"hidden,omitempty"`
	Marked                  bool      `json:"marked" bson:"marked,omitempty"`
	Comment                 string    `json:"comment" bson:"comment,omitempty"`
	Service                 Service   `json:"service" bson:"-"`
	ClientTlshHash          string    `json:"client_tlsh_hash" bson:"client_tlsh_hash"`
	ClientByteHistogramHash []byte    `json:"client_byte_histogram_hash" bson:"client_byte_histogram_hash"`
	ServerTlshHash          string    `json:"server_tlsh_hash" bson:"server_tlsh_hash"`
	ServerByteHistogramHash []byte    `json:"server_byte_histogram_hash" bson:"server_byte_histogram_hash"`
}

type ConnectionsFilter struct {
	From              string   `form:"from" binding:"omitempty,hexadecimal,len=24"`
	To                string   `form:"to" binding:"omitempty,hexadecimal,len=24"`
	ServicePort       uint16   `form:"service_port"`
	ClientAddress     string   `form:"client_address" binding:"omitempty,ip"`
	ClientPort        uint16   `form:"client_port"`
	MinDuration       uint     `form:"min_duration"`
	MaxDuration       uint     `form:"max_duration" binding:"omitempty,gtefield=MinDuration"`
	MinBytes          uint     `form:"min_bytes"`
	MaxBytes          uint     `form:"max_bytes" binding:"omitempty,gtefield=MinBytes"`
	StartedAfter      int64    `form:"started_after" `
	StartedBefore     int64    `form:"started_before" binding:"omitempty,gtefield=StartedAfter"`
	ClosedAfter       int64    `form:"closed_after" `
	ClosedBefore      int64    `form:"closed_before" binding:"omitempty,gtefield=ClosedAfter"`
	Hidden            bool     `form:"hidden"`
	Marked            bool     `form:"marked"`
	MatchedRules      []string `form:"matched_rules" binding:"dive,hexadecimal,len=24"`
	PerformedSearch   string   `form:"performed_search" binding:"omitempty,hexadecimal,len=24"`
	Limit             int64    `form:"limit"`
	SimilarToID       string   `form:"similar_to_id" binding:"omitempty,hexadecimal,len=24"`
	ClientSimilarToID string   `form:"client_similar_to_id" binding:"omitempty,hexadecimal,len=24"`
	ServerSimilarToID string   `form:"server_similar_to_id" binding:"omitempty,hexadecimal,len=24"`
}

type ConnectionsController struct {
	storage            Storage
	searchController   *SearchController
	servicesController *ServicesController
}

func NewConnectionsController(storage Storage, searchesController *SearchController,
	servicesController *ServicesController) ConnectionsController {
	return ConnectionsController{
		storage:            storage,
		searchController:   searchesController,
		servicesController: servicesController,
	}
}

func (cc ConnectionsController) GetConnections(c context.Context, filter ConnectionsFilter) []Connection {
	var connections []Connection
	query := cc.storage.Find(Connections).Context(c)

	from, _ := RowIDFromHex(filter.From)
	if !from.IsZero() {
		query = query.Filter(OrderedDocument{{"_id", UnorderedDocument{"$lte": from}}})
	}
	to, _ := RowIDFromHex(filter.To)
	if !to.IsZero() {
		query = query.Filter(OrderedDocument{{"_id", UnorderedDocument{"$gte": to}}})
	} else {
		query = query.Sort("_id", false)
	}
	if filter.ServicePort > 0 {
		query = query.Filter(OrderedDocument{{"port_dst", filter.ServicePort}})
	}
	if len(filter.ClientAddress) > 0 {
		query = query.Filter(OrderedDocument{{"ip_src", filter.ClientAddress}})
	}
	if filter.ClientPort > 0 {
		query = query.Filter(OrderedDocument{{"port_src", filter.ClientPort}})
	}
	if filter.MinDuration > 0 {
		query = query.Filter(OrderedDocument{{"$where", fmt.Sprintf("this.closed_at - this.started_at >= %v", filter.MinDuration)}})
	}
	if filter.MaxDuration > 0 {
		query = query.Filter(OrderedDocument{{"$where", fmt.Sprintf("this.closed_at - this.started_at <= %v", filter.MaxDuration)}})
	}
	if filter.MinBytes > 0 {
		query = query.Filter(OrderedDocument{{"$where", fmt.Sprintf("this.client_bytes + this.server_bytes >= %v", filter.MinBytes)}})
	}
	if filter.MaxBytes > 0 {
		query = query.Filter(OrderedDocument{{"$where", fmt.Sprintf("this.client_bytes + this.server_bytes <= %v", filter.MaxBytes)}})
	}
	if filter.StartedAfter > 0 {
		query = query.Filter(OrderedDocument{{"started_at", UnorderedDocument{"$gt": time.Unix(filter.StartedAfter, 0)}}})
	}
	if filter.StartedBefore > 0 {
		query = query.Filter(OrderedDocument{{"started_at", UnorderedDocument{"$lt": time.Unix(filter.StartedBefore, 0)}}})
	}
	if filter.ClosedAfter > 0 {
		query = query.Filter(OrderedDocument{{"closed_at", UnorderedDocument{"$gt": time.Unix(filter.ClosedAfter, 0)}}})
	}
	if filter.ClosedBefore > 0 {
		query = query.Filter(OrderedDocument{{"closed_at", UnorderedDocument{"$lt": time.Unix(filter.ClosedBefore, 0)}}})
	}
	if filter.Hidden {
		query = query.Filter(OrderedDocument{{"hidden", true}})
	}
	if filter.Marked {
		query = query.Filter(OrderedDocument{{"marked", true}})
	}
	if filter.MatchedRules != nil && len(filter.MatchedRules) > 0 {
		matchedRules := make([]RowID, len(filter.MatchedRules))
		for i, elem := range filter.MatchedRules {
			if id, err := RowIDFromHex(elem); err != nil {
				log.WithError(err).WithField("filter", filter).Panic("failed to convert matched_rules ids")
			} else {
				matchedRules[i] = id
			}
		}

		query = query.Filter(OrderedDocument{{"matched_rules", UnorderedDocument{"$all": matchedRules}}})
	}
	performedSearchID, _ := RowIDFromHex(filter.PerformedSearch)
	if !performedSearchID.IsZero() {
		performedSearch := cc.searchController.GetPerformedSearch(performedSearchID)
		if !performedSearch.ID.IsZero() && len(performedSearch.AffectedConnections) > 0 {
			query = query.Filter(OrderedDocument{{"_id", UnorderedDocument{"$in": performedSearch.AffectedConnections}}})
		}
	}
	var limit int64
	if filter.Limit > 0 && filter.Limit <= MaxQueryLimit {
		limit = filter.Limit
	} else {
		limit = DefaultQueryLimit
	}
	query = query.Limit(limit)

	var similarTo int
	var similarToID RowID
	if id, err := RowIDFromHex(filter.SimilarToID); err == nil {
		similarTo = SimilarToBoth
		similarToID = id
	} else if id, err := RowIDFromHex(filter.ClientSimilarToID); err == nil {
		similarTo = SimilarToClient
		similarToID = id
	} else if id, err := RowIDFromHex(filter.ServerSimilarToID); err == nil {
		similarTo = SimilarToServer
		similarToID = id
	}

	var err error
	if !similarToID.IsZero() {
		if similarToConn, ok := cc.GetConnection(c, similarToID); ok {
			connections = make([]Connection, 0, limit)
			// enforce similarity on same service to reduce the number of documents extracted
			query = query.Filter(OrderedDocument{{"port_dst", similarToConn.DestinationPort}})

			if similarTo == SimilarToBoth || similarTo == SimilarToClient { // filter documents with similar client size
				query = query.Filter(OrderedDocument{{"$expr", UnorderedDocument{
					"$gt": []interface{}{
						UnorderedDocument{"$divide": []UnorderedDocument{
							{"$min": []interface{}{
								"$client_bytes", similarToConn.ClientBytes + 1, // avoid division by zero
							}},
							{"$max": []interface{}{
								"$client_bytes", similarToConn.ClientBytes + 1,
							}},
						}},
						MinSimilaritySizeRatioThreshold,
					},
				}}})
			}

			if similarTo == SimilarToBoth || similarTo == SimilarToServer { // filter documents with similar server size
				query = query.Filter(OrderedDocument{{"$expr", UnorderedDocument{
					"$gt": []interface{}{
						UnorderedDocument{"$divide": []UnorderedDocument{
							{"$min": []interface{}{
								"$server_bytes", similarToConn.ServerBytes + 1, // avoid division by zero
							}},
							{"$max": []interface{}{
								"$server_bytes", similarToConn.ServerBytes + 1,
							}},
						}},
						MinSimilaritySizeRatioThreshold,
					},
				}}})
			}

			err = query.Traverse(func() interface{} { return &Connection{} },
				cc.buildSimilarityFunc(similarToConn, &connections, similarTo))
		}
	} else {
		err = query.All(&connections)
	}

	if err != nil {
		log.WithError(err).WithField("filter", filter).Panic("failed to get connections")
	}

	if connections == nil {
		return []Connection{}
	}

	services := cc.servicesController.GetServices()
	for i, connection := range connections {
		if service, isPresent := services[connection.DestinationPort]; isPresent {
			connections[i].Service = service
		}
	}

	if !to.IsZero() {
		connections = reverseConnections(connections)
	}

	return connections
}

func (cc ConnectionsController) GetConnection(c context.Context, id RowID) (Connection, bool) {
	var connection Connection
	if err := cc.storage.Find(Connections).Context(c).Filter(byID(id)).First(&connection); err != nil {
		log.WithError(err).WithField("id", id).Panic("failed to get connection")
	}

	return connection, !connection.ID.IsZero()
}

func (cc ConnectionsController) SetHidden(c context.Context, id RowID, hidden bool) bool {
	return cc.setProperty(c, id, "hidden", hidden)
}

func (cc ConnectionsController) SetMarked(c context.Context, id RowID, marked bool) bool {
	return cc.setProperty(c, id, "marked", marked)
}

func (cc ConnectionsController) SetComment(c context.Context, id RowID, comment string) bool {
	return cc.setProperty(c, id, "comment", comment)
}

func (cc ConnectionsController) setProperty(c context.Context, id RowID, propertyName string, propertyValue interface{}) bool {
	updated, err := cc.storage.Update(Connections).Context(c).Filter(byID(id)).
		One(UnorderedDocument{propertyName: propertyValue})
	if err != nil {
		log.WithError(err).WithField("id", id).Panic("failed to update a connection property")
	}
	return updated
}

func (cc ConnectionsController) buildSimilarityFunc(similarToConn Connection,
	results *[]Connection, similarTo int) func(interface{}) bool {
	connClientComparable := similarToConn.ClientStreamComparable()
	connServerComparable := similarToConn.ServerStreamComparable()

	// duplicated code for optimization
	if similarTo == SimilarToClient {
		return func(row interface{}) bool {
			conn := row.(*Connection)
			if connClientComparable.IsSimilarTo(conn.ClientStreamComparable()) {
				*results = append(*results, *conn)
				return true
			}
			return false
		}
	} else if similarTo == SimilarToClient {
		return func(row interface{}) bool {
			conn := row.(*Connection)
			if connServerComparable.IsSimilarTo(conn.ServerStreamComparable()) {
				*results = append(*results, *conn)
				return true
			}
			return false
		}
	} else {
		return func(row interface{}) bool {
			conn := row.(*Connection)
			if connClientComparable.IsSimilarTo(conn.ClientStreamComparable()) &&
				connServerComparable.IsSimilarTo(conn.ServerStreamComparable()) {
				*results = append(*results, *conn)
				return true
			}
			return false
		}
	}
}

func (c Connection) ClientStreamComparable() similarity.ComparableStream {
	return similarity.ComparableStream{
		Size:          c.ClientBytes,
		Tlsh:          similarity.ParseTlshDigest(c.ClientTlshHash),
		ByteHistogram: similarity.ByteHistogramFromDigest(c.ClientByteHistogramHash),
	}
}

func (c Connection) ServerStreamComparable() similarity.ComparableStream {
	return similarity.ComparableStream{
		Size:          c.ServerBytes,
		Tlsh:          similarity.ParseTlshDigest(c.ServerTlshHash),
		ByteHistogram: similarity.ByteHistogramFromDigest(c.ServerByteHistogramHash),
	}
}

func reverseConnections(connections []Connection) []Connection {
	for i := 0; i < len(connections)/2; i++ {
		j := len(connections) - i - 1
		connections[i], connections[j] = connections[j], connections[i]
	}
	return connections
}
