package store

import (
	"fmt"
	"sync"
	"time"
)


type StoreData struct{
	Data string
	ch   chan bool
}

var store map[string]StoreData
var mutex = &sync.Mutex{}

type Store map[string]StoreData

func GetStore() Store{
	if store == nil{
		store = map[string]StoreData{}
	} 
	return store
}

func(s Store) Set(key string, value string) {
	mutex.Lock()
	defer mutex.Unlock()
	storeData := store[key]

	if storeData != (StoreData{}) {
		storeData.Data = value
	}
	store[key] = StoreData{
		Data : value,
	}
	fmt.Println("Data set", key, value)
}

func(s Store) SetWithExpiry(key string, value string, duration int64) {
	mutex.Lock()
	storeData, exists := store[key]
	if exists && storeData.ch != nil {
		storeData.ch <- true
	}
	
	storeData = StoreData{
		Data: value,
		ch:   make(chan bool),
	}
	store[key] = storeData
	mutex.Unlock()

	fmt.Println("Setting value with duration:", duration)

	go func() {
		// Wait for duration or termination signal
		select {
		case <-storeData.ch:
			fmt.Println("Termination signal received for key:", key)
		case <-time.After(time.Millisecond * time.Duration(duration)):
			mutex.Lock()
			delete(store, key)
			mutex.Unlock()
			fmt.Println("Expired and deleted key:", key)
		}
	}()
}

func(s Store) Get(key string) string {
	mutex.Lock()
	ans := store[key].Data
	mutex.Unlock()
	return ans
}