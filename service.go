package main

import(
    "fmt"
    "log"
    "net/http"
    "encoding/json"
)

var registry *OrderRegistry

type NewOrderRequest struct {
    Items []int 
}

type OrderRequest struct {
    OrderId int
}

// TODO: use golang djinn for response handling

// use simplejson for dynamic json marshaling tbh

func main() {
    registry = MakeOrderRegistry()
    http.HandleFunc("/", requestOrderHandler)
    http.HandleFunc("/getStatus", getOrderStatusHandler)
    http.HandleFunc("/cancel", cancelOrderHandler)
    log.Fatal(http.ListenAndServe("localhost:8001", nil))
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

type OrderRegistry struct {
    orders map[int]*Order
    nextOrderId int
}

type Order struct {
    items []int
    status string
}

func MakeOrderRegistry() *OrderRegistry {
    orders := make(map[int]*Order)
    return &OrderRegistry{orders, 1}
}

func (reg *OrderRegistry) AddOrder(items []int) {
    // todo: lock
    order := &Order{items, "pending"}
    reg.orders[reg.nextOrderId] = order
    reg.nextOrderId += 1
}

func (reg *OrderRegistry) GetOrder(orderId int) (*Order, bool) {
    val, ok := reg.orders[orderId]
    return val, ok
}

func (reg *OrderRegistry) CancelOrder(orderId int) bool {
    if order, ok := reg.orders[orderId]; ok {
        order.status = "canceled"
    }
    return false
}