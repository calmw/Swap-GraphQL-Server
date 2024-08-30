package main

import (
	"Swap-Server/data"
	"Swap-Server/db"
	gp "Swap-Server/graphql"
	"Swap-Server/models"
	"log"
	"net/http"
	"time"
)

func main() {
	db.InitPg()
	db.InitPebble("./db/pebble_data")
	_ = db.PG.AutoMigrate(&models.SwapTrace{}, &models.Order{}, &models.Pair{}, &models.FailedTx{})

	go data.GetPairFromGraph()
	go data.GetSwapFromGraph()
	go data.GetTransferFromGraph()
	go data.GetTransferWBNBFromGraph()
	go data.GetWithDrawFromGraph()

	time.Sleep(time.Minute)
	data.SetPair()
	go data.UpdatePair()
	go data.NewOrder().Task()

	http.Handle("/graphql", gp.Handle1())
	//http.HandleFunc("/subscriptions", gp.SubscriptionsHandler)

	log.Println("GraphQL Server running on [POST]: localhost:8081/graphql")
	log.Println("GraphQL Playground running on [GET]: localhost:8081/graphql")

	log.Fatal(http.ListenAndServe(":8081", nil))
}
