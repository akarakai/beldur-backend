package character

type AbilityStat string

var (
	AbilityStrength     AbilityStat = "strength"
	AbilityDexterity    AbilityStat = "dexterity"
	AbilityConstitution AbilityStat = "constitution"
	AbilityIntelligence AbilityStat = "intelligence"
	AbilityWisdom       AbilityStat = "wisdom"
	AbilityCharisma     AbilityStat = "charisma"
)

type Abilities struct {
	abilityMap map[AbilityStat]int
}

func NewDefaultAbilities() Abilities {
	abilityMap := make(map[AbilityStat]int, 6)
	abilityMap[AbilityStrength] = 0
	abilityMap[AbilityDexterity] = 0
	abilityMap[AbilityConstitution] = 0
	abilityMap[AbilityIntelligence] = 0
	abilityMap[AbilityWisdom] = 0
	abilityMap[AbilityCharisma] = 0
	return Abilities{abilityMap}
}

func NewAbilities(strength, dexterity, constitution, intelligence, winsdom, charisma int) Abilities {
	abilityMap := make(map[AbilityStat]int, 6)
	abilityMap[AbilityStrength] = strength
	abilityMap[AbilityDexterity] = dexterity
	abilityMap[AbilityConstitution] = constitution
	abilityMap[AbilityIntelligence] = intelligence
	abilityMap[AbilityWisdom] = winsdom
	abilityMap[AbilityCharisma] = charisma
	return Abilities{abilityMap}
}

func (a *Abilities) Set(ability AbilityStat, val int) {
	a.abilityMap[ability] = val
}

func (a *Abilities) Adjust(ability AbilityStat, val int) {
	oldVal := a.abilityMap[ability]
	if oldVal+val < 0 {
		a.abilityMap[ability] = 0
		return
	}
	a.abilityMap[ability] += val
}

func (a *Abilities) Get(ability AbilityStat) int {
	return a.abilityMap[ability]
}
