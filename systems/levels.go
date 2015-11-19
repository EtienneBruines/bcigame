package systems

import (
	"github.com/paked/engi"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

type Tile uint8

const (
	TilePlayer Tile = iota
	TileWall
	TileBlank
	TileGoal
	TileRoute
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
	GridEntities [][]*engi.Entity

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
				}
			}
			lvl.Grid = append(lvl.Grid, gameRow)
		}

		levels = append(levels, lvl)
	}
	return
}
