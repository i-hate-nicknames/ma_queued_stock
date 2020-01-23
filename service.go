package main

import (
	"log"
	"net/http"
)

// todo: pass store around as an argument to handlers instead of a global variable
var store *Store

func main() {
	db := DbConnect()
	Migrate(db)
	store = LoadStore(db)
	// todo: implement reading machine settings from command line
	// or better add methods to create/update machines at runtime
	if len(store.machines) == 0 {
		store.generateMachines()
	}
	http.HandleFunc("/submit", submitOrderHandler)
	http.HandleFunc("/getStatus", getOrderStatusHandler)
	http.HandleFunc("/cancel", cancelOrderHandler)
	log.Fatal(http.ListenAndServe("localhost:8001", nil))
}
