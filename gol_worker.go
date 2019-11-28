package main

import (
//  "fmt"
  "sync"
)

// Params passed to a worker
type workerParams struct {
  id int
  gameParams golParams
  seg segment
  inputChan  chan cell
  outputChan chan cell
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

func golWorker(wParams workerParams, wg *sync.WaitGroup) {
  defer wg.Done()
  // Obtain work
  //fmt.Println("worker ", wParams.id, " is processing...")

  // Obtain worker state from input channel
  wState := getNewWorkerStateFromChan(wParams.gameParams, wParams.inputChan)

  for y := wParams.seg.startY; y < wParams.seg.endY; y++ {
    for x := 0; x < wParams.gameParams.imageWidth; x++ {
      val := getNewLifeValue(wState.world, x, y)
      
      if (val != 0) {
        wParams.outputChan <- cell{x: x, y: y}
      }
    }
  }

  close(wParams.outputChan)

  //fmt.Println("worker ", wParams.id, " is done processing.")
}