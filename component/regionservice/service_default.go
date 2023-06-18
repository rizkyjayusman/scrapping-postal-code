package regionservice

import (
	"scrapper/component/regionstore"
)

type Config struct {
	Store regionstore.Store
}

type Default struct {
	Config Config
}

func New(cfg Config) (*Default, error) {
	e := &Default{
		Config: cfg,
	}
	return e, nil
}

func (e *Default) GetBpsCodesByLevel(level int) ([]string, error) {
	return e.Config.Store.GetBpsCodesByLevel(level)
}

func (e *Default) InsertAll(regions []Region, parent string, level int) error {
	var regionStores []regionstore.Region
	for _, region := range regions {
		regionStores = append(regionStores, regionstore.Region(region))
	}
	return e.Config.Store.InsertAll(regionStores, parent, level)
}
