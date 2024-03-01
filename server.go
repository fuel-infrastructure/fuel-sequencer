package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var (
	resultIndex = 0
	results     = []int{
		1, 1, 1, 1, 1, // batch 1
		1, 1, 1, 1, 1, // batch 2

		1, 1, 1, 1, 2, // batch 3
		1, 1, 1, 1, 2, // batch 4

		1, 1, 1, 2, 2, // batch 5
		1, 1, 1, 2, 2, // batch 6
	}
	mutex sync.Mutex
)

func main() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Define the handler function
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()

		if resultIndex == len(results) {
			resultIndex = 0
			fmt.Println()
		}

		// Write the random number as a response
		fmt.Fprintf(w, "%d", results[resultIndex])

		fmt.Println(fmt.Sprintf("SERVING %d", resultIndex))

		resultIndex += 1
	})

	// Start the server on port 8080
	fmt.Println("Server listening on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error:", err)
	}
}
