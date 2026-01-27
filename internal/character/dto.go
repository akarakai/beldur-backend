package character

type CreateCharacterRequest struct {
	Name        string     `json:"name" validate:"required"`
	Description string     `json:"description" validate:"required"`
	Abilities   AbilityDto `json:"abilities" validate:"required"`
}

type AbilityDto struct {
	Strength     int `json:"strength" validate:"required"`
	Dexterity    int `json:"dexterity" validate:"required"`
	Constitution int `json:"constitution" validate:"required"`
	Intelligence int `json:"intelligence" validate:"required"`
	Wisdom       int `json:"wisdom" validate:"required"`
	Charisma     int `json:"charisma" validate:"required"`
}

type CreateCharacterResponse struct {
	Id          int        `json:"character_id"`
	CampaignId  int        `json:"campaign_id"`
	Name        string     `json:"name" validate:"required"`
	Description string     `json:"description" validate:"required"`
	Abilities   AbilityDto `json:"abilities" validate:"required"`
}
