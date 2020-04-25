package main

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

const DefaultQueryLimit = 50
const MaxQueryLimit = 200

type Connection struct {
	ID              RowID     `json:"id" bson:"_id"`
	SourceIP        string    `json:"ip_src" bson:"ip_src"`
	DestinationIP   string    `json:"ip_dst" bson:"ip_dst"`
	SourcePort      uint16    `json:"port_src" bson:"port_src"`
	DestinationPort uint16    `json:"port_dst" bson:"port_dst"`
	StartedAt       time.Time `json:"started_at" bson:"started_at"`
	ClosedAt        time.Time `json:"closed_at" bson:"closed_at"`
	ClientBytes     int       `json:"client_bytes" bson:"client_bytes"`
	ServerBytes     int       `json:"server_bytes" bson:"server_bytes"`
	ClientDocuments int       `json:"client_documents" bson:"client_documents"`
	ServerDocuments int       `json:"server_documents" bson:"server_documents"`
	ProcessedAt     time.Time `json:"processed_at" bson:"processed_at"`
	MatchedRules    []RowID   `json:"matched_rules" bson:"matched_rules"`
	Hidden          bool      `json:"hidden" bson:"hidden,omitempty"`
	Marked          bool      `json:"marked" bson:"marked,omitempty"`
	Comment         string    `json:"comment" bson:"comment,omitempty"`
	Service         Service   `json:"service" bson:"-"`
}

type ConnectionsFilter struct {
	From          string  `form:"from" binding:"omitempty,hexadecimal,len=24"`
	To            string  `form:"to" binding:"omitempty,hexadecimal,len=24"`
	ServicePort   uint16  `form:"service_port"`
	ClientAddress string  `form:"client_address" binding:"omitempty,ip"`
	ClientPort    uint16  `form:"client_port"`
	MinDuration   uint    `form:"min_duration"`
	MaxDuration   uint    `form:"max_duration" binding:"omitempty,gtefield=MinDuration"`
	MinBytes      uint    `form:"min_bytes"`
	MaxBytes      uint    `form:"max_bytes" binding:"omitempty,gtefield=MinBytes"`
	StartedAfter  int64   `form:"started_after" `
	StartedBefore int64   `form:"started_before" binding:"omitempty,gtefield=StartedAfter"`
	ClosedAfter   int64   `form:"closed_after" `
	ClosedBefore  int64   `form:"closed_before" binding:"omitempty,gtefield=ClosedAfter"`
	Hidden        bool    `form:"hidden"`
	Marked        bool    `form:"marked"`
	MatchedRules  []RowID `form:"matched_rules"`
	Limit         int64   `form:"limit"`
}

type ConnectionsController struct {
	storage Storage
	servicesController *ServicesController
}

func NewConnectionsController(storage Storage, servicesController *ServicesController) ConnectionsController {
	return ConnectionsController{
		storage: storage,
		servicesController: servicesController,
	}
}

func (cc ConnectionsController) GetConnections(c context.Context, filter ConnectionsFilter) []Connection {
	var connections []Connection
	query := cc.storage.Find(Connections).Context(c).Sort("_id", false)

	from, _ := RowIDFromHex(filter.From)
	if !from.IsZero() {
		query = query.Filter(OrderedDocument{{"_id", UnorderedDocument{"$lt": from}}})
	}
	to, _ := RowIDFromHex(filter.To)
	if !to.IsZero() {
		query = query.Filter(OrderedDocument{{"_id", UnorderedDocument{"$gt": to}}})
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
		query = query.Filter(OrderedDocument{{"matched_rules", UnorderedDocument{"$all": filter.MatchedRules}}})
	}
	if filter.Limit > 0 && filter.Limit <= MaxQueryLimit {
		query = query.Limit(filter.Limit)
	} else {
		query = query.Limit(DefaultQueryLimit)
	}

	if err := query.All(&connections); err != nil {
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
