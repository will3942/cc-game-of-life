package main

import (
//	"fmt"
	"strconv"
	"strings"
	"sync"
	"github.com/ChrisGora/semaphore"
)

type segment struct {
	startY	int
	endY		int
}

// Return the life value of a neighbour
func getNeighbourLifeValue(world [][]byte, x int, y int) byte {
	worldHeight := cap(world)
	worldWidth := cap(world[0])

	wrappedX := 0
	wrappedY := 0

	if (x > (worldWidth - 1)) {
		wrappedX = (x - worldWidth)
	} else if (x < 0) {
		wrappedX = (worldWidth + x)
	} else {
		wrappedX = x
	}

	if (y > (worldHeight - 1)) {
		wrappedY = (y - worldHeight)
	} else if (y < 0) {
		wrappedY = (worldHeight + y)
	} else {
		wrappedY = y
	}

	return world[wrappedY][wrappedX]
}

// Return a number of living neighbours for a given x,y coordinate
func getNumLiveNeighbours(world [][]byte, x int, y int) int {
	numLiveNeighbours := 0

	neighbourOffsets := [][]int{
	  {-1, -1},
	  {-1, 0},
	  {-1, 1},
	  {0, -1},
	  {0, 1},
	  {1, -1},
	  {1, 0},
	  {1, 1},
	}

	for i := range neighbourOffsets {
		offsettedX := x + neighbourOffsets[i][0]
		offsettedY := y + neighbourOffsets[i][1]
		
		if (getNeighbourLifeValue(world, offsettedX, offsettedY) != 0) {
			numLiveNeighbours++
		}
	}

	return numLiveNeighbours
}

// Return a new life value for a given world and coordinates
func getNewLifeValue(world [][]byte, x int, y int) byte {
	initialLifeValue := world[y][x]

	numLiveNeighbours := getNumLiveNeighbours(world, x, y)

	if (initialLifeValue == 0) {
		if (numLiveNeighbours != 3) {
			return initialLifeValue
		}
		
		return (initialLifeValue ^ 0xFF)
	} else {
		if (numLiveNeighbours < 2) {
			return (initialLifeValue ^ 0xFF)
		}

		if (numLiveNeighbours <= 3) {
			return initialLifeValue
		}
		
		return (initialLifeValue ^ 0xFF)
	}
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

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p golParams, d distributorChans, alive chan []cell) {

	// Create the 2D slice to store the world.
	world := createNewWorld(p.imageWidth, p.imageHeight)

	// Request the io goroutine to read in the image with the given filename.
	d.io.command <- ioInput
	d.io.filename <- strings.Join([]string{strconv.Itoa(p.imageWidth), strconv.Itoa(p.imageHeight)}, "x")

	// Create array to hold slice
	var aliveCells []cell

	// The io goroutine sends the requested image byte by byte, in rows.
	for y := 0; y < p.imageHeight; y++ {
		for x := 0; x < p.imageWidth; x++ {
			val := <-d.io.inputVal
			if val != 0 {
				//fmt.Println("Alive cell at", x, y)
				world[y][x] = val

				aliveCells = append(aliveCells, cell{x: x, y: y})
			}
		}
	}

	// Create the individual worker segments
	segments := splitWorldIntoSegments(p)

	// Initialise work buffer and response buffer
	buffer := newBuffer(p.threads)
	mutex := &sync.Mutex{}

	spaceAvailable := semaphore.Init(1, 1)
	workAvailable := semaphore.Init(1, 0)

	resBuffer := newBuffer(p.threads)
	resMutex := &sync.Mutex{}

	resSpaceAvailable := semaphore.Init(1, 1)
	resWorkAvailable := semaphore.Init(1, 0)

	// Test: Add worker data to buffer
	for i := range segments {
		spaceAvailable.Wait()
		mutex.Lock()
		
		buffer.put(workerData{s: segments[i], aliveCells: aliveCells, params: p})
		
		mutex.Unlock()
		workAvailable.Post()
	}

	go golWorker(&buffer, spaceAvailable, workAvailable, mutex, &resBuffer, resSpaceAvailable, resWorkAvailable, resMutex)

	// Calculate the new state of Game of Life after the given number of turns.
	for turns := 0; turns < p.turns; turns++ {
		newWorld := createNewWorld(p.imageWidth, p.imageHeight)

		for y := 0; y < p.imageHeight; y++ {
			for x := 0; x < p.imageWidth; x++ {
				newWorld[y][x] = getNewLifeValue(world, x, y)
			}
		}

		world = newWorld
	}

	// Create an empty slice to store coordinates of cells that are still alive after p.turns are done.
	var finalAlive []cell
	// Go through the world and append the cells that are still alive.
	for y := 0; y < p.imageHeight; y++ {
		for x := 0; x < p.imageWidth; x++ {
			if world[y][x] != 0 {
				finalAlive = append(finalAlive, cell{x: x, y: y})
			}
		}
	}

	// Make sure that the Io has finished any output before exiting.
	d.io.command <- ioCheckIdle
	<-d.io.idle

	// Return the coordinates of cells that are still alive.
	alive <- finalAlive
}
