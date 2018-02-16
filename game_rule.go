package worlds

import "reflect"

const (
	GameRuleCommandBlockOutput  = "commandblockoutput"
	GameRuleDoDaylightCycle     = "dodaylightcycle"
	GameRuleDoEntityDrops       = "doentitydrops"
	GameRuleDoFireTick          = "dofiretick"
	GameRuleDoMobLoot           = "domobloot"
	GameRuleDoMobSpawning       = "domobspawning"
	GameRuleDoTileDrops         = "dotiledrops"
	GameRuleDoWeatherCycle      = "doweathercycle"
	GameRuleDrowningDamage      = "drowningdamage"
	GameRuleFallDamage          = "falldamage"
	GameRuleFireDamage          = "firedamage"
	GameRuleKeepInventory       = "keepinventory"
	GameRuleMobGriefing         = "mobgriefing"
	GameRuleNaturalRegeneration = "naturalregeneration"
	GameRulePvp                 = "pvp"
	GameRuleSendCommandFeedback = "sendcommandfeedback"
	GameRuleShowCoordinates     = "showcoordinates"
	GameRuleRandomTickSpeed     = "randomtickspeed"
	GameRuleTntExplodes         = "tntexplodes"
)

// GameRule is a struct holding a name and data of either uint32, bool or float32.
type GameRule struct {
	name  string
	value interface{}
}

// NewGameRule returns a new game rule with the given name and value.
func NewGameRule(name string, value interface{}) *GameRule {
	return &GameRule{name, value}
}

// GetName returns the name of the game rule.
func (rule *GameRule) GetName() string {
	return rule.name
}

// GetValue returns the value of the game rule.
// Game rules may hold either a uint32, a bool or a float32.
func (rule *GameRule) GetValue() interface{} {
	return rule.value
}

// SetValue sets the value of this game rule.
// Returns false if the new value does not have a type compatible with the old value.
func (rule *GameRule) SetValue(value interface{}) bool {
	if reflect.TypeOf(value).Kind() != reflect.TypeOf(rule.value).Kind() {
		return false
	}
	rule.value = value
	return true
}
