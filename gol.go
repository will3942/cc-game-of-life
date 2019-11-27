package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	//"sync"
	//"github.com/ChrisGora/semaphore"
)

// type of a segment to do work on
type segment struct {
  startY  int
  endY    int
}

// type of a current state of a world
type state struct {
	world [][]byte
	aliveCells []cell
}

// Return a new world slice with a given size
func createNewWorld(width int, height int) [][]byte {
	// Create the 2D slice to store the world.
	world := make([][]byte, height)
	for i := range world {
		world[i] = make([]byte, width)
	}
	return world
}

// Return an array of world segments
func splitWorldIntoSegments(p golParams) []segment {
	segments := make([]segment, p.threads)

	heightOfASegment := p.imageHeight / p.threads

	for i := range segments {
		segments[i] = segment{
			startY: (i * heightOfASegment),
			endY:		(i * heightOfASegment) + heightOfASegment,
		}
	}

	return segments
}

// Calculates new state of the world using channels without memory sharing
func createWorkers(p golParams, segments []segment, world [][]byte) []workerParams {
	// Create array of workers
	workers := make([]workerParams, p.threads)

	// Get number of bytes in image
	numBytesInImage := (p.imageWidth * p.imageHeight)
	
	// Create input and output channels for workers
	for i := range segments {
		workers[i] = workerParams{
			id: i,
			gameParams: p,
			seg: segments[i],
			inputChan: make(chan uint8, numBytesInImage),
			outputChan: make(chan uint8, numBytesInImage),
			start: make(chan bool, 1),
			done: make(chan bool, 1),
		}

		go golWorker(workers[i])
	}

	return workers
}

// Sends world to workers
func sendWorldToWorkers(workers []workerParams, world [][]byte) {
	// Send world to workers
	for _, worker := range workers {
		worker.start <- true
	}

	for y := range world {
		for _, b := range world[y] {
			for _, worker := range workers {
				worker.inputChan <- b
			}
		}
	}
}

// Return alive cells.
func getAliveCells(p golParams, world [][]byte) []cell{
	var aliveCells []cell
	// Go through the world and append the cells that are still alive.
	for y := 0; y < p.imageHeight; y++ {
		for x := 0; x < p.imageWidth; x++ {
			if world[y][x] != 0 {
				aliveCells = append(aliveCells, cell{x: x, y: y})
			}
		}
	}
	return aliveCells
}

// Populates world with an array of alive cells
func populateWorldWithAliveCells(world [][]byte, aliveCells []cell) [][]byte {
  for i := range aliveCells {
    cell := aliveCells[i]
    world[cell.y][cell.x] = 0xFF
  }

  return world
}

// Get new state from golParams and a channel of bytes
func getNewStateFromChan(p golParams, world [][]byte, c <-chan uint8) state {
	// Create array to hold alive cells
	var aliveCells []cell

	// The channel sends the world byte by byte, in rows.
	// This populates the initial alive cells
	for y := 0; y < p.imageHeight; y++ {
		for x := 0; x < p.imageWidth; x++ {
			val := <-c

			world[y][x] = val

			if val != 0 {
				aliveCells = append(aliveCells, cell{x: x, y: y})
			}
		}
	}

	newState := state{
		world: world,
		aliveCells: aliveCells,
	}

	return newState
}

// Get new state from workers
func getNewStateFromWorkers(world [][]byte, workers []workerParams) state {
	// Get golParams from first worker
	p := workers[0].gameParams

	//world := createNewWorld(p.imageWidth, p.imageHeight)

	for y := 0; y < p.imageHeight; y++ {
		for x := 0; x < p.imageWidth; x++ {
			for _, worker := range workers {
				val := <-worker.outputChan

				world[y][x] = val
			}
		}
	}

	// Create array to hold alive cells
	aliveCells := getAliveCells(p, world)

	newState := state{
		world: world,
		aliveCells: aliveCells,
	}

	return newState
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p golParams, d distributorChans, alive chan []cell) {
	// Request the io goroutine to read in the image with the given filename.
	d.io.command <- ioInput
	d.io.filename <- strings.Join([]string{strconv.Itoa(p.imageWidth), strconv.Itoa(p.imageHeight)}, "x")

	// Create new 2D slice to store world
	world := createNewWorld(p.imageWidth, p.imageHeight)

	dState := getNewStateFromChan(p, world, d.io.inputVal)

	// Create the individual worker segments
	segments := splitWorldIntoSegments(p)

	// Create workers
	workers := createWorkers(p, segments, dState.world)

	// Calculate the new state of Game of Life after the given number of turns.
	for turns := 0; turns < p.turns; turns++ {
		// Create workers array
		fmt.Println(time.Now(), ": turn started = ", turns)

		sendWorldToWorkers(workers, dState.world)

		for _, worker := range workers {
			<-worker.done
		}

		//fmt.Println(time.Now(), ": turn construction started = ", turns)

		// Get new state from workers
		dState = getNewStateFromWorkers(world, workers)

		//fmt.Println(time.Now(), ": turn finished = ", turns)

		select {
    	case keyPress := <-d.keyChan:
        handleKeyPress(p, d, keyPress, turns, dState.world)
    	default:
        // Receiving would block if no key press occurred
    }
	}

	// Make sure that the Io has finished any output before exiting.
	d.io.command <- ioCheckIdle
	<-d.io.idle

	// Return the coordinates of cells that are still alive.
	alive <- dState.aliveCells
}
