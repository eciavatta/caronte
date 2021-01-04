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
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

const (
	secondsToNano     = 1000 * 1000 * 1000
	maxSearchTimeout  = 10 * secondsToNano
	maxRecentSearches = 200
)

type PerformedSearch struct {
	ID                       RowID         `bson:"_id" json:"id"`
	SearchOptions            SearchOptions `bson:"search_options" json:"search_options"`
	AffectedConnections      []RowID       `bson:"affected_connections" json:"-"`
	AffectedConnectionsCount int           `bson:"affected_connections_count" json:"affected_connections_count"`
	StartedAt                time.Time     `bson:"started_at" json:"started_at"`
	FinishedAt               time.Time     `bson:"finished_at" json:"finished_at"`
	UpdatedAt                time.Time     `bson:"updated_at" json:"updated_at"`
	Timeout                  time.Duration `bson:"timeout" json:"timeout"`
}

type SearchOptions struct {
	TextSearch  TextSearch    `bson:"text_search" json:"text_search"`
	RegexSearch RegexSearch   `bson:"regex_search" json:"regex_search"`
	Timeout     time.Duration `bson:"timeout" json:"timeout" binding:"max=60"`
}

type TextSearch struct {
	Terms         []string `bson:"terms" json:"terms" binding:"isdefault|min=1,dive,min=3"`
	ExcludedTerms []string `bson:"excluded_terms" json:"excluded_terms" binding:"isdefault|min=1,dive,min=3"`
	ExactPhrase   string   `bson:"exact_phrase" json:"exact_phrase" binding:"isdefault|min=3"`
	CaseSensitive bool     `bson:"case_sensitive" json:"case_sensitive"`
}

type RegexSearch struct {
	Pattern           string `bson:"pattern" json:"pattern" binding:"isdefault|min=3"`
	NotPattern        string `bson:"not_pattern" json:"not_pattern" binding:"isdefault|min=3"`
	CaseInsensitive   bool   `bson:"case_insensitive" json:"case_insensitive"`
	MultiLine         bool   `bson:"multi_line" json:"multi_line"`
	IgnoreWhitespaces bool   `bson:"ignore_whitespaces" json:"ignore_whitespaces"`
	DotCharacter      bool   `bson:"dot_character" json:"dot_character"`
}

type SearchController struct {
	storage           Storage
	performedSearches []PerformedSearch
	mutex             sync.Mutex
}

func NewSearchController(storage Storage) *SearchController {
	var searches []PerformedSearch
	if err := storage.Find(Searches).Limit(maxRecentSearches).All(&searches); err != nil {
		log.WithError(err).Panic("failed to retrieve performed searches")
	}

	if searches == nil {
		searches = []PerformedSearch{}
	}

	return &SearchController{
		storage:           storage,
		performedSearches: searches,
	}
}

func (sc *SearchController) GetPerformedSearches() []PerformedSearch {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	return sc.performedSearches
}

func (sc *SearchController) GetPerformedSearch(id RowID) PerformedSearch {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	var performedSearch PerformedSearch
	for _, search := range sc.performedSearches {
		if search.ID == id {
			performedSearch = search
		}
	}

	return performedSearch
}

func (sc *SearchController) PerformSearch(c context.Context, options SearchOptions) PerformedSearch {
	findQuery := sc.storage.Find(ConnectionStreams).Projection(OrderedDocument{{"connection_id", 1}}).Context(c)
	timeout := options.Timeout * secondsToNano
	if timeout <= 0 || timeout > maxSearchTimeout {
		timeout = maxSearchTimeout
	}
	findQuery = findQuery.MaxTime(timeout)

	if !options.TextSearch.isZero() {
		var text string
		if options.TextSearch.ExactPhrase != "" {
			text = "\"" + options.TextSearch.ExactPhrase + "\""
		} else {
			text = strings.Join(options.TextSearch.Terms, " ")
			if options.TextSearch.ExcludedTerms != nil {
				text += " -" + strings.Join(options.TextSearch.ExcludedTerms, " -")
			}
		}

		findQuery = findQuery.Filter(OrderedDocument{{"$text", UnorderedDocument{
			"$search":             text,
			"$language":           "none",
			"$caseSensitive":      options.TextSearch.CaseSensitive,
			"$diacriticSensitive": false,
		}}})
	} else {
		var regexOptions string
		if options.RegexSearch.CaseInsensitive {
			regexOptions += "i"
		}
		if options.RegexSearch.MultiLine {
			regexOptions += "m"
		}
		if options.RegexSearch.IgnoreWhitespaces {
			regexOptions += "x"
		}
		if options.RegexSearch.DotCharacter {
			regexOptions += "s"
		}

		var regex UnorderedDocument
		if options.RegexSearch.Pattern != "" {
			regex = UnorderedDocument{"$regex": options.RegexSearch.Pattern, "$options": regexOptions}
		} else {
			regex = UnorderedDocument{"$not": UnorderedDocument{"$regex": options.RegexSearch.NotPattern, "$options": regexOptions}}
		}

		findQuery = findQuery.Filter(OrderedDocument{{"payload_string", regex}})
	}

	var connections []ConnectionStream
	startedAt := time.Now()
	if err := findQuery.All(&connections); err != nil {
		log.WithError(err).Error("oh no")
	}
	affectedConnections := uniqueConnectionIds(connections)

	finishedAt := time.Now()
	performedSearch := PerformedSearch{
		ID:                       NewRowID(),
		SearchOptions:            options,
		AffectedConnections:      affectedConnections,
		AffectedConnectionsCount: len(affectedConnections),
		StartedAt:                startedAt,
		FinishedAt:               finishedAt,
		UpdatedAt:                finishedAt,
		Timeout:                  options.Timeout,
	}
	if _, err := sc.storage.Insert(Searches).Context(c).One(performedSearch); err != nil {
		log.WithError(err).Panic("failed to insert a new performed search")
	}

	sc.mutex.Lock()
	sc.performedSearches = append([]PerformedSearch{performedSearch}, sc.performedSearches...)
	if len(sc.performedSearches) > maxRecentSearches {
		sc.performedSearches = sc.performedSearches[:200]
	}
	sc.mutex.Unlock()

	return performedSearch
}

func (sc TextSearch) isZero() bool {
	return sc.Terms == nil && sc.ExcludedTerms == nil && sc.ExactPhrase == ""
}

func (sc RegexSearch) isZero() bool {
	return RegexSearch{} == sc
}

func uniqueConnectionIds(connections []ConnectionStream) []RowID {
	keys := make(map[RowID]bool)
	var out []RowID
	for _, entry := range connections {
		if _, value := keys[entry.ConnectionID]; !value {
			keys[entry.ConnectionID] = true
			out = append(out, entry.ConnectionID)
		}
	}
	return out
}
