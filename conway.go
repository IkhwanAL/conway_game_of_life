package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
)

func interrupt(notify chan os.Signal, screen tcell.Screen) {
	signal.Notify(notify, os.Interrupt)

	go func() {
		<-notify // Receive
		screen.Fini()
		os.Exit(0)
	}()
}

func main() {

	screen, err := tcell.NewScreen()

	if err != nil {
		log.Fatal(err)
	}

	var notifyChan chan os.Signal = make(chan os.Signal, 1)

	interrupt(notifyChan, screen)

	if err = screen.Init(); err != nil {
		log.Fatal(err)
	}

	screen.Clear()

	maxWidth := 60
	maxHeight := 60

	emptyGrid := rune(' ')
	fillGrid := rune('█')

	population := 5
	generation := 0

	populationText := fmt.Sprintf("Population: %d", population)
	for i, r := range populationText {
		screen.SetContent(i, 0, r, nil, tcell.StyleDefault)
	}

	generationText := fmt.Sprintf("Generation: %d", generation)
	for i, r := range generationText {
		screen.SetContent(i, 1, r, nil, tcell.StyleDefault)
	}

	heightTakeOver := 2

	glider := [][2]int{
		{5, 5}, {6, 6}, {6, 7}, {7, 5}, {7, 6},
	}

	for _, cell := range glider {
		x, y := cell[0], cell[1]
		screen.SetContent(x, y, fillGrid, nil, tcell.StyleDefault)
	}

	screen.Show()
	screen.EnableMouse()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	topLeft := []int{-1, -1}
	left := []int{-1, 0}
	bottomLeft := []int{-1, 1}

	top := []int{0, -1}
	bottom := []int{0, 1}

	topRight := []int{1, -1}
	right := []int{1, 0}
	bottomRight := []int{1, 1}

	directions := [][]int{
		topLeft,
		left,
		bottomLeft,
		top,
		bottom,
		right,
		topRight,
		bottomRight,
	}

	for {
		select {
		case <-ticker.C:
			nextState := make(map[[2]int]rune)

			// Loop All Rectangle Grid
			for h := 0; h < maxHeight; h++ {
				for w := 0; w < maxWidth; w++ {
					livingCell := 0
					for dir := range directions {
						x := (w + directions[dir][0] + maxWidth) % maxWidth
						y := (h + heightTakeOver + directions[dir][1] + maxHeight) % maxHeight

						content, _, _, _ := screen.GetContent(x, y)
						if content == fillGrid {
							livingCell++
						}
					}

					currentContent, _, _, _ := screen.GetContent(w, h+heightTakeOver)

					// If Living Neighbor From Current Cell is less than 2 or more than 3
					// Current Cell Dead
					if currentContent == fillGrid && (livingCell < 2 || livingCell > 3) {
						nextState[[2]int{w, h + heightTakeOver}] = emptyGrid
						population--
					} else if currentContent == emptyGrid && (livingCell == 3) {
						// If Living Neighbor From Dead Cell is equal 3
						// Current Cell Alive
						nextState[[2]int{w, h + heightTakeOver}] = fillGrid
						population++
					}
				}
			}
			for pos, state := range nextState {
				screen.SetContent(pos[0], pos[1], state, nil, tcell.StyleDefault)
			}

			generation++
			populationChars := strconv.Itoa(population)

			for i, r := range populationChars {
				screen.SetContent(12+i, 0, rune(r), nil, tcell.StyleDefault)
			}

			genChars := strconv.Itoa(generation)
			for i, r := range genChars {
				screen.SetContent(12+i, 1, rune(r), nil, tcell.StyleDefault)
			}

			screen.Show()
		default:
			if screen.HasPendingEvent() {
				ev := screen.PollEvent()
				switch ev := ev.(type) {
				case *tcell.EventKey:
					if ev.Key() == tcell.KeyEscape || ev.Rune() == 'q' {
						screen.Fini()
						os.Exit(0)
					}
				case *tcell.EventMouse:
					x, y := ev.Position()

					buttonAction := ev.Buttons()

					if buttonAction == tcell.ButtonPrimary {
						screen.SetContent(x, y, fillGrid, nil, tcell.StyleDefault)
						population++
						screen.Show()
					}

					if buttonAction == tcell.ButtonSecondary {
						screen.SetContent(x, y, emptyGrid, nil, tcell.StyleDefault)
						population--
						screen.Show()
					}
				}
			}
		}
	}
}
