package diff

type Diff struct {
}

var (
	Build   string
	Version string
)

func New() *Diff {
	return &Diff{}
}
