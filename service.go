package main

import (
	"log"
	"net/http"
)

var registry *OrderRegistry

// TODO: use golang djinn for response handling

// todo: divide over multiple files

// use simplejson for dynamic json marshaling tbh

// todo: move here hw3 automata, init their state somehow
// when order request comes, try to satisfy it
// use locks to protect automata state

func main() {
	registry = MakeOrderRegistry()
	http.HandleFunc("/", requestOrderHandler)
	http.HandleFunc("/getStatus", getOrderStatusHandler)
	http.HandleFunc("/cancel", cancelOrderHandler)
	log.Fatal(http.ListenAndServe("localhost:8001", nil))
}
