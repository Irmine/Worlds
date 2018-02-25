package worlds

import "reflect"

// GameRuleName is the Minecraft name used for a game rule.
type GameRuleName string

const (
	GameRuleCommandBlockOutput  GameRuleName = "commandblockoutput"
	GameRuleDoDaylightCycle     GameRuleName = "dodaylightcycle"
	GameRuleDoEntityDrops       GameRuleName = "doentitydrops"
	GameRuleDoFireTick          GameRuleName = "dofiretick"
	GameRuleDoMobLoot           GameRuleName = "domobloot"
	GameRuleDoMobSpawning       GameRuleName = "domobspawning"
	GameRuleDoTileDrops         GameRuleName = "dotiledrops"
	GameRuleDoWeatherCycle      GameRuleName = "doweathercycle"
	GameRuleDrowningDamage      GameRuleName = "drowningdamage"
	GameRuleFallDamage          GameRuleName = "falldamage"
	GameRuleFireDamage          GameRuleName = "firedamage"
	GameRuleKeepInventory       GameRuleName = "keepinventory"
	GameRuleMobGriefing         GameRuleName = "mobgriefing"
	GameRuleNaturalRegeneration GameRuleName = "naturalregeneration"
	GameRulePvp                 GameRuleName = "pvp"
	GameRuleSendCommandFeedback GameRuleName = "sendcommandfeedback"
	GameRuleShowCoordinates     GameRuleName = "showcoordinates"
	GameRuleRandomTickSpeed     GameRuleName = "randomtickspeed"
	GameRuleTntExplodes         GameRuleName = "tntexplodes"
)

// GameRule is a struct holding a name and data of either uint32, bool or float32.
type GameRule struct {
	name  GameRuleName
	value interface{}
}

// NewGameRule returns a new game rule with the given name and value.
func NewGameRule(name GameRuleName, value interface{}) *GameRule {
	return &GameRule{name, value}
}

// GetName returns the name of the game rule.
func (rule *GameRule) GetName() GameRuleName {
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
