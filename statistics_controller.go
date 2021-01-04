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
	RangeEnd              time.Time        `json:"range_end"`
	ConnectionsPerService map[uint16]int64 `json:"connections_per_service" bson:"connections_per_service"`
	ClientBytesPerService map[uint16]int64 `json:"client_bytes_per_service" bson:"client_bytes_per_service"`
	ServerBytesPerService map[uint16]int64 `json:"server_bytes_per_service" bson:"server_bytes_per_service"`
	TotalBytesPerService  map[uint16]int64 `json:"total_bytes_per_service" bson:"total_bytes_per_service"`
	DurationPerService    map[uint16]int64 `json:"duration_per_service" bson:"duration_per_service"`
	MatchedRules          map[string]int64 `json:"matched_rules" bson:"matched_rules"`
}

type StatisticsFilter struct {
	RangeFrom time.Time `form:"range_from"`
	RangeTo   time.Time `form:"range_to"`
	Ports     []uint16  `form:"ports"`
	RulesIDs  []string  `form:"rules_ids"`
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
	query := sc.storage.Find(Statistics).Context(context).Sort("_id", true)
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
			query = query.Projection(OrderedDocument{{fmt.Sprintf("matched_rules.%s", ruleID), 1}})
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

	for i, _ := range statisticRecords {
		statisticRecords[i].RangeEnd = statisticRecords[i].RangeStart.Add(time.Minute)
	}

	return statisticRecords
}

func (sc *StatisticsController) GetTotalStatistics(context context.Context, filter StatisticsFilter) StatisticRecord {
	totalStats := StatisticRecord{}
	statisticsPerMinute := sc.GetStatistics(context, filter)

	if len(statisticsPerMinute) == 0 {
		return totalStats
	}

	totalStats.RangeStart = statisticsPerMinute[0].RangeStart
	totalStats.RangeEnd = statisticsPerMinute[len(statisticsPerMinute)-1].RangeEnd

	if statisticsPerMinute[0].ConnectionsPerService != nil {
		totalStats.ConnectionsPerService = make(map[uint16]int64)
	}
	if statisticsPerMinute[0].ClientBytesPerService != nil {
		totalStats.ClientBytesPerService = make(map[uint16]int64)
	}
	if statisticsPerMinute[0].ServerBytesPerService != nil {
		totalStats.ServerBytesPerService = make(map[uint16]int64)
	}
	if statisticsPerMinute[0].TotalBytesPerService != nil {
		totalStats.TotalBytesPerService = make(map[uint16]int64)
	}
	if statisticsPerMinute[0].DurationPerService != nil {
		totalStats.DurationPerService = make(map[uint16]int64)
	}
	if statisticsPerMinute[0].MatchedRules != nil {
		totalStats.MatchedRules = make(map[string]int64)
	}

	aggregateServicesMap := func(accumulator map[uint16]int64, record map[uint16]int64) {
		if accumulator == nil || record == nil {
			return
		}
		for k, v := range record {
			accumulator[k] += v
		}
	}

	aggregateMatchedRulesMap := func(accumulator map[string]int64, record map[string]int64) {
		if accumulator == nil || record == nil {
			return
		}
		for k, v := range record {
			accumulator[k] += v
		}
	}

	for _, record := range statisticsPerMinute {
		aggregateServicesMap(totalStats.ConnectionsPerService, record.ConnectionsPerService)
		aggregateServicesMap(totalStats.ClientBytesPerService, record.ClientBytesPerService)
		aggregateServicesMap(totalStats.ServerBytesPerService, record.ServerBytesPerService)
		aggregateServicesMap(totalStats.TotalBytesPerService, record.TotalBytesPerService)
		aggregateServicesMap(totalStats.DurationPerService, record.DurationPerService)
		aggregateMatchedRulesMap(totalStats.MatchedRules, record.MatchedRules)
	}

	return totalStats
}
