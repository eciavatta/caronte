package main

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

type StatisticRecord struct {
	RangeStart            time.Time        `json:"range_start" bson:"_id"`
	ConnectionsPerService map[uint16]int   `json:"connections_per_service" bson:"connections_per_service"`
	ClientBytesPerService map[uint16]int   `json:"client_bytes_per_service" bson:"client_bytes_per_service"`
	ServerBytesPerService map[uint16]int   `json:"server_bytes_per_service" bson:"server_bytes_per_service"`
	DurationPerService    map[uint16]int64 `json:"duration_per_service" bson:"duration_per_service"`
}

type StatisticsFilter struct {
	RangeFrom time.Time `form:"range_from"`
	RangeTo   time.Time `form:"range_to"`
	Ports     []uint16  `form:"ports"`
	Metric    string    `form:"metric"`
}

type StatisticsController struct {
	storage Storage
	metrics []string
}

func NewStatisticsController(storage Storage) StatisticsController {
	return StatisticsController{
		storage: storage,
		metrics: []string{"connections_per_service", "client_bytes_per_service",
			"server_bytes_per_service", "duration_per_service"},
	}
}

func (sc *StatisticsController) GetStatistics(context context.Context, filter StatisticsFilter) []StatisticRecord {
	var statisticRecords []StatisticRecord
	query := sc.storage.Find(Statistics).Context(context)
	if !filter.RangeFrom.IsZero() {
		query = query.Filter(OrderedDocument{{"_id", UnorderedDocument{"$lt": filter.RangeFrom}}})
	}
	if !filter.RangeTo.IsZero() {
		query = query.Filter(OrderedDocument{{"_id", UnorderedDocument{"$gt": filter.RangeTo}}})
	}
	for _, port := range filter.Ports {
		for _, metric := range sc.metrics {
			if filter.Metric == "" || filter.Metric == metric {
				query = query.Projection(OrderedDocument{{fmt.Sprintf("%s.%d", metric, port), 1}})
			}
		}
	}
	if filter.Metric != "" && len(filter.Ports) == 0 {
		for _, metric := range sc.metrics {
			if filter.Metric == metric {
				query = query.Projection(OrderedDocument{{metric, 1}})
			}
		}
	}

	if err := query.All(&statisticRecords); err != nil {
		log.WithError(err).WithField("filter", filter).Error("failed to retrieve statistics")
		return []StatisticRecord{}
	}
	if statisticRecords == nil {
		return []StatisticRecord{}
	}

	return statisticRecords
}
