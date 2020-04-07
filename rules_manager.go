package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/flier/gohs/hyperscan"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"sync"
	"time"
)


type RegexFlags struct {
	Caseless bool `json:"caseless"`    // Set case-insensitive matching.
	DotAll bool `json:"dot_all"`    // Matching a `.` will not exclude newlines.
	MultiLine bool `json:"multi_line"`  // Set multi-line anchoring.
	SingleMatch bool `json:"single_match"`  // Set single-match only mode.
	Utf8Mode bool `json:"utf_8_mode"`         // Enable UTF-8 mode for this expression.
	UnicodeProperty bool `json:"unicode_property"`  // Enable Unicode property support for this expression
}

type Pattern struct {
	Regex string `json:"regex"`
	Flags RegexFlags `json:"flags"`
	MinOccurrences int `json:"min_occurrences"`
	MaxOccurrences int `json:"max_occurrences"`
	internalId int
	compiledPattern *hyperscan.Pattern
}

type Filter struct {
	ServicePort int
	ClientAddress string
	ClientPort int
	MinDuration int
	MaxDuration int
	MinPackets int
	MaxPackets int
	MinSize int
	MaxSize int
}

type Rule struct {
	Id string `json:"-" bson:"_id,omitempty"`
	Name string `json:"name" binding:"required,min=3" bson:"name"`
	Color string `json:"color" binding:"required,hexcolor" bson:"color"`
	Notes string `json:"notes" bson:"notes,omitempty"`
	Enabled bool `json:"enabled" bson:"enabled"`
	Patterns []Pattern `json:"patterns" binding:"required,min=1" bson:"patterns"`
	Filter Filter `json:"filter" bson:"filter,omitempty"`
	Version int64 `json:"version" bson:"version"`
}

type RulesManager struct {
	storage         Storage
	rules           map[string]Rule
	rulesByName     map[string]Rule
	ruleIndex       int
	patterns        map[string]Pattern
	mPatterns       sync.Mutex
	databaseUpdated chan interface{}
}

func NewRulesManager(storage Storage) RulesManager {
	return RulesManager{
		storage:        storage,
		rules:          make(map[string]Rule),
		patterns:       make(map[string]Pattern),
		mPatterns:      sync.Mutex{},
	}
}



func (rm RulesManager) LoadRules() error {
	var rules []Rule
	if err := rm.storage.Find(nil, Rules, NoFilters, &rules); err != nil {
		return err
	}

	var version int64
	for _, rule := range rules {
		if err := rm.validateAndAddRuleLocal(&rule); err != nil {
			log.Printf("failed to import rule %s: %s\n", rule.Name, err)
			continue
		}
		if rule.Version > version {
			version = rule.Version
		}
	}

	rm.ruleIndex = len(rules)
	return rm.generateDatabase(0)
}

func (rm RulesManager) AddRule(context context.Context, rule Rule) (string, error) {
	rm.mPatterns.Lock()

	rule.Id = UniqueKey(time.Now(), uint32(rm.ruleIndex))
	rule.Enabled = true

	if err := rm.validateAndAddRuleLocal(&rule); err != nil {
		rm.mPatterns.Unlock()
		return "", err
	}
	rm.mPatterns.Unlock()

	if _, err := rm.storage.InsertOne(context, Rules, rule); err != nil {
		return "", err
	}

	return rule.Id, rm.generateDatabase(rule.Id)
}



func (rm RulesManager) validateAndAddRuleLocal(rule *Rule) error {
	if _, alreadyPresent := rm.rulesByName[rule.Name]; alreadyPresent {
		return errors.New("rule name must be unique")
	}

	newPatterns := make(map[string]Pattern)
	for i, pattern := range rule.Patterns {
		hash := pattern.Hash()
		if existingPattern, isPresent := rm.patterns[hash]; isPresent {
			rule.Patterns[i] = existingPattern
			continue
		}
		err := pattern.BuildPattern()
		if err != nil {
			return err
		}
		pattern.internalId = len(rm.patterns) + len(newPatterns)
		newPatterns[hash] = pattern
	}

	for key, value := range newPatterns {
		rm.patterns[key] = value
	}

	rm.rules[rule.Id] = *rule
	rm.rulesByName[rule.Name] = *rule

	return nil
}

func (rm RulesManager) generateDatabase(version string) error {
	patterns := make([]*hyperscan.Pattern, len(rm.patterns))
	for _, pattern := range rm.patterns {
		patterns = append(patterns, pattern.compiledPattern)
	}
	database, err := hyperscan.NewStreamDatabase(patterns...)
	if err != nil {
		return err
	}

	rm.databaseUpdated <- database
	return nil
}


func (p Pattern) BuildPattern() error {
	if p.compiledPattern != nil {
		return nil
	}
	if p.MinOccurrences <= 0 {
		return errors.New("min_occurrences can't be lower than zero")
	}
	if p.MaxOccurrences != -1 && p.MinOccurrences < p.MinOccurrences {
		return errors.New("max_occurrences can't be lower than min_occurrences")
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

	return nil
}

func (p Pattern) Hash() string {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%s|%v|%v|%v", p.Regex, p.Flags, p.MinOccurrences, p.MaxOccurrences)))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func test() {
	user := &Pattern{Regex: "Frank"}
	b, err := json.Marshal(user)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(b))

	p, _ := hyperscan.ParsePattern("/a/")
	p1, _ := hyperscan.ParsePattern("/a/")
	fmt.Println(p1.String(), p1.Flags)
	//p1.Id = 1

	fmt.Println(*p == *p1)
	db, _ := hyperscan.NewBlockDatabase(p, p1)
	s, _ := hyperscan.NewScratch(db)
	db.Scan([]byte("Ciao"), s, onMatch, nil)




}

func onMatch(id uint, from uint64, to uint64, flags uint, context interface{}) error {
	fmt.Println(id)

	return nil
}