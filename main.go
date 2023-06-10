package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Region struct {
	KodeBps string `json:"kode_bps"`
	NamaBps string `json:"nama_bps"`
	KodePos string `json:"kode_pos"`
	NamaPos string `json:"nama_pos"`
}

func connectDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", "root:root1234@tcp(127.0.0.1:3306)/postal_code_scrapping")
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	driver, err := migrate.New("file://postal_code_scrapper/db/migrations", "root:root1234@tcp(127.0.0.1:3306)/postal_code_scrapping")
	if err != nil {
		return err
	}

	err = driver.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func getIdsByLevel(db *sql.DB, level int) ([]string, error) {
	// Execute the SELECT query to fetch the IDs
	query := "SELECT kode_bps FROM regions_test where level = ?"
	rows, err := db.Query(query, level)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over the result rows and extract the IDs
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

func inserRegions(db *sql.DB, regions []Region, parent string, level int) error {
	stmt, err := db.Prepare("INSERT INTO regions_test (kode_bps, nama_bps, kode_pos, nama_pos, parent_id, level) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Begin a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Iterate over the persons and execute the prepared statement
	for _, region := range regions {
		_, err := tx.Stmt(stmt).Exec(region.KodeBps, region.NamaBps, region.KodePos, region.NamaPos, parent, level)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func main() {

	// Connect to the database
	db, err := connectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Println("DB connected successfully!")

	// Run the migrations
	// err = runMigrations(db)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Println("Migrations applied successfully!")

	parent := "0"
	level := 1

	insert(db, parent, level)

	ids, err := getIdsByLevel(db, level)
	if err != nil {
		log.Fatal(err)
	}

	level++
	for _, id := range ids {
		time.Sleep(3 * time.Second)
		insert(db, id, level)
	}

	ids2, err := getIdsByLevel(db, level)
	if err != nil {
		log.Fatal(err)
	}

	level++
	for _, id := range ids2 {
		time.Sleep(3 * time.Second)
		insert(db, id, level)
	}

	ids3, err := getIdsByLevel(db, level)
	if err != nil {
		log.Fatal(err)
	}

	level++
	for _, id := range ids3 {
		time.Sleep(3 * time.Second)
		insert(db, id, level)
	}

	log.Println("Completed!")
}

func insert(db *sql.DB, parent string, level int) {
	regions := getRegion(db, parent, level)

	// Insert the regions into the database
	err := inserRegions(db, regions, parent, level)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Regions inserted successfully!")
}

func getRegion(db *sql.DB, parent string, level int) []Region {
	var url string
	if level == 1 {
		url = fmt.Sprintf("https://sig.bps.go.id/rest-bridging-pos/getwilayah?level=provinsi&parent=%s", parent)
	} else if level == 2 {
		url = fmt.Sprintf("https://sig.bps.go.id/rest-bridging-pos/getwilayah?level=kabupaten&parent=%s", parent)
	} else if level == 3 {
		url = fmt.Sprintf("https://sig.bps.go.id/rest-bridging-pos/getwilayah?level=kecamatan&parent=%s", parent)
	} else {
		url = fmt.Sprintf("https://sig.bps.go.id/rest-bridging-pos/getwilayah?level=desa&parent=%s", parent)
	}

	// Create a new HTTP client
	client := http.DefaultClient
	client.Timeout = 2 * time.Second

	// Create a new GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Print the response
	fmt.Println(string(body))
	fmt.Println("=================================================")

	var regions []Region

	err = json.Unmarshal([]byte(body), &regions)
	if err != nil {
		log.Fatal(err)
	}

	var newRegions []Region
	hmap := make(map[string]struct{})
	for _, region := range regions {
		if region.KodeBps == "" && region.KodePos == "" && region.NamaBps == "" && region.NamaPos == "" {
			continue
		}

		if _, exist := hmap[region.KodeBps]; !exist {
			newRegions = append(newRegions, region)
			hmap[region.KodeBps] = struct{}{}
		}
	}

	return newRegions
}