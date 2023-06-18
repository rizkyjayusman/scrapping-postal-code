package regionservice

type Service interface {
	GetBpsCodesByLevel(level int) ([]string, error)
	InsertAll(regions []Region, parent string, level int) error
}

type Region struct {
	KodeBps  string
	NamaBps  string
	KodePos  string
	NamaPos  string
	ParentID string
	Level    int
}
