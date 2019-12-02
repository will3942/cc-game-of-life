package main

import (
  "fmt"
  "time"
)

// Structure to hold all the channels required by the ticker thread
type tickerChans struct {
  numAliveCells chan int
  pause    chan bool
  stop chan bool
}

// Prints number of alive cells every 2 seconds
func ticker(p golParams, tChans tickerChans,) {
  ticker := time.NewTicker(2 * time.Second)

  isPaused := false
  numAliveCells := 0

  // Handle receiving number of alive cells and printing in a goroutine
  go func() {
    for {
      select {
      case <-tChans.stop:
        // Stop ticker
        ticker.Stop()

        return
      case pause := <-tChans.pause:
        // Pause printing
        switch pause {
        case true:
          isPaused = true
        case false:
          isPaused = false
        }
      case n := <-tChans.numAliveCells:
        numAliveCells = n
      case t := <-ticker.C:
        if (!isPaused) {
          fmt.Println("There are ", numAliveCells, " still alive at ", t)
        }
      }
    }
  }()
}