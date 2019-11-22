package main

import (
	"fmt"
	"os"
	"github.com/nsf/termbox-go"
	"strings"
	"strconv"
)

// getKeyboardCommand sends all keys pressed on the keyboard as runes (characters) on the key chan.
// getKeyboardCommand will NOT work if termbox isn't initialised (in startControlServer)
func getKeyboardCommand(key chan<- rune) {
	for {
		event := termbox.PollEvent()
		if event.Type == termbox.EventKey {
			fmt.Println("key press")
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
func writeOutputImage(p golParams, d distributorChans, world [][]byte) {
	// Request the io goroutine to write out the image with the given filename.
	d.io.command <- ioOutput
	d.io.filename <- strings.Join([]string{strconv.Itoa(p.imageWidth), strconv.Itoa(p.imageHeight)}, "x")

	for y := range world {
		for _, b := range world[y] {
			d.io.outputVal <- b
		}
	}
}

// handleKeyPress fires when a key is pressed and is passed the params, distributor channels, key pressed, current turn and the world
func handleKeyPress(p golParams, d distributorChans, keyPressed rune, currentTurn int, world [][]byte) {
	switch keyPressed {
	case 112:
		// p pressed: pause processing until p pressed again
		paused := true
		
		fmt.Println("Current turn is ", currentTurn);
		
		for paused {
			keyPress := <- d.keyChan

			if (keyPress == 112) {
				paused = false
				fmt.Println("Continuing...");
			} else {
				handleKeyPress(p, d, keyPress, currentTurn, world)
			}
		}
	case 113:
		// q pressed: output a pgm file and then quit program
		fmt.Println("q pressed");

		writeOutputImage(p, d, world)
		
		StopControlServer()
		os.Exit(0)
	case 115:
		// s pressed: output a pgm file
		fmt.Println("s pressed");

		writeOutputImage(p, d, world)
	default:
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
