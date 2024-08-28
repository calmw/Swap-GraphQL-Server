package db

import (
	"github.com/cockroachdb/pebble"
	"log"
)

var Pebble *pebble.DB

func InitPebble(dbDir string) {
	var err error
	Pebble, err = pebble.Open(dbDir, &pebble.Options{})
	if err != nil {
		log.Fatal(err)
	}
}
