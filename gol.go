package main

import (
//	"fmt"
	"strconv"
	"strings"
	//"sync"
	//"github.com/ChrisGora/semaphore"
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

//calculate new state of all segments using buffers and return new world.
//HAS TOO MANY ARGUMENTS?
func calculateNewState(p golParams, workBufferParams bufferParams, resBufferParams bufferParams, segments []segment, aliveCells []cell) [][]byte{
	// Add worker data to buffer
	for i := range segments {
		workBufferParams.spaceAvailable.Wait()
		workBufferParams.mutex.Lock()
	
		workBufferParams.buffer.put(workerData{s: segments[i], aliveCells: aliveCells, params: p})
		
		workBufferParams.mutex.Unlock()
		workBufferParams.workAvailable.Post()

		go golWorker(workBufferParams, resBufferParams)
	}

	newWorld := createNewWorld(p.imageWidth, p.imageHeight)

	for i := 0; i < len(segments); i++ {
	// Obtain newWorld
	resBufferParams.workAvailable.Wait()
	resBufferParams.mutex.Lock()

	resData := resBufferParams.buffer.get()

	newWorld = populateWorldWithAliveCells(newWorld, resData.aliveCells)

	resBufferParams.mutex.Unlock()
	resBufferParams.spaceAvailable.Post()
	}

	return newWorld
}

//Return final cells alive.
func getFinalAlive(p golParams, world [][]byte) []cell{
	var finalAlive []cell
	// Go through the world and append the cells that are still alive.
	for y := 0; y < p.imageHeight; y++ {
		for x := 0; x < p.imageWidth; x++ {
			if world[y][x] != 0 {
				finalAlive = append(finalAlive, cell{x: x, y: y})
			}
		}
	}
	return finalAlive
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
				world[y][x] = val
				aliveCells = append(aliveCells, cell{x: x, y: y})
			}
		}
	}

	// Create the individual worker segments
	segments := splitWorldIntoSegments(p)

	workBufferParams := newBufferParams(p.threads)
	resBufferParams := newBufferParams(p.threads)

	// Calculate the new state of Game of Life after the given number of turns.
	for turns := 0; turns < p.turns; turns++ {
		newWorld := calculateNewState(p, workBufferParams, resBufferParams, segments, aliveCells)
		world = newWorld

		var newAliveCells []cell
		for y := 0; y < p.imageHeight; y++ {
			for x := 0; x < p.imageWidth; x++ {
				if world[y][x] != 0 {
					newAliveCells = append(newAliveCells, cell{x: x, y: y})
				}
			}
		}

		aliveCells = newAliveCells
		

		select {
    	case keyPress := <-d.keyChan:
        handleKeyPress(p, d, keyPress, turns, world)
    	default:
        // Receiving would block if no key press occurred
    }
	}

	finalAlive := getFinalAlive(p, world)
	
	// Make sure that the Io has finished any output before exiting.
	d.io.command <- ioCheckIdle
	<-d.io.idle

	// Return the coordinates of cells that are still alive.
	alive <- finalAlive
}
