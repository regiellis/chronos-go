package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/regiellis/chronos-go/chronos"
	"github.com/regiellis/chronos-go/db"
)

var projects = []string{"Chronos", "Apollo", "Hermes", "Zeus"}
var clients = []string{"Acme Corp", "Globex", "Initech", "Umbrella"}
var tasks = []string{"Coding", "Review", "Meeting", "Testing"}
var descriptions = []string{
	"Implemented feature X",
	"Fixed bug Y",
	"Reviewed PR",
	"Team sync",
	"Wrote documentation",
	"Refactored module",
}

func randomEntry() *chronos.Entry {
	return &chronos.Entry{
		Project:     projects[rand.Intn(len(projects))],
		Client:      clients[rand.Intn(len(clients))],
		Task:        tasks[rand.Intn(len(tasks))],
		Description: descriptions[rand.Intn(len(descriptions))],
		Duration:    int64(rand.Intn(120) + 15),                                   // 15-135 min
		EntryTime:   time.Now().Add(-time.Duration(rand.Intn(30*24)) * time.Hour), // within last 30 days
		CreatedAt:   time.Now(),
		Billable:    rand.Intn(2) == 0,
		Rate:        float64(rand.Intn(100) + 50),
		Invoiced:    rand.Intn(2) == 0,
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	dbPath := "chronos.db"
	dbStore, err := db.NewStore(dbPath)
	if err != nil {
		panic(err)
	}
	if err := dbStore.InitSchema(); err != nil {
		panic(err)
	}
	count := 50 // Number of entries to insert
	for i := 0; i < count; i++ {
		entry := randomEntry()
		err := dbStore.AddEntry(entry)
		if err != nil {
			fmt.Printf("Failed to insert entry %d: %v\n", i, err)
		}
	}
	fmt.Printf("Inserted %d random entries.\n", count)
}
