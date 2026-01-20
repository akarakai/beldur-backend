package player

const (
	UsernameMaxCharacters = 20
)

type Player struct {
	Id   int
	Name string
}

func New(name string) (*Player, error) {
	if err := validateUsername(name); err != nil {
		return nil, err
	}

	return &Player{
		Name: name,
	}, nil
}

func (p Player) String() string {
	return p.Name
}

// TODO better validation, remove also duplication ?
func validateUsername(value string) error {
	if len(value) > UsernameMaxCharacters {
		return ErrInvalidPlayerName
	}
	return nil
}
