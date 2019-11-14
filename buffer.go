package main

import (
//"fmt"
)

type workerData struct {
  s          segment
  aliveCells []cell
  params     golParams
}

type buffer struct {
  b           []workerData
  size, read, write int
}

func newBuffer(size int) buffer {
  return buffer{
    b:     make([]workerData, size),
    size:  size,
    read:  0,
    write: 0,
  }
}

func (buffer *buffer) get() workerData {
  wData := buffer.b[buffer.read]
  //fmt.Println("Get\t", wData, "\t", buffer)
  buffer.read = (buffer.read + 1) % len(buffer.b)
  return wData
}

func (buffer *buffer) put(wData workerData) {
  buffer.b[buffer.write] = wData
  //fmt.Println("Put\t", wData, "\t", buffer)
  //fmt.Println("Put\t", wData)
  buffer.write = (buffer.write + 1) % len(buffer.b)
}