package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type MachineRecord struct {
	ID    uint `gorm:"primary_key"`
	Items string
}

type OrderRecord struct {
	ID           uint `gorm:"primary_key"`
	Items        string
	FetchedItems string
	Status       string
}

func DbConnect() *gorm.DB {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		fmt.Println(err.Error())
		panic("failed to connect database")
	}
	return db
}

func LoadMachines(db *gorm.DB) []*Machine {
	ms := make([]*Machine, 0)
	return ms
}

func LoadOrders(db *gorm.DB) []*Order {
	os := make([]*Order, 0)
	return os
}

// SaveMachine saves given machine to db. If order.ID is 0,
// this will create a new order. Return orderID
func SaveMachine(db *gorm.DB, m *Machine) uint {
	record := machineToRecord(m)
	db.Save(record)
	return record.ID
}

// SaveOrder saves given order to db. If order.ID is 0,
// this will create a new order. Return orderID
func SaveOrder(db *gorm.DB, o *Order) uint {
	record := orderToRecord(o)
	db.Save(record)
	return record.ID
}

func Migrate(db *gorm.DB) {
	db.AutoMigrate(&MachineRecord{}, &OrderRecord{})
}

func machineToRecord(m *Machine) *MachineRecord {
	items, err := json.Marshal(m.GetAllItems())
	if err != nil {
		log.Println("Malformed machine items: %q", items)
		return nil
	}
	return &MachineRecord{
		ID:    m.ID,
		Items: string(items),
	}
}

func recordToMachine(mr *MachineRecord) *Machine {
	var items []int
	err := json.Unmarshal([]byte(mr.Items), &items)
	if err != nil {
		log.Println("Malformed data in the database: " + mr.Items)
		return nil
	}
	return MakeMachine(mr.ID, items)
}

func orderToRecord(o *Order) *OrderRecord {
	items, err := json.Marshal(o.items)
	if err != nil {
		log.Println("Malformed order: %q", o.items)
		return nil
	}
	fetchedItems, err := json.Marshal(o.fetchedItems)
	if err != nil {
		log.Println("Malformed order: %q", o.items)
		return nil
	}
	return &OrderRecord{
		ID:           o.ID,
		Items:        string(items),
		FetchedItems: string(fetchedItems),
		Status:       o.status,
	}
}

func recordToOrder(or *OrderRecord) *Order {
	var items, fetchedItems []int
	err := json.Unmarshal([]byte(or.FetchedItems), &fetchedItems)
	if err != nil {
		log.Println("Malformed data in the database: " + or.FetchedItems)
		return nil
	}
	err = json.Unmarshal([]byte(or.Items), &items)
	if err != nil {
		log.Println("Malformed data in the database: " + or.Items)
		return nil
	}
	return &Order{
		ID:           or.ID,
		items:        items,
		fetchedItems: fetchedItems,
		status:       or.Status,
	}
}
