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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddAndGetAllRules(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	wrapper.AddCollection(Rules)

	rulesManager, err := LoadRulesManager(wrapper.Storage, "FLAG{test}")
	require.NoError(t, err)
	impl := rulesManager.(*rulesManagerImpl)
	checkVersion(t, rulesManager, impl.rulesByName["flag_out"].ID)
	checkVersion(t, rulesManager, impl.rulesByName["flag_in"].ID)
	emptyRule := Rule{Name: "empty", Color: "#fff", Enabled: true}
	emptyID, err := rulesManager.AddRule(wrapper.Context, emptyRule)
	assert.NoError(t, err)
	assert.NotNil(t, emptyID)
	checkVersion(t, rulesManager, emptyID)

	duplicateRule, err := rulesManager.AddRule(wrapper.Context, Rule{Name: "empty", Color: "#eee"})
	assert.Error(t, err)
	assert.Zero(t, duplicateRule)

	invalidPattern, err := rulesManager.AddRule(wrapper.Context, Rule{
		Name:  "invalidPattern",
		Color: "#eee",
		Patterns: []Pattern{
			{
				Regex: "invalid)",
			},
		},
	})
	assert.Error(t, err)
	assert.Zero(t, invalidPattern)

	rule1 := Rule{
		Name:  "rule1",
		Color: "#eee",
		Patterns: []Pattern{
			{
				Regex: "pattern1",
				Flags: RegexFlags{
					Caseless:        true,
					DotAll:          true,
					MultiLine:       true,
					Utf8Mode:        true,
					UnicodeProperty: true,
				},
				MinOccurrences: 1,
				MaxOccurrences: 3,
				Direction:      DirectionBoth,
			}},
		Enabled: true,
	}
	rule1ID, err := rulesManager.AddRule(wrapper.Context, rule1)
	assert.NoError(t, err)
	assert.NotNil(t, rule1ID)
	checkVersion(t, rulesManager, rule1ID)

	rule2 := Rule{
		Name:  "rule2",
		Color: "#ddd",
		Patterns: []Pattern{
			{Regex: "pattern1"},
			{Regex: "pattern2"},
		},
		Enabled: true,
	}
	rule2ID, err := rulesManager.AddRule(wrapper.Context, rule2)
	assert.NoError(t, err)
	assert.NotNil(t, rule2ID)
	checkVersion(t, rulesManager, rule2ID)

	rule3 := Rule{
		Name:  "rule3",
		Color: "#ccc",
		Patterns: []Pattern{
			{Regex: "pattern2"},
			{Regex: "pattern3"},
		},
		Enabled: true,
	}
	rule3ID, err := rulesManager.AddRule(wrapper.Context, rule3)
	assert.NoError(t, err)
	assert.NotNil(t, rule3ID)
	checkVersion(t, rulesManager, rule3ID)

	checkRule := func(expected Rule, patternIDs []int) {
		var rule Rule
		err := wrapper.Storage.Find(Rules).Context(wrapper.Context).
			Filter(OrderedDocument{{"_id", expected.ID}}).First(&rule)
		require.NoError(t, err)

		for i, id := range patternIDs {
			rule.Patterns[i].internalID = uint(id)
		}
		assert.Equal(t, expected, rule)
		assert.Equal(t, expected, impl.rules[expected.ID])
		assert.Equal(t, expected, impl.rulesByName[expected.Name])
	}

	assert.Len(t, impl.rules, 6)
	assert.Len(t, impl.rulesByName, 6)
	assert.Len(t, impl.patterns, 5)
	assert.Len(t, impl.patternsIds, 5)

	emptyRule.ID = emptyID
	rule1.ID = rule1ID
	rule2.ID = rule2ID
	rule3.ID = rule3ID

	checkRule(emptyRule, []int{})
	checkRule(rule1, []int{1})
	checkRule(rule2, []int{2, 3})
	checkRule(rule3, []int{3, 4})

	assert.Len(t, rulesManager.GetRules(), 6)
	assert.ElementsMatch(t, []Rule{impl.rulesByName["flag_out"], impl.rulesByName["flag_in"], emptyRule,
		rule1, rule2, rule3}, rulesManager.GetRules())

	wrapper.Destroy(t)
}

func TestLoadAndUpdateRules(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	wrapper.AddCollection(Rules)

	expectedIds := []RowID{NewRowID(), NewRowID(), NewRowID(), NewRowID()}
	rules := []interface{}{
		Rule{ID: expectedIds[0], Name: "rule1", Color: "#fff", Patterns: []Pattern{
			{Regex: "/pattern1/", Flags: RegexFlags{Caseless: true}, Direction: DirectionToClient, internalID: 0},
		}},
		Rule{ID: expectedIds[1], Name: "rule2", Color: "#eee", Patterns: []Pattern{
			{Regex: "/pattern2/", MinOccurrences: 1, MaxOccurrences: 3, Direction: DirectionToServer, internalID: 1},
		}},
		Rule{ID: expectedIds[2], Name: "rule3", Color: "#ddd", Patterns: []Pattern{
			{Regex: "/pattern2/", Direction: DirectionBoth, internalID: 1},
			{Regex: "/pattern3/", Flags: RegexFlags{MultiLine: true}, internalID: 2},
		}},
		Rule{ID: expectedIds[3], Name: "rule4", Color: "#ccc", Patterns: []Pattern{
			{Regex: "/pattern3/", internalID: 3},
		}},
	}
	ids, err := wrapper.Storage.Insert(Rules).Context(wrapper.Context).Many(rules)
	require.NoError(t, err)
	assert.ElementsMatch(t, expectedIds, ids)

	rulesManager, err := LoadRulesManager(wrapper.Storage, "FLAG{nope}")
	require.NoError(t, err)

	rule, isPresent := rulesManager.GetRule(NewRowID())
	assert.Zero(t, rule)
	assert.False(t, isPresent)

	for _, objRule := range rules {
		expected := objRule.(Rule)
		rule, isPresent := rulesManager.GetRule(expected.ID)
		assert.True(t, isPresent)
		assert.Equal(t, expected, rule)
	}

	updated, err := rulesManager.UpdateRule(wrapper.Context, NewRowID(), Rule{})
	assert.False(t, updated)
	assert.NoError(t, err)

	updated, err = rulesManager.UpdateRule(wrapper.Context, expectedIds[0], Rule{Name: "rule2", Color: "#fff"})
	assert.False(t, updated)
	assert.Error(t, err)

	for _, objRule := range rules {
		expected := objRule.(Rule)
		expected.Name = expected.ID.Hex()
		expected.Color = "#000"
		updated, err := rulesManager.UpdateRule(wrapper.Context, expected.ID, expected)
		assert.True(t, updated)
		assert.NoError(t, err)

		rule, isPresent := rulesManager.GetRule(expected.ID)
		assert.True(t, isPresent)
		assert.Equal(t, expected, rule)
	}

	wrapper.Destroy(t)
}

func TestFillWithMatchedRules(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	wrapper.AddCollection(Rules)

	rulesManager, err := LoadRulesManager(wrapper.Storage, "FLAG{test}")
	require.NoError(t, err)
	impl := rulesManager.(*rulesManagerImpl)
	checkVersion(t, rulesManager, impl.rulesByName["flag_out"].ID)
	checkVersion(t, rulesManager, impl.rulesByName["flag_in"].ID)

	emptyRule, err := rulesManager.AddRule(wrapper.Context, Rule{Name: "empty", Color: "#fff"})
	require.NoError(t, err)
	checkVersion(t, rulesManager, emptyRule)

	conn := &Connection{}
	rulesManager.FillWithMatchedRules(conn, map[uint][]PatternSlice{}, map[uint][]PatternSlice{})
	assert.ElementsMatch(t, []RowID{emptyRule}, conn.MatchedRules)

	filterRule, err := rulesManager.AddRule(wrapper.Context, Rule{
		Name:  "filter",
		Color: "#fff",
		Filter: Filter{
			ServicePort:   80,
			ClientAddress: "10.10.10.10",
			ClientPort:    60000,
			MinDuration:   2000,
			MaxDuration:   4000,
			MinBytes:      64,
			MaxBytes:      64,
		},
	})
	require.NoError(t, err)
	checkVersion(t, rulesManager, filterRule)
	conn = &Connection{
		SourceIP:        "10.10.10.10",
		SourcePort:      60000,
		DestinationPort: 80,
		ClientBytes:     32,
		ServerBytes:     32,
		StartedAt:       time.Now(),
		ClosedAt:        time.Now().Add(3 * time.Second),
	}
	rulesManager.FillWithMatchedRules(conn, map[uint][]PatternSlice{}, map[uint][]PatternSlice{})
	assert.ElementsMatch(t, []RowID{emptyRule, filterRule}, conn.MatchedRules)

	patternRule, err := rulesManager.AddRule(wrapper.Context, Rule{
		Name:  "pattern",
		Color: "#fff",
		Patterns: []Pattern{
			{Regex: "pattern1", Direction: DirectionToClient, MinOccurrences: 1},
			{Regex: "pattern2", Direction: DirectionToServer, MaxOccurrences: 2},
			{Regex: "pattern3", Direction: DirectionBoth, MinOccurrences: 2, MaxOccurrences: 2},
		},
	})
	require.NoError(t, err)
	checkVersion(t, rulesManager, patternRule)
	conn = &Connection{}
	rulesManager.FillWithMatchedRules(conn, map[uint][]PatternSlice{2: {{0, 0}, {0, 0}}, 3: {{0, 0}}},
		map[uint][]PatternSlice{1: {{0, 0}}, 3: {{0, 0}}})
	assert.ElementsMatch(t, []RowID{emptyRule, patternRule}, conn.MatchedRules)

	rulesManager.FillWithMatchedRules(conn, map[uint][]PatternSlice{2: {{0, 0}, {0, 0}}},
		map[uint][]PatternSlice{1: {{0, 0}}, 3: {{0, 0}, {0, 0}}})
	assert.ElementsMatch(t, []RowID{emptyRule, patternRule}, conn.MatchedRules)

	rulesManager.FillWithMatchedRules(conn, map[uint][]PatternSlice{2: {{0, 0}, {0, 0}}, 3: {{0, 0}, {0, 0}}},
		map[uint][]PatternSlice{1: {{0, 0}}})
	assert.ElementsMatch(t, []RowID{emptyRule, patternRule}, conn.MatchedRules)

	rulesManager.FillWithMatchedRules(conn, map[uint][]PatternSlice{2: {{0, 0}, {0, 0}}, 3: {{0, 0}}},
		map[uint][]PatternSlice{3: {{0, 0}}})
	assert.ElementsMatch(t, []RowID{emptyRule}, conn.MatchedRules)

	rulesManager.FillWithMatchedRules(conn, map[uint][]PatternSlice{2: {{0, 0}, {0, 0}, {0, 0}}, 3: {{0, 0}}},
		map[uint][]PatternSlice{1: {{0, 0}}, 3: {{0, 0}}})
	assert.ElementsMatch(t, []RowID{emptyRule}, conn.MatchedRules)

	rulesManager.FillWithMatchedRules(conn, map[uint][]PatternSlice{2: {{0, 0}, {0, 0}}, 3: {{0, 0}}},
		map[uint][]PatternSlice{1: {{0, 0}}, 3: {{0, 0}, {0, 0}}})
	assert.ElementsMatch(t, []RowID{emptyRule}, conn.MatchedRules)

	wrapper.Destroy(t)
}

func checkVersion(t *testing.T, rulesManager RulesManager, id RowID) {
	timeout := time.Tick(1 * time.Second)

	select {
	case database := <-rulesManager.DatabaseUpdateChannel():
		assert.Equal(t, id, database.version)
	case <-timeout:
		t.Fatal("timeout")
	}
}
