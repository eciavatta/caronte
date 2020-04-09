package main

import "time"

type Connection struct {
	ID              RowID     `json:"id" bson:"_id"`
	SourceIP        string    `json:"ip_src" bson:"ip_src"`
	DestinationIP   string    `json:"ip_dst" bson:"ip_dst"`
	SourcePort      uint16    `json:"port_src" bson:"port_src"`
	DestinationPort uint16    `json:"port_dst" bson:"port_dst"`
	StartedAt       time.Time `json:"started_at" bson:"started_at"`
	ClosedAt        time.Time `json:"closed_at" bson:"closed_at"`
	ClientPackets   int       `json:"client_packets" bson:"client_packets"`
	ServerPackets   int       `json:"server_packets" bson:"server_packets"`
	ClientBytes     int       `json:"client_bytes" bson:"client_bytes"`
	ServerBytes     int       `json:"server_bytes" bson:"server_bytes"`
	ClientDocuments int       `json:"client_documents" bson:"client_documents"`
	ServerDocuments int       `json:"server_documents" bson:"server_documents"`
	ProcessedAt     time.Time `json:"processed_at" bson:"processed_at"`
}
