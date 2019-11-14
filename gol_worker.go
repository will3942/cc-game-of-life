package main

import (
  "fmt"
  "sync"
  "github.com/ChrisGora/semaphore"
)

func populateWorldWithAliveCells(world [][]byte, aliveCells []cell) [][]byte {
  for i := range aliveCells {
    cell := aliveCells[i]
    world[cell.y][cell.x] = 0xFF
  }

  return world
}

func golWorker(buffer *buffer, spaceAvailable, workAvailable semaphore.Semaphore, mutex *sync.Mutex, resBuffer *buffer, resSpaceAvailable, resWorkAvailable semaphore.Semaphore, resMutex *sync.Mutex) {
  for {
    fmt.Println("golWorker")
    // Obtain work
    workAvailable.Wait()
    mutex.Lock()

    wData := buffer.get()

    // Generate new world based on worker data
    world := createNewWorld(wData.params.imageWidth, wData.params.imageHeight)
    
    world = populateWorldWithAliveCells(world, wData.aliveCells)

    newWorld := createNewWorld(wData.params.imageWidth, wData.params.imageHeight)
    var newAliveCells []cell

    for y := wData.s.startY; y < wData.s.endY; y++ {
      for x := 0; x < wData.params.imageWidth; x++ {
        newWorld[y][x] = getNewLifeValue(world, x, y)

        if (newWorld[y][x] != world[y][x]) {
          newAliveCells = append(newAliveCells, cell{x: x, y: y})
        }
      }
    }

    // Add to response buffer
    resSpaceAvailable.Wait()
    resMutex.Lock()
    resBuffer.put(workerData{s: wData.s, aliveCells: newAliveCells, params: wData.params})
    resMutex.Unlock()
    resWorkAvailable.Post()

    // Release worker to obtain more work
    mutex.Unlock()
    spaceAvailable.Post()
  }
}