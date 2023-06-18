package regionstore

import (
	"database/sql"
	"fmt"
)

const (
	TableName = "regions_test"
)

type Config struct {
	DB *sql.DB
}

type MySQL struct {
	Config Config
}

func New(cfg Config) (*MySQL, error) {
	e := &MySQL{
		Config: cfg,
	}
	return e, nil
}

func (e *MySQL) GetBpsCodesByLevel(level int) ([]string, error) {
	query := fmt.Sprintf("SELECT kode_bps FROM %s where level = ?", TableName)
	rows, err := e.Config.DB.Query(query, level)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

func (e *MySQL) InsertAll(regions []Region, parent string, level int) error {
	query := fmt.Sprintf("INSERT INTO %s (kode_bps, nama_bps, kode_pos, nama_pos, parent_id, level) VALUES (?, ?, ?, ?, ?, ?)", TableName)
	stmt, err := e.Config.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	tx, err := e.Config.DB.Begin()
	if err != nil {
		return err
	}

	for _, region := range regions {
		_, err := tx.Stmt(stmt).Exec(region.KodeBps, region.NamaBps, region.KodePos, region.NamaPos, parent, level)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
