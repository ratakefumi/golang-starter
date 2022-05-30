package db

import (
	"log"

	scribble "github.com/nanobox-io/golang-scribble"
)

type ScribleDB interface {
	Query() *scribble.Driver
}

type scribleDB struct {
	db *scribble.Driver
}

func NewScribleClient() ScribleDB {
	db, err := scribble.New("temp/db", nil)
	if err != nil {
		log.Println("Error", err)
	}

	return &scribleDB{
		db: db,
	}
}

func (db *scribleDB) Query() *scribble.Driver {
	return db.db
}
