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
	var records []MachineRecord
	db.Find(&records)
	machines := make([]*Machine, 0)
	for _, record := range records {
		machines = append(machines, recordToMachine(&record))
	}
	return machines
}

func LoadOrders(db *gorm.DB) []*Order {
	var records []OrderRecord
	db.Find(&records)
	orders := make([]*Order, 0)
	for _, record := range records {
		orders = append(orders, recordToOrder(&record))
	}
	return orders
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

// UpdateAtomically updates both order and machine within a single transaction
// return non-nil error in case transaction failed
func UpdateAtomically(db *gorm.DB, o *Order, m *Machine) error {
	return db.Transaction(func(tx *gorm.DB) error {
		orderRecord := orderToRecord(o)
		if err := tx.Save(orderRecord).Error; err != nil {
			return err
		}
		machineRecord := machineToRecord(m)
		if err := tx.Save(machineRecord).Error; err != nil {
			return err
		}
		return nil
	})
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
