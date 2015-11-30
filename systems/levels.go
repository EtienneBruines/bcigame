package systems

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/paked/engi/ecs"
)

type Tile uint8

const (
	_               = iota
	TilePlayer Tile = iota
	TileWall
	TileBlank
	TileGoal
	TileRoute
	TileHiddenError
	TileError
)

func (t Tile) String() string {
	switch t {
	case TilePlayer:
		return "X"
	case TileWall:
		return "-"
	case TileBlank:
		return " "
	case TileGoal:
		return "G"
	case TileRoute:
		return "+"
	case TileError:
		return "E"
	case TileHiddenError:
		return "H"
	default:
		return ""
	}
}

type Level struct {
	ID           int
	Name         string
	Width        int
	Height       int
	Grid         [][]Tile
	GridEntities [][]*ecs.Entity

	PlayerX, PlayerY int
}

func NewLevel() Level {
	idCounter++
	return Level{ID: idCounter}
}

func (l *Level) IsAvailable(x, y int) bool {
	if x < 0 || x >= l.Width || y < 0 || y >= l.Height {
		return false
	}

	return l.Grid[y][x] != TileWall
}

func (l *Level) Copy() Level {
	lvl := Level{
		ID:      l.ID,
		Name:    l.Name,
		Width:   l.Width,
		Height:  l.Height,
		PlayerX: l.PlayerX,
		PlayerY: l.PlayerY,
	}

	lvl.Grid = make([][]Tile, len(l.Grid))
	for rowIndex, row := range l.Grid {
		lvl.Grid[rowIndex] = make([]Tile, len(row))
		for cellIndex, cell := range row {
			lvl.Grid[rowIndex][cellIndex] = cell
		}
	}

	return lvl
}

var emptyLevel = NewLevel()
var idCounter = 0

func LoadLevels(dir string) (levels []Level) {
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	var files []string

	for _, info := range infos {
		if !info.IsDir() {
			ext := filepath.Ext(info.Name())
			if ext[1:] == "maze" {
				files = append(files, filepath.Join(dir, info.Name()))
			}
		}
	}

	for _, file := range files {
		lvl := NewLevel()

		b, err := ioutil.ReadFile(file)
		if err != nil {
			continue // with other files
		}

		content := string(b)

		lines := strings.Split(content, "\n")
		lvl.Height = len(lines)

		for lineIndex, line := range lines {
			if lineIndex == 0 {
				lvl.Name = line
				continue // with the actual maze
			}
			if len(line) > lvl.Width {
				lvl.Width = len(line)
			}

			gameRow := make([]Tile, len(line))
			for index, char := range line {
				switch char {
				case 'X':
					gameRow[index] = TilePlayer
				case '-':
					gameRow[index] = TileWall
				case 'G':
					gameRow[index] = TileGoal
				case ' ':
					gameRow[index] = TileBlank
				case '+':
					gameRow[index] = TileRoute
				case 'E':
					gameRow[index] = TileError
				case 'H':
					gameRow[index] = TileHiddenError
				}
			}
			lvl.Grid = append(lvl.Grid, gameRow)
		}

		levels = append(levels, lvl)
	}
	return
}

func (l *Level) Save(file string) {
	// Find goal
	var goalX, goalY int
	for rowIndex, row := range l.Grid {
		for cellIndex, cell := range row {
			if cell == TileGoal {
				goalX, goalY = cellIndex, rowIndex
			}
		}
	}

	// Add route
	l.PlayerX = l.Width - 2
	l.PlayerY = l.Height - 2
	route := computeRoute(l, l.PlayerX, l.PlayerY, goalX, goalY)
	for _, action := range route {
		l.Grid[l.PlayerY][l.PlayerX] = TileRoute
		switch action {
		case ActionUp:
			l.PlayerY--
		case ActionDown:
			l.PlayerY++
		case ActionLeft:
			l.PlayerX--
		case ActionRight:
			l.PlayerX++
		}
	}

	l.Grid[l.Height-2][l.Width-2] = TilePlayer

	// Actually saving
	var buffer bytes.Buffer
	for _, row := range l.Grid {
		for _, cell := range row {
			buffer.WriteString(cell.String())
		}
		buffer.WriteString("\n")
	}

	fileBuffer, err := os.Create(file)
	if err != nil {
		log.Println(err)
	}
	defer fileBuffer.Close()

	buffer.WriteTo(fileBuffer)

}

func NewRandomLevel(minWidth, maxWidth int, minHeight, maxHeight int) Level {
	lvl := NewLevel()
	lvl.Width = minWidth + rand.Intn(maxWidth-minWidth)
	lvl.Height = minHeight + rand.Intn(maxHeight-minHeight)

	if lvl.Width%2 == 0 {
		lvl.Width++
	}
	if lvl.Height%2 == 0 {
		lvl.Height++
	}

	// Initialize grid
	lvl.Grid = make([][]Tile, lvl.Height)
	for rowIndex := range lvl.Grid {
		lvl.Grid[rowIndex] = make([]Tile, lvl.Width)
		for cellIndex := range lvl.Grid[rowIndex] {
			lvl.Grid[rowIndex][cellIndex] = TileBlank
		}
	}

	lvl.Name = fmt.Sprintf("Random %d by %d", lvl.Width, lvl.Height)

	// Create walls at the border
	for row := 0; row < lvl.Height; row += lvl.Height - 1 {
		for cell := 0; cell < lvl.Width; cell++ {
			lvl.Grid[row][cell] = TileWall
		}
	}
	for row := 0; row < lvl.Height; row++ {
		for cell := 0; cell < lvl.Width; cell += lvl.Width - 1 {
			lvl.Grid[row][cell] = TileWall
		}
	}

	// Create small spaces everywhere
	for row := 2; row < lvl.Height; row += 2 {
		for cell := 0; cell < lvl.Width; cell++ {
			lvl.Grid[row][cell] = TileWall
		}
	}
	for row := 0; row < lvl.Height; row++ {
		for cell := 0; cell < lvl.Width; cell += 2 {
			lvl.Grid[row][cell] = TileWall
		}
	}

	// Connect those spaces with each other
	type point struct{ x, y int }
	var unconnected []point
	for row := 1; row < lvl.Height; row += 2 {
		for cell := 1; cell < lvl.Width; cell += 2 {
			unconnected = append(unconnected, point{cell, row})
		}
	}

	// Randomly locate goal node and player
	goalX, goalY := rand.Intn(lvl.Width-2)+1, rand.Intn(lvl.Height-2)+1
	lvl.Grid[goalY][goalX] = TileGoal

	lvl.PlayerX, lvl.PlayerY = lvl.Width-2, lvl.Height-2

	counter := 0
	for len(unconnected) > 0 && counter < 20 {
		counter++
		var removeIndex []int
		// Make sure the player can reach every location in unconnected
		for index, un := range unconnected {
			route := computeRoute(&lvl, lvl.PlayerX, lvl.PlayerY, un.x, un.y)
			if len(route) != 1 || route[0] != ActionStop {
				removeIndex = append(removeIndex, index)
				continue // with other nodes
			}

			// Break one of four walls

			var pos []Action
			if un.x > 1 && lvl.Grid[un.y][un.x-1] == TileWall {
				pos = append(pos, ActionLeft)
			}
			if un.x < lvl.Width-2 && lvl.Grid[un.y][un.x+1] == TileWall {
				pos = append(pos, ActionRight)
			}
			if un.y > 1 && lvl.Grid[un.y-1][un.x] == TileWall {
				pos = append(pos, ActionUp)
			}
			if un.y < lvl.Height-2 && lvl.Grid[un.y+1][un.x] == TileWall {
				pos = append(pos, ActionDown)
			}
			if len(pos) == 0 {
				continue // with other nodes
			}

			randy := rand.Intn(len(pos))
			var newX, newY int
			var newX2, newY2 int
			switch pos[randy] {
			case ActionLeft:
				newY, newX = un.y, un.x-1
				newY2, newX2 = un.y, un.x-2
			case ActionRight:
				newY, newX = un.y, un.x+1
				newY2, newX2 = un.y, un.x+2
			case ActionUp:
				newY, newX = un.y-1, un.x
				newY2, newX2 = un.y-2, un.x
			case ActionDown:
				newY, newX = un.y+1, un.x
				newY2, newX2 = un.y+2, un.x
			}

			// Check to see if we can already reach that point
			route = computeRoute(&lvl, un.x, un.y, newX2, newY2)
			if len(route) != 1 || route[0] != ActionStop {
				removeIndex = append(removeIndex, index)
				continue // with other nodes
			}

			lvl.Grid[newY][newX] = TileBlank
		}

		// Remove all those who are connected
		for removeI, unconnectedIndex := range removeIndex {
			if unconnectedIndex < len(unconnected)-1 {
				unconnected = append(unconnected[0:unconnectedIndex-removeI], unconnected[unconnectedIndex-removeI+1:]...)
			} else {
				unconnected = unconnected[0 : unconnectedIndex-removeI]
			}
		}
	}

	return lvl
}
