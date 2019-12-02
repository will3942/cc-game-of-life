package main

import (
	"fmt"
	"os"
	"strings"
	"strconv"
	"github.com/nsf/termbox-go"
)

// getKeyboardCommand sends all keys pressed on the keyboard as runes (characters) on the key chan.
// getKeyboardCommand will NOT work if termbox isn't initialised (in startControlServer)
func getKeyboardCommand(key chan<- rune) {
	for {
		event := termbox.PollEvent()
		if event.Type == termbox.EventKey {
			if event.Key == termbox.KeyCtrlC {
				StopControlServer()
				os.Exit(0)
			} else if key != nil {
				if event.Key != 0 {
					key <- rune(event.Key)
				} else if event.Ch != 0 {
					key <- event.Ch
				}
			}
		}
	}
}

// writeOutputImage writes an output pgm file by coverting the world to a series of bytes and sending this on the output IO chan
func writeOutputImage(p golParams, d distributorChans, currentTurn int, world [][]byte) {
	// Request the io goroutine to write out the image with the given filename.
	d.io.command <- ioOutput
	d.io.filename <- strings.Join([]string{strconv.Itoa(p.imageWidth), strconv.Itoa(p.imageHeight), strconv.Itoa(currentTurn)}, "x")

	for y := range world {
		for _, b := range world[y] {
			d.io.outputVal <- b
		}
	}
}


// startControlServer initialises termbox and prints basic information about the game configuration.
func startControlServer(p golParams) {
	e := termbox.Init()
	check(e)

	fmt.Println("Threads:", p.threads)
	fmt.Println("Width:", p.imageWidth)
	fmt.Println("Height:", p.imageHeight)
}

// stopControlServer closes termbox.
// If the program is terminated without closing termbox the terminal window may misbehave.
func StopControlServer() {
	termbox.Close()
}
