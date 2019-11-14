package main

import (
  "fmt"
  "sync"
  "github.com/ChrisGora/semaphore"
)

func golWorker(buffer *buffer, spaceAvailable, workAvailable semaphore.Semaphore, mutex *sync.Mutex, resBuffer *buffer, resSpaceAvailable, resWorkAvailable semaphore.Semaphore, resMutex *sync.Mutex) {
  for {
    fmt.Println("golWorker")
    // Obtain work
    workAvailable.Wait()
    mutex.Lock()

    wData := buffer.get()
    
    fmt.Println(wData)

    // Release worker to obtain more work
    mutex.Unlock()
    spaceAvailable.Post()
  }
}