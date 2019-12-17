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
	OrderId int
}

func requestOrderHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var reqData NewOrderRequest
	err := decoder.Decode(&reqData)
	if err != nil {
		panic(err)
	}
	registry.AddOrder(reqData.Items)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%q", reqData)))
}

func getOrderStatusHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var reqData OrderRequest
	err := decoder.Decode(&reqData)
	if err != nil {
		panic(err)
	}
	var response []byte
	order, ok := registry.GetOrder(reqData.OrderId)
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
	order, ok := registry.GetOrder(reqData.OrderId)
	if !ok {
		response = []byte("Order not found!")
	} else {
		registry.CancelOrder(reqData.OrderId)
		response = []byte(order.status)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%q", response)))
}
