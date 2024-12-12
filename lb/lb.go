package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	ACTIVE   = 0
	INACTIVE = 1
)

type Connection struct {
	ID     uuid.UUID `json:"id"`
	URL    string    `json:"url"`
	Status int       `json:"status"`
	MUTEX  sync.Mutex
}

type LoadBalancer struct {
	target      int
	connections []Connection
	mutex       sync.RWMutex
}

func newLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		target:      0,
		connections: []Connection{},
	}
}

func (lb *LoadBalancer) addConnection(newUrl string) {
	id := uuid.New()
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	lb.connections = append(lb.connections, Connection{
		ID:     id,
		URL:    newUrl,
		Status: INACTIVE,
	})

	fmt.Print("AFTER ADDING")
	fmt.Print(lb.connections)
}

func (lb *LoadBalancer) removeConnection(id string) {

	parsedId, err := uuid.Parse(id)

	if err != nil {
		fmt.Printf("Error parsing ID: %v\n", err)
		return
	}

	targetIdx := -1

	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	for i, connection := range lb.connections {
		if connection.ID == parsedId {
			targetIdx = i
		}
	}

	if targetIdx != -1 {
		lb.connections = append(lb.connections[:targetIdx], lb.connections[targetIdx+1:]...)
		fmt.Print("AFTER REMOVING")
		fmt.Print(lb.connections)
	}
}

func (lb *LoadBalancer) getNextTarget() (Connection, error) {
	target := lb.target

	n := len(lb.connections)
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	for i := 0; i < n; i++ {
		target = (lb.target + i) % n
		if lb.connections[target].Status == ACTIVE {
			lb.target = (target + 1) % n
			return lb.connections[target], nil
		}
	}
	return Connection{}, fmt.Errorf("ALL CONNECTIONS ARE INACTIVE")

}

func (lb *LoadBalancer) initHealthCheck(seconds int) {

	ticker := time.NewTicker(time.Second * time.Duration(seconds))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			lb.mutex.RLock()
			for idx, _ := range lb.connections {
				response := lb.healthCheck(idx)
				lb.connections[idx].MUTEX.Lock()
				if response {
					lb.connections[idx].Status = ACTIVE
				} else {
					lb.connections[idx].Status = INACTIVE
				}
				lb.connections[idx].MUTEX.Unlock()
			}
			lb.mutex.RUnlock()
		}
	}
}

func (lb *LoadBalancer) healthCheck(id int) bool {

	_, err := makeRequest(lb.connections[id].URL)
	if err != nil {
		return false
	} else {
		return true
	}
}

func makeRequest(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return []byte{}, err
	}

	return body, nil
}
