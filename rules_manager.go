package main

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/flier/gohs/hyperscan"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type RegexFlags struct {
	Caseless        bool `json:"caseless" bson:"caseless,omitempty"`                 // Set case-insensitive matching.
	DotAll          bool `json:"dot_all" bson:"dot_all,omitempty"`                   // Matching a `.` will not exclude newlines.
	MultiLine       bool `json:"multi_line" bson:"multi_line,omitempty"`             // Set multi-line anchoring.
	SingleMatch     bool `json:"single_match" bson:"single_match,omitempty"`         // Set single-match only mode.
	Utf8Mode        bool `json:"utf_8_mode" bson:"utf_8_mode,omitempty"`             // Enable UTF-8 mode for this expression.
	UnicodeProperty bool `json:"unicode_property" bson:"unicode_property,omitempty"` // Enable Unicode property support for this expression
}

type Pattern struct {
	Regex           string     `json:"regex" binding:"min=1" bson:"regex"`
	Flags           RegexFlags `json:"flags" bson:"flags,omitempty"`
	MinOccurrences  uint       `json:"min_occurrences" bson:"min_occurrences,omitempty"`
	MaxOccurrences  uint       `json:"max_occurrences" binding:"omitempty,gtefield=MinOccurrences" bson:"max_occurrences,omitempty"`
	internalID      int
	compiledPattern *hyperscan.Pattern
}

type Filter struct {
	ServicePort   uint16 `json:"service_port" bson:"service_port,omitempty"`
	ClientAddress string `json:"client_address" binding:"ip_addr" bson:"client_address,omitempty"`
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
	Patterns []Pattern `json:"patterns" binding:"required,min=1" bson:"patterns"`
	Filter   Filter    `json:"filter" bson:"filter,omitempty"`
	Version  int64     `json:"version" bson:"version"`
}

type RulesDatabase struct {
	database     hyperscan.StreamDatabase
	databaseSize int
	version      RowID
}

type RulesManager interface {
	LoadRules() error
	AddRule(context context.Context, rule Rule) (RowID, error)
	GetRule(id RowID) (Rule, bool)
	UpdateRule(context context.Context, rule Rule) bool
	GetRules() []Rule
	FillWithMatchedRules(connection *Connection, clientMatches map[uint][]PatternSlice, serverMatches map[uint][]PatternSlice)
	DatabaseUpdateChannel() chan RulesDatabase
}

type rulesManagerImpl struct {
	storage         Storage
	rules           map[RowID]Rule
	rulesByName     map[string]Rule
	patterns        map[string]Pattern
	mutex           sync.Mutex
	databaseUpdated chan RulesDatabase
	validate        *validator.Validate
}

func NewRulesManager(storage Storage) RulesManager {
	return &rulesManagerImpl{
		storage:         storage,
		rules:           make(map[RowID]Rule),
		rulesByName:     make(map[string]Rule),
		patterns:        make(map[string]Pattern),
		mutex:           sync.Mutex{},
		databaseUpdated: make(chan RulesDatabase),
		validate:        validator.New(),
	}
}

func (rm *rulesManagerImpl) LoadRules() error {
	var rules []Rule
	if err := rm.storage.Find(Rules).Sort("_id", true).All(&rules); err != nil {
		return err
	}

	for _, rule := range rules {
		if err := rm.validateAndAddRuleLocal(&rule); err != nil {
			log.WithError(err).WithField("rule", rule).Warn("failed to import rule")
		}
	}

	return rm.generateDatabase(rules[len(rules)-1].ID)
}

func (rm *rulesManagerImpl) AddRule(context context.Context, rule Rule) (RowID, error) {
	rm.mutex.Lock()

	rule.ID = rm.storage.NewCustomRowID(uint64(len(rm.rules)), time.Now())
	rule.Enabled = true

	if err := rm.validateAndAddRuleLocal(&rule); err != nil {
		rm.mutex.Unlock()
		return rule.ID, err
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

func (rm *rulesManagerImpl) UpdateRule(context context.Context, rule Rule) bool {
	updated, err := rm.storage.Update(Rules).Context(context).Filter(OrderedDocument{{"_id", rule.ID}}).
		One(UnorderedDocument{"name": rule.Name, "color": rule.Color})
	if err != nil {
		log.WithError(err).WithField("rule", rule).Panic("failed to update rule on database")
	}

	if updated {
		rm.mutex.Lock()
		newRule := rm.rules[rule.ID]
		newRule.Name = rule.Name
		newRule.Color = rule.Color

		delete(rm.rulesByName, newRule.Name)
		rm.rulesByName[rule.Name] = newRule
		rm.rules[rule.ID] = newRule
		rm.mutex.Unlock()
	}

	return updated
}

func (rm *rulesManagerImpl) GetRules() []Rule {
	rules := make([]Rule, 0, len(rm.rules))

	for _, rule := range rm.rules {
		rules = append(rules, rule)
	}

	return rules
}

func (rm *rulesManagerImpl) FillWithMatchedRules(connection *Connection, clientMatches map[uint][]PatternSlice,
	serverMatches map[uint][]PatternSlice) {
}

func (rm *rulesManagerImpl) DatabaseUpdateChannel() chan RulesDatabase {
	return rm.databaseUpdated
}

func (rm *rulesManagerImpl) validateAndAddRuleLocal(rule *Rule) error {
	if _, alreadyPresent := rm.rulesByName[rule.Name]; alreadyPresent {
		return errors.New("rule name must be unique")
	}

	newPatterns := make(map[string]Pattern)
	for i, pattern := range rule.Patterns {
		if err := rm.validate.Struct(pattern); err != nil {
			return err
		}

		hash := pattern.Hash()
		if existingPattern, isPresent := rm.patterns[hash]; isPresent {
			rule.Patterns[i] = existingPattern
			continue
		}

		if err := pattern.BuildPattern(); err != nil {
			return err
		}
		pattern.internalID = len(rm.patterns) + len(newPatterns)
		newPatterns[hash] = pattern
	}

	for key, value := range newPatterns {
		rm.patterns[key] = value
	}

	rm.rules[rule.ID] = *rule
	rm.rulesByName[rule.Name] = *rule

	return nil
}

func (rm *rulesManagerImpl) generateDatabase(version RowID) error {
	patterns := make([]*hyperscan.Pattern, 0, len(rm.patterns))
	for _, pattern := range rm.patterns {
		patterns = append(patterns, pattern.compiledPattern)
	}
	database, err := hyperscan.NewStreamDatabase(patterns...)
	if err != nil {
		return err
	}

	rm.databaseUpdated <- RulesDatabase{
		database:     database,
		databaseSize: len(patterns),
		version:      version,
	}
	return nil
}

func (p *Pattern) BuildPattern() error {
	if p.compiledPattern != nil {
		return nil
	}

	hp, err := hyperscan.ParsePattern(fmt.Sprintf("/%s/", p.Regex))
	if err != nil {
		return err
	}

	if p.Flags.Caseless {
		hp.Flags |= hyperscan.Caseless
	}
	if p.Flags.DotAll {
		hp.Flags |= hyperscan.DotAll
	}
	if p.Flags.MultiLine {
		hp.Flags |= hyperscan.MultiLine
	}
	if p.Flags.SingleMatch {
		hp.Flags |= hyperscan.SingleMatch
	}
	if p.Flags.Utf8Mode {
		hp.Flags |= hyperscan.Utf8Mode
	}
	if p.Flags.UnicodeProperty {
		hp.Flags |= hyperscan.UnicodeProperty
	}

	if !hp.IsValid() {
		return errors.New("can't validate the pattern")
	}

	p.compiledPattern = hp

	return nil
}

func (p Pattern) Hash() string {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%s|%v|%v|%v", p.Regex, p.Flags, p.MinOccurrences, p.MaxOccurrences)))
	return fmt.Sprintf("%x", hash.Sum(nil))
}
