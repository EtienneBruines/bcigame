package systems

import (
	"github.com/EtienneBruines/bcigame/helpers"
	"github.com/paked/engi"
	"image/color"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

var (
	mazeBackground         = color.NRGBA{255, 255, 255, 255}
	tileWidth      float32 = 80
	tileHeight     float32 = 80
)

type Tile uint8

const (
	TilePlayer Tile = iota
	TileWall
	TileBlank
	TileGoal
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
	default:
		return ""
	}
}

type Level struct {
	Width  int
	Height int
	Grid   [][]Tile

	PlayerX, PlayerY int
}

func (l *Level) IsAvailable(x, y int) bool {
	if x < 0 || x >= l.Width || y < 0 || y >= l.Height {
		return false
	}

	return l.Grid[y][x] != TileWall
}

type Maze struct {
	*engi.System
	LevelDirectory string

	active bool

	levels []Level

	currentLevel *Level
	playerEntity *engi.Entity
}

func (Maze) Type() string { return "MazeSystem" }

func (m *Maze) New() {
	m.System = engi.NewSystem()

	m.loadLevels()

	engi.Mailbox.Listen("MazeMessage", func(msg engi.Message) {
		_, ok := msg.(MazeMessage)
		if !ok {
			return
		}
		if m.active {
			m.cleanup()
		}
		m.initialize()
	})
}

func (m *Maze) loadLevels() {
	infos, err := ioutil.ReadDir(m.LevelDirectory)
	if err != nil {
		log.Fatal(err)
	}

	var files []string

	for _, info := range infos {
		if !info.IsDir() {
			ext := filepath.Ext(info.Name())
			if ext[1:] == "maze" {
				files = append(files, filepath.Join(m.LevelDirectory, info.Name()))
			}
		}
	}

	for _, file := range files {
		lvl := Level{}

		b, err := ioutil.ReadFile(file)
		if err != nil {
			continue // with other files
		}

		content := string(b)

		lines := strings.Split(content, "\n")
		lvl.Height = len(lines)

		for _, line := range lines {
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
				}
			}
			lvl.Grid = append(lvl.Grid, gameRow)
		}

		m.levels = append(m.levels, lvl)
	}
}

func (m *Maze) cleanup() {
	m.active = false
}

func (m *Maze) initialize() {
	log.Println("Initialized", m.levels)
	m.active = true

	if len(m.levels) == 0 {
		return
	}

	m.currentLevel = &m.levels[0]

	// Create world
	engi.WorldBounds.Max = engi.Point{float32(m.currentLevel.Width) * tileWidth, float32(m.currentLevel.Height) * tileHeight}

	log.Println(float32(m.currentLevel.Width)*tileWidth, float32(m.currentLevel.Height)*tileHeight)

	engi.Mailbox.Dispatch(engi.CameraMessage{engi.XAxis, float32(m.currentLevel.Width) * tileWidth / 2, false})
	engi.Mailbox.Dispatch(engi.CameraMessage{engi.YAxis, float32(m.currentLevel.Height) * tileHeight / 2, false})

	// Initialize the tiles
	for rowNumber, tileRow := range m.currentLevel.Grid {
		for columnNumber, tile := range tileRow {
			var c color.NRGBA

			switch tile {
			case TilePlayer:
				// set player location
				m.currentLevel.PlayerX, m.currentLevel.PlayerY = columnNumber, rowNumber
				fallthrough
			case TileBlank:
				c = color.NRGBA{180, 180, 180, 255}
			case TileWall:
				c = color.NRGBA{0, 100, 0, 255}
			case TileGoal:
				c = color.NRGBA{0, 255, 255, 255}
			}

			m.World.AddEntity(helpers.GenerateSquare(c, c,
				tileWidth, tileHeight,
				float32(columnNumber)*tileWidth, float32(rowNumber)*tileHeight, engi.ScenicGround))
		}
	}

	c := color.NRGBA{0, 0, 100, 255}
	// Draw the player
	m.playerEntity = helpers.GenerateSquare(c, c,
		tileWidth, tileHeight,
		float32(m.currentLevel.PlayerX)*tileWidth, float32(m.currentLevel.PlayerY)*tileHeight, engi.MiddleGround, m.Type())
	m.World.AddEntity(m.playerEntity)
}

func (m *Maze) Update(entity *engi.Entity, dt float32) {
	if entity.ID() != m.playerEntity.ID() {
		return
	}

	var changed bool

	if engi.Keys.Get(engi.D).Down() {
		if m.currentLevel.IsAvailable(m.currentLevel.PlayerX+1, m.currentLevel.PlayerY) {
			m.currentLevel.PlayerX += 1
			changed = true
		}
	} else if engi.Keys.Get(engi.A).Down() {
		if m.currentLevel.IsAvailable(m.currentLevel.PlayerX-1, m.currentLevel.PlayerY) {
			m.currentLevel.PlayerX -= 1
			changed = true
		}
	}

	if engi.Keys.Get(engi.S).Down() {
		if m.currentLevel.IsAvailable(m.currentLevel.PlayerX, m.currentLevel.PlayerY+1) {
			m.currentLevel.PlayerY += 1
			changed = true
		}
	} else if engi.Keys.Get(engi.W).Down() {
		if m.currentLevel.IsAvailable(m.currentLevel.PlayerX, m.currentLevel.PlayerY-1) {
			m.currentLevel.PlayerY -= 1
			changed = true
		}
	}

	if !changed {
		return
	}

	var space *engi.SpaceComponent
	if !m.playerEntity.GetComponent(&space) {
		return
	}

	space.Position.X = float32(m.currentLevel.PlayerX) * tileWidth
	space.Position.Y = float32(m.currentLevel.PlayerY) * tileHeight
}

type MazeMessage struct{}

func (MazeMessage) Type() string { return "MazeMessage" }
