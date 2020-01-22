package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type NewOrderRequest struct {
	Items []int
}

type OrderRequest struct {
	OrderId uint
}

func submitOrderHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var reqData NewOrderRequest
	err := decoder.Decode(&reqData)
	if err != nil {
		panic(err)
	}
	orderId := store.SubmitOrder(reqData.Items)
	status, err := store.ResolveOrder(orderId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Failed to create an order"))
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Order id: %d\nOrder status: %s", orderId, status)))
}

func getOrderStatusHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var reqData OrderRequest
	err := decoder.Decode(&reqData)
	if err != nil {
		panic(err)
	}
	var response []byte
	order, ok := store.GetOrder(reqData.OrderId)
	if !ok {
		response = []byte("Order not found!")
	} else {
		response = []byte(order.status)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%q", response)))
}

func cancelOrderHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var reqData OrderRequest
	err := decoder.Decode(&reqData)
	if err != nil {
		panic(err)
	}
	var response []byte
	order, ok := store.GetOrder(reqData.OrderId)
	if !ok {
		response = []byte("Order not found!")
	} else if order.status == STATUS_COMPLETED {
		response = []byte("Order has beel already completed")
	} else {
		store.CancelOrder(reqData.OrderId)
		response = []byte(order.status)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%q", response)))
}
