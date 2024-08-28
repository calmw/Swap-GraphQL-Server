package main

import (
	"Swap-Server/db"
	gp "Swap-Server/graphql"
	"Swap-Server/models"
	"log"
	"net/http"
)

func main() {
	db.InitPg()
	db.InitPebble("./db/pebble_data")
	_ = db.PG.AutoMigrate(&models.SwapTrace{}, &models.Order{}, &models.Pair{})

	http.Handle("/graphql", gp.Handle1())
	http.HandleFunc("/subscriptions", gp.SubscriptionsHandler)

	log.Println("GraphQL Server running on [POST]: localhost:8081/graphql")
	log.Println("GraphQL Playground running on [GET]: localhost:8081/graphql")

	log.Fatal(http.ListenAndServe(":8081", nil))
}
