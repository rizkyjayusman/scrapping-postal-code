package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"scrapper/component/regionservice"
	"scrapper/component/regionstore"
	"scrapper/pkg/clientbps"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func connectDB() (*sql.DB, error) {
	dbUsername := os.Getenv("DB_USERNAME")
	if dbUsername == "" {
		log.Fatal("database username is empty")
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		log.Fatal("database password is empty")
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		log.Fatal("database host is empty")
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		log.Fatal("database port is empty")
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		log.Fatal("database name is empty")
	}

	url := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUsername, dbPassword, dbHost, dbPort, dbName)
	db, err := sql.Open("mysql", url)
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

func main() {
	initiate()

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

	var regionStore *regionstore.MySQL
	regionStore, err = regionstore.New(regionstore.Config{
		DB: db,
	})

	var regionService regionservice.Service
	regionService, err = regionservice.New(regionservice.Config{
		Store: regionStore,
	})
	if err != nil {
		log.Fatal(err)
	}

	getRegionViaApiAndStore(parent, level)

	ids, err := regionService.GetBpsCodesByLevel(level)
	if err != nil {
		log.Fatal(err)
	}

	level++
	for _, id := range ids {
		time.Sleep(3 * time.Second)
		regions := getRegionViaApiAndStore(id, level)
		regionService.InsertAll(regions, parent, level)
	}

	ids2, err := regionService.GetBpsCodesByLevel(level)
	if err != nil {
		log.Fatal(err)
	}

	level++
	for _, id := range ids2 {
		time.Sleep(3 * time.Second)
		regions := getRegionViaApiAndStore(id, level)
		regionService.InsertAll(regions, parent, level)
	}

	ids3, err := regionService.GetBpsCodesByLevel(level)
	if err != nil {
		log.Fatal(err)
	}

	level++
	for _, id := range ids3 {
		time.Sleep(3 * time.Second)
		regions := getRegionViaApiAndStore(id, level)
		regionService.InsertAll(regions, parent, level)
	}

	log.Println("Completed!")
}

func initiate() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
}

func getRegionViaApiAndStore(parent string, level int) []regionservice.Region {
	regions, err := getRegion(parent, level)
	if err != nil {
		log.Println(err)
		count := 1
		for count < 3 {
			log.Printf("retry %v ...\n", count)
			regions, err = getRegion(parent, level)
			if err == nil {
				break
			}

			log.Println(err)
			time.Sleep(5 * time.Second)
			count++
		}
	}

	return regions
}

func getRegion(parent string, level int) ([]regionservice.Region, error) {
	var bpsclient clientbps.Client
	bpsclient, err := clientbps.NewHTTP()
	if err != nil {
		log.Fatal(err)
	}

	res, err := bpsclient.GetRegion(parent, level)
	if err != nil {
		log.Fatal(err)
	}

	var newRegions []regionservice.Region
	hmap := make(map[string]struct{})
	for _, region := range res {
		if region.KodeBps == "" && region.KodePos == "" && region.NamaBps == "" && region.NamaPos == "" {
			continue
		}

		if _, exist := hmap[region.KodeBps]; !exist {
			item := regionservice.Region{
				KodeBps: region.KodeBps,
				NamaBps: region.NamaBps,
				KodePos: region.KodePos,
				NamaPos: region.NamaPos,
			}
			newRegions = append(newRegions, item)
			hmap[region.KodeBps] = struct{}{}
		}
	}

	return newRegions, nil
}
