package launch

type State int

const (
	Draft State = iota
	Review
	Declined
	Published
	Archived
)

func (s State) String() string {
	return [...]string{"draft", "review", "declined", "published", "archived"}[s]
}

func ParseState(s string) (State, error) {
	switch s {
	case "draft":
		return Draft, nil
	case "review":
		return Review, nil
	case "declined":
		return Declined, nil
	case "published":
		return Published, nil
	case "archived":
		return Archived, nil
	default:
		return Draft, nil
	}
}
