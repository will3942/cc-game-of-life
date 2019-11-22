package main

import (
//  "fmt"
  "sync"
  "github.com/ChrisGora/semaphore"
)


type bufferParams struct{
  buffer *buffer
  spaceAvailable semaphore.Semaphore
  workAvailable semaphore.Semaphore
  mutex *sync.Mutex
}

func newBufferParams(threads int) bufferParams{
  return bufferParams{
    buffer: &buffer{
      b:     make([]workerData, threads),
      size:  threads,
      read:  0,
      write: 0,
    },
    spaceAvailable: semaphore.Init(threads, threads),
    workAvailable : semaphore.Init(threads, 0),
    mutex: &sync.Mutex{},
  }
}


func populateWorldWithAliveCells(world [][]byte, aliveCells []cell) [][]byte {
  for i := range aliveCells {
    cell := aliveCells[i]
    world[cell.y][cell.x] = 0xFF
  }
  return world
}


func golWorker(workBufferParams bufferParams, responseBufferParams bufferParams) {
  // Obtain work
  workBufferParams.workAvailable.Wait()
  workBufferParams.mutex.Lock()

  wData := workBufferParams.buffer.get()

  // Generate new world based on worker data
  world := createNewWorld(wData.params.imageWidth, wData.params.imageHeight)
  
  world = populateWorldWithAliveCells(world, wData.aliveCells)

  newWorld := createNewWorld(wData.params.imageWidth, wData.params.imageHeight)
  var newAliveCells []cell

  for y := wData.s.startY; y < wData.s.endY; y++ {
    for x := 0; x < wData.params.imageWidth; x++ {
      newWorld[y][x] = getNewLifeValue(world, x, y)
      if (newWorld[y][x] != 0) {
        newAliveCells = append(newAliveCells, cell{x: x, y: y})
      }
    }
  }

  // Add to response buffer
  responseBufferParams.spaceAvailable.Wait()
  responseBufferParams.mutex.Lock()
  responseBufferParams.buffer.put(workerData{s: wData.s, aliveCells: newAliveCells, params: wData.params})
  responseBufferParams.mutex.Unlock()
  responseBufferParams.workAvailable.Post()

  // Release worker to obtain more work
  workBufferParams.mutex.Unlock()
  workBufferParams.spaceAvailable.Post()
}