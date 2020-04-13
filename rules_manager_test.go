package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAddRule(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	wrapper.AddCollection(Rules)

	rulesManager := NewRulesManager(wrapper.Storage).(*rulesManagerImpl)

	checkVersion := func(id RowID) {
		timeout := time.Tick(1 * time.Second)

		select {
		case database := <-rulesManager.databaseUpdated:
			assert.Equal(t, id, database.version)
		case <-timeout:
			t.Fatal("timeout")
		}
	}

	err := rulesManager.SetFlag(wrapper.Context, "FLAG{test}")
	assert.NoError(t, err)
	checkVersion(rulesManager.rulesByName["flag"].ID)
	emptyID, err := rulesManager.AddRule(wrapper.Context, Rule{Name: "empty", Color: "#fff"})
	assert.NoError(t, err)
	assert.NotNil(t, emptyID)
	checkVersion(emptyID)

	rule1 := Rule{
		Name:  "rule1",
		Color: "#eee",
		Patterns: []Pattern{
			{Regex: "nope", Flags: RegexFlags{Caseless: true}},
		},
	}
	rule1ID, err := rulesManager.AddRule(wrapper.Context, rule1)
	assert.NoError(t, err)
	assert.NotNil(t, rule1ID)
	checkVersion(rule1ID)

	rule2 := Rule{
		Name:  "rule2",
		Color: "#ddd",
		Patterns: []Pattern{
			{Regex: "nope", Flags: RegexFlags{Caseless: true}},
			{Regex: "yep"},
		},
	}
	rule2ID, err := rulesManager.AddRule(wrapper.Context, rule2)
	assert.NoError(t, err)
	assert.NotNil(t, rule2ID)
	checkVersion(rule2ID)

	assert.Len(t, rulesManager.rules, 4)
	assert.Len(t, rulesManager.rulesByName, 4)
	assert.Len(t, rulesManager.patterns, 3)
	assert.Len(t, rulesManager.patternsIds, 3)
	assert.Equal(t, emptyID, rulesManager.rules[emptyID].ID)
	assert.Equal(t, emptyID, rulesManager.rulesByName["empty"].ID)
	assert.Len(t, rulesManager.rules[emptyID].Patterns, 0)
	assert.Equal(t, rule1ID, rulesManager.rules[rule1ID].ID)
	assert.Equal(t, rule1ID, rulesManager.rulesByName[rule1.Name].ID)
	assert.Len(t, rulesManager.rules[rule1ID].Patterns, 1)
	assert.Equal(t, 1, rulesManager.rules[rule1ID].Patterns[0].internalID)
	assert.Equal(t, rule2ID, rulesManager.rules[rule2ID].ID)
	assert.Equal(t, rule2ID, rulesManager.rulesByName[rule2.Name].ID)
	assert.Len(t, rulesManager.rules[rule2ID].Patterns, 2)
	assert.Equal(t, 1, rulesManager.rules[rule2ID].Patterns[0].internalID)
	assert.Equal(t, 2, rulesManager.rules[rule2ID].Patterns[1].internalID)

	wrapper.Destroy(t)
}
