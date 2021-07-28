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
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/flier/gohs/hyperscan"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

const DirectionBoth = 0
const DirectionToServer = 1
const DirectionToClient = 2

type RegexFlags struct {
	Caseless        bool `json:"caseless" bson:"caseless,omitempty"`                 // Set case-insensitive matching.
	DotAll          bool `json:"dot_all" bson:"dot_all,omitempty"`                   // Matching a `.` will not exclude newlines.
	MultiLine       bool `json:"multi_line" bson:"multi_line,omitempty"`             // Set multi-line anchoring.
	Utf8Mode        bool `json:"utf_8_mode" bson:"utf_8_mode,omitempty"`             // Enable UTF-8 mode for this expression.
	UnicodeProperty bool `json:"unicode_property" bson:"unicode_property,omitempty"` // Enable Unicode property support for this expression
}

type Pattern struct {
	Regex          string     `json:"regex" binding:"required,min=1" bson:"regex"`
	Flags          RegexFlags `json:"flags" bson:"flags,omitempty"`
	MinOccurrences uint       `json:"min_occurrences" bson:"min_occurrences,omitempty"`
	MaxOccurrences uint       `json:"max_occurrences" binding:"omitempty,gtefield=MinOccurrences" bson:"max_occurrences,omitempty"`
	Direction      uint8      `json:"direction" binding:"omitempty,max=2" bson:"direction,omitempty"`
	internalID     uint
}

type Filter struct {
	ServicePort   uint16 `json:"service_port" bson:"service_port,omitempty"`
	ClientAddress string `json:"client_address" binding:"omitempty,ip" bson:"client_address,omitempty"`
	ClientPort    uint16 `json:"client_port" bson:"client_port,omitempty"`
	MinDuration   uint   `json:"min_duration" bson:"min_duration,omitempty"`
	MaxDuration   uint   `json:"max_duration" binding:"omitempty,gtefield=MinDuration" bson:"max_duration,omitempty"`
	MinBytes      uint   `json:"min_bytes" bson:"min_bytes,omitempty"`
	MaxBytes      uint   `json:"max_bytes" binding:"omitempty,gtefield=MinBytes" bson:"max_bytes,omitempty"`
}

type Rule struct {
	ID       RowID     `json:"id" bson:"_id,omitempty"`
	Name     string    `json:"name" binding:"min=3" bson:"name"`
	Color    string    `json:"color" binding:"hexcolor" bson:"color"`
	Notes    string    `json:"notes" bson:"notes,omitempty"`
	Enabled  bool      `json:"enabled" bson:"enabled"`
	Patterns []Pattern `json:"patterns" bson:"patterns"`
	Filter   Filter    `json:"filter" bson:"filter,omitempty"`
	Version  int64     `json:"version" bson:"version"`
}

type RulesDatabase struct {
	database     hyperscan.StreamDatabase
	databaseSize int
	version      RowID
}

type RulesManager interface {
	AddRule(context context.Context, rule Rule) (RowID, error)
	GetRule(id RowID) (Rule, bool)
	UpdateRule(context context.Context, id RowID, rule Rule) (bool, error)
	GetRules() []Rule
	FillWithMatchedRules(connection *Connection, clientMatches map[uint][]PatternSlice, serverMatches map[uint][]PatternSlice)
	DatabaseUpdateChannel() chan RulesDatabase
}

type rulesManagerImpl struct {
	storage         Storage
	rules           map[RowID]Rule
	rulesByName     map[string]Rule
	patterns        []*hyperscan.Pattern
	patternsIds     map[string]uint
	mutex           sync.Mutex
	databaseUpdated chan RulesDatabase
	validate        *validator.Validate
}

func LoadRulesManager(storage Storage, flagRegex string) (RulesManager, error) {
	var rules []Rule
	if err := storage.Find(Rules).Sort("_id", true).All(&rules); err != nil {
		return nil, err
	}

	rulesManager := rulesManagerImpl{
		storage:         storage,
		rules:           make(map[RowID]Rule),
		rulesByName:     make(map[string]Rule),
		patterns:        make([]*hyperscan.Pattern, 0),
		patternsIds:     make(map[string]uint),
		mutex:           sync.Mutex{},
		databaseUpdated: make(chan RulesDatabase, 1),
		validate:        validator.New(),
	}

	for _, rule := range rules {
		if err := rulesManager.validateAndAddRuleLocal(&rule); err != nil {
			return nil, err
		}
	}

	// if there are no rules in database (e.g. first run), set flagRegex as first rule
	if len(rulesManager.rules) == 0 {
		_, _ = rulesManager.AddRule(context.Background(), Rule{
			Name:  "flag_out",
			Color: "#e53935",
			Notes: "Mark connections where the flags are stolen",
			Patterns: []Pattern{
				{Regex: flagRegex, Direction: DirectionToClient, Flags: RegexFlags{Utf8Mode: true}},
			},
		})
		_, _ = rulesManager.AddRule(context.Background(), Rule{
			Name:  "flag_in",
			Color: "#43A047",
			Notes: "Mark connections where the flags are placed",
			Patterns: []Pattern{
				{Regex: flagRegex, Direction: DirectionToServer, Flags: RegexFlags{Utf8Mode: true}},
			},
		})
	} else {
		if err := rulesManager.generateDatabase(rules[len(rules)-1].ID); err != nil {
			return nil, err
		}
	}

	return &rulesManager, nil
}

func (rm *rulesManagerImpl) AddRule(context context.Context, rule Rule) (RowID, error) {
	rm.mutex.Lock()

	rule.ID = CustomRowID(uint64(len(rm.rules)), time.Now())
	rule.Enabled = true

	if err := rm.validateAndAddRuleLocal(&rule); err != nil {
		rm.mutex.Unlock()
		return EmptyRowID(), err
	}

	if err := rm.generateDatabase(rule.ID); err != nil {
		rm.mutex.Unlock()
		log.WithError(err).WithField("rule", rule).Panic("failed to generate database")
	}
	rm.mutex.Unlock()

	if _, err := rm.storage.Insert(Rules).Context(context).One(rule); err != nil {
		log.WithError(err).WithField("rule", rule).Panic("failed to insert rule on database")
	}

	return rule.ID, nil
}

func (rm *rulesManagerImpl) GetRule(id RowID) (Rule, bool) {
	rule, isPresent := rm.rules[id]
	return rule, isPresent
}

func (rm *rulesManagerImpl) UpdateRule(context context.Context, id RowID, rule Rule) (bool, error) {
	newRule, isPresent := rm.rules[id]
	if !isPresent {
		return false, nil
	}

	sameName, isPresent := rm.rulesByName[rule.Name]
	if isPresent && sameName.ID != id {
		return false, errors.New("already exists another rule with the same name")
	}

	updated, err := rm.storage.Update(Rules).Context(context).Filter(OrderedDocument{{"_id", id}}).
		One(UnorderedDocument{"name": rule.Name, "color": rule.Color})
	if err != nil {
		log.WithError(err).WithField("rule", rule).Panic("failed to update rule on database")
	}

	if updated {
		rm.mutex.Lock()
		newRule.Name = rule.Name
		newRule.Color = rule.Color

		delete(rm.rulesByName, newRule.Name)
		rm.rulesByName[newRule.Name] = newRule
		rm.rules[id] = newRule
		rm.mutex.Unlock()
	}

	return updated, nil
}

func (rm *rulesManagerImpl) GetRules() []Rule {
	rules := make([]Rule, 0, len(rm.rules))

	for _, rule := range rm.rules {
		rules = append(rules, rule)
	}

	sort.Slice(rules, func(i, j int) bool {
		return rules[i].ID.Timestamp().Before(rules[j].ID.Timestamp())
	})

	return rules
}

func (rm *rulesManagerImpl) FillWithMatchedRules(connection *Connection, clientMatches map[uint][]PatternSlice,
	serverMatches map[uint][]PatternSlice) {
	rm.mutex.Lock()

	filterFunctions := []func(rule Rule) bool{
		func(rule Rule) bool {
			return rule.Filter.ClientAddress == "" || connection.SourceIP == rule.Filter.ClientAddress
		},
		func(rule Rule) bool {
			return rule.Filter.ClientPort == 0 || connection.SourcePort == rule.Filter.ClientPort
		},
		func(rule Rule) bool {
			return rule.Filter.ServicePort == 0 || connection.DestinationPort == rule.Filter.ServicePort
		},
		func(rule Rule) bool {
			return rule.Filter.MinDuration == 0 || uint(connection.ClosedAt.Sub(connection.StartedAt).Milliseconds()) >=
				rule.Filter.MinDuration
		},
		func(rule Rule) bool {
			return rule.Filter.MaxDuration == 0 || uint(connection.ClosedAt.Sub(connection.StartedAt).Milliseconds()) <=
				rule.Filter.MaxDuration
		},
		func(rule Rule) bool {
			return rule.Filter.MinBytes == 0 || uint(connection.ClientBytes+connection.ServerBytes) >=
				rule.Filter.MinBytes
		},
		func(rule Rule) bool {
			return rule.Filter.MaxBytes == 0 || uint(connection.ClientBytes+connection.ServerBytes) <=
				rule.Filter.MinBytes
		},
	}

	connection.MatchedRules = make([]RowID, 0)
	for _, rule := range rm.rules {
		matching := true
		for _, f := range filterFunctions {
			if !f(rule) {
				matching = false
				break
			}
		}

		for _, p := range rule.Patterns {
			checkOccurrences := func(occurrences []PatternSlice) bool {
				return (p.MinOccurrences == 0 || uint(len(occurrences)) >= p.MinOccurrences) &&
					(p.MaxOccurrences == 0 || uint(len(occurrences)) <= p.MaxOccurrences)
			}
			clientOccurrences, clientPresent := clientMatches[p.internalID]
			serverOccurrences, serverPresent := serverMatches[p.internalID]

			if p.Direction == DirectionToServer {
				if !clientPresent || !checkOccurrences(clientOccurrences) {
					matching = false
					break
				}
			} else if p.Direction == DirectionToClient {
				if !serverPresent || !checkOccurrences(serverOccurrences) {
					matching = false
					break
				}
			} else {
				if !(clientPresent || serverPresent) || !checkOccurrences(append(clientOccurrences, serverOccurrences...)) {
					matching = false
					break
				}
			}
		}

		if matching {
			connection.MatchedRules = append(connection.MatchedRules, rule.ID)
		}
	}

	rm.mutex.Unlock()
}

func (rm *rulesManagerImpl) DatabaseUpdateChannel() chan RulesDatabase {
	return rm.databaseUpdated
}

func (rm *rulesManagerImpl) validateAndAddRuleLocal(rule *Rule) error {
	if _, alreadyPresent := rm.rulesByName[rule.Name]; alreadyPresent {
		return errors.New("rule name must be unique")
	}

	newPatterns := make([]*hyperscan.Pattern, 0, len(rule.Patterns))
	duplicatePatterns := make(map[string]bool)
	for i, pattern := range rule.Patterns {
		if err := rm.validate.Struct(pattern); err != nil {
			return err
		}

		regex := pattern.Regex
		if !strings.HasPrefix(regex, "/") {
			regex = fmt.Sprintf("/%s", regex)
		}
		if !strings.HasSuffix(regex, "/") {
			regex = fmt.Sprintf("%s/", regex)
		}
		rule.Patterns[i].Regex = regex

		compiledPattern, err := pattern.BuildPattern()
		if err != nil {
			return err
		}
		regex = compiledPattern.String()
		if _, isPresent := duplicatePatterns[regex]; isPresent {
			return errors.New("duplicate pattern")
		}
		if existingPattern, isPresent := rm.patternsIds[regex]; isPresent {
			rule.Patterns[i].internalID = existingPattern
			continue
		}

		id := len(rm.patternsIds) + len(newPatterns)
		rule.Patterns[i].internalID = uint(id)
		compiledPattern.Id = id
		newPatterns = append(newPatterns, compiledPattern)
		duplicatePatterns[regex] = true
	}

	startID := len(rm.patterns)
	for id, pattern := range newPatterns {
		rm.patterns = append(rm.patterns, pattern)
		regex := pattern.String()
		rm.patternsIds[regex[strings.IndexByte(regex, ':')+1:]] = uint(startID + id)
	}

	rm.rules[rule.ID] = *rule
	rm.rulesByName[rule.Name] = *rule

	return nil
}

func (rm *rulesManagerImpl) generateDatabase(version RowID) error {
	database, err := hyperscan.NewStreamDatabase(rm.patterns...)
	if err != nil {
		return err
	}

	go func() {
		rm.databaseUpdated <- RulesDatabase{
			database:     database,
			databaseSize: len(rm.patterns),
			version:      version,
		}
	}()

	return nil
}

func (p *Pattern) BuildPattern() (*hyperscan.Pattern, error) {
	hp, err := hyperscan.ParsePattern(p.Regex)
	if err != nil {
		return nil, err
	}

	hp.Flags |= hyperscan.SomLeftMost
	if p.Flags.Caseless {
		hp.Flags |= hyperscan.Caseless
	}
	if p.Flags.DotAll {
		hp.Flags |= hyperscan.DotAll
	}
	if p.Flags.MultiLine {
		hp.Flags |= hyperscan.MultiLine
	}
	if p.Flags.Utf8Mode {
		hp.Flags |= hyperscan.Utf8Mode
	}
	if p.Flags.UnicodeProperty {
		hp.Flags |= hyperscan.UnicodeProperty
	}

	if !hp.IsValid() {
		return nil, errors.New("can't validate the pattern")
	}

	return hp, nil
}
