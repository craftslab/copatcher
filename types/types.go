package types

type Differ struct {
	Image1   string `json:"Image1"`
	Image2   string `json:"Image2"`
	DiffType string `json:"DiffType"`
	Diff     Diff   `json:"Diff"`
}

type Diff struct {
	Packages1 []Package  `json:"Packages1"`
	Packages2 []Package  `json:"Packages2"`
	InfoDiff  []InfoDiff `json:"InfoDiff"`
}

type Package struct {
	Name    string `json:"Name"`
	Version string `json:"Version"`
	Size    int    `json:"Size"`
}

type InfoDiff struct {
	Package string `json:"Package"`
	Info1   Info   `json:"Info1"`
	Info2   Info   `json:"Info2"`
}

type Info struct {
	Version string `json:"Version"`
	Size    int    `json:"Size"`
}

func New() *[]Differ {
	return &[]Differ{}
}
