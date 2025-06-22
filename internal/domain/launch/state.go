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

func ParseState(s string) State {
	switch s {
	case "draft":
		return Draft
	case "review":
		return Review
	case "declined":
		return Declined
	case "published":
		return Published
	case "archived":
		return Archived
	default:
		return Draft
	}
}
