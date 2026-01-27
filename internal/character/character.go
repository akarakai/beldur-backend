package character

import "beldur/internal/id"

// TODO probably... I dont think description is important
//type CharacterProfile struct {
//	characterId id.CharacterId
//	description string
//	avatarUrl   string
//}

type Option func(*Character)

func WithAbilities(abilities Abilities) Option {
	return func(char *Character) {
		char.abilities = abilities
	}
}

type Character struct {
	id          id.CharacterId
	name        string
	description string
	abilities   Abilities
	inventory   Inventory
}

// New should have validation? Or does the creator have full creativity right?
func New(name, description string, opt ...Option) *Character {
	c := &Character{
		name:        name,
		description: description,
		abilities:   NewDefaultAbilities(),
		inventory:   NewEmptyInventory(),
	}
	for _, o := range opt {
		o(c)
	}
	return c
}

func (c *Character) AbilityPoint(ability AbilityStat) int {
	return c.abilities.Get(ability)
}

// Should item be pointer or value?
func (c *Character) AddItem(item Item) {
	// 1. calculate how much can carry
}
