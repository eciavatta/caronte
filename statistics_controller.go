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
	log "github.com/sirupsen/logrus"
	"time"
)

type StatisticRecord struct {
	RangeStart            time.Time        `json:"range_start" bson:"_id"`
	ConnectionsPerService map[uint16]int   `json:"connections_per_service,omitempty" bson:"connections_per_service"`
	ClientBytesPerService map[uint16]int   `json:"client_bytes_per_service,omitempty" bson:"client_bytes_per_service"`
	ServerBytesPerService map[uint16]int   `json:"server_bytes_per_service,omitempty" bson:"server_bytes_per_service"`
	TotalBytesPerService  map[uint16]int   `json:"total_bytes_per_service,omitempty" bson:"total_bytes_per_service"`
	DurationPerService    map[uint16]int64 `json:"duration_per_service,omitempty" bson:"duration_per_service"`
	MatchedRules          map[RowID]int64  `json:"matched_rules,omitempty" bson:"matched_rules"`
}

type StatisticsFilter struct {
	RangeFrom time.Time `form:"range_from"`
	RangeTo   time.Time `form:"range_to"`
	Ports     []uint16  `form:"ports"`
	RulesIDs  []RowID   `form:"rules_ids"`
	Metric    string    `form:"metric"`
}

type StatisticsController struct {
	storage         Storage
	servicesMetrics []string
}

func NewStatisticsController(storage Storage) StatisticsController {
	return StatisticsController{
		storage: storage,
		servicesMetrics: []string{"connections_per_service", "client_bytes_per_service",
			"server_bytes_per_service", "total_bytes_per_service", "duration_per_service"},
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
		for _, metric := range sc.servicesMetrics {
			if filter.Metric == "" || filter.Metric == metric {
				query = query.Projection(OrderedDocument{{fmt.Sprintf("%s.%d", metric, port), 1}})
			}
		}

	}
	if filter.Metric != "" && len(filter.Ports) == 0 {
		for _, metric := range sc.servicesMetrics {
			if filter.Metric == metric {
				query = query.Projection(OrderedDocument{{metric, 1}})
			}
		}
	}
	for _, ruleID := range filter.RulesIDs {
		if filter.Metric == "" || filter.Metric == "matched_rules" {
			query = query.Projection(OrderedDocument{{fmt.Sprintf("matched_rules.%s", ruleID.Hex()), 1}})
		}

	}
	if filter.Metric != "" && len(filter.RulesIDs) == 0 {
		if filter.Metric == "matched_rules" {
			query = query.Projection(OrderedDocument{{"matched_rules", 1}})
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
