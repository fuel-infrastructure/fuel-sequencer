// This is a server intended for experimenting with accepting and rejecting blocks at ProcessProposal.
// It assumes a network of 5 validators that will query the server for a number, which will be 1 or 2.
// If we give them 1, they ACCEPT, and if we give them 2, they REJECT. A mutex ensures no race conditions.
// The results array allows us to give some validators a different result and observe the repercussions.

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
		1, 1, 1, 1, 1, // batch 1 (100% agree)
		1, 1, 1, 1, 1, // batch 2 (100% agree)

		1, 1, 1, 1, 2, // batch 3 (80% agree)
		1, 1, 1, 1, 2, // batch 4 (80% agree)

		1, 1, 1, 2, 2, // batch 5 (60% agree)
		1, 1, 1, 2, 2, // batch 6 (60% agree)
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
