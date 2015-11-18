package systems

import (
	"github.com/EtienneBruines/bcigame/helpers"
	"github.com/paked/engi"
	"image/color"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"time"
)

var (
	tileWidth  float32 = 80
	tileHeight float32 = 80

	tilePlayerColor = color.NRGBA{0, 0, 100, 255}
	tileWallColor   = color.NRGBA{0, 100, 0, 255}
	tileBlankColor  = color.NRGBA{180, 180, 180, 255}
	tileGoalColor   = color.NRGBA{0, 255, 255, 255}
	tileRouteColor  = color.NRGBA{255, 0, 0, 255}

	tilePlayer *engi.RenderComponent
	tileWall   *engi.RenderComponent
	tileBlank  *engi.RenderComponent
	tileGoal   *engi.RenderComponent
	tileRoute  *engi.RenderComponent
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
	Width        int
	Height       int
	Grid         [][]Tile
	GridEntities [][]*engi.Entity

	PlayerX, PlayerY int
}

func (l *Level) IsAvailable(x, y int) bool {
	if x < 0 || x >= l.Width || y < 0 || y >= l.Height {
		return false
	}

	return l.Grid[y][x] != TileWall
}

type Controller func(Level) Action

type Action uint8

const (
	ActionUp Action = iota
	ActionRight
	ActionDown
	ActionLeft
	ActionStop
)

type Maze struct {
	*engi.System
	LevelDirectory string
	Controller     Controller

	active bool

	levels []Level

	currentLevel *Level
	playerEntity *engi.Entity
}

func (Maze) Type() string { return "MazeSystem" }

func (m *Maze) New() {
	m.System = engi.NewSystem()

	tilePlayer = helpers.GenerateSquareComonent(tilePlayerColor, tilePlayerColor, tileWidth, tileHeight, engi.MiddleGround)
	tileWall = helpers.GenerateSquareComonent(tileWallColor, tileWallColor, tileWidth, tileHeight, engi.ScenicGround+1)
	tileBlank = helpers.GenerateSquareComonent(tileBlankColor, tileBlankColor, tileWidth, tileHeight, engi.ScenicGround+2)
	tileGoal = helpers.GenerateSquareComonent(tileGoalColor, tileGoalColor, tileWidth, tileHeight, engi.ScenicGround+3)
	tileRoute = helpers.GenerateSquareComonent(tileRouteColor, tileRouteColor, tileWidth, tileHeight, engi.ScenicGround+4)

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
				case '+':
					gameRow[index] = TileRoute
				}
			}
			lvl.Grid = append(lvl.Grid, gameRow)
		}

		m.levels = append(m.levels, lvl)
	}
}

func (m *Maze) cleanup() {
	m.active = false

	m.currentLevel = nil
	m.playerEntity = nil

	for _, entity := range m.Entities() {
		m.World.RemoveEntity(entity)
	}
}

func (m *Maze) initialize() {
	m.active = true

	if len(m.levels) < 4 {
		return
	}

	m.currentLevel = &m.levels[2]

	// Create world
	engi.WorldBounds.Max = engi.Point{float32(m.currentLevel.Width) * tileWidth, float32(m.currentLevel.Height) * tileHeight}

	engi.Mailbox.Dispatch(engi.CameraMessage{engi.XAxis, float32(m.currentLevel.Width) * tileWidth / 2, false})
	engi.Mailbox.Dispatch(engi.CameraMessage{engi.YAxis, float32(m.currentLevel.Height) * tileHeight / 2, false})

	// Initialize the tiles
	m.currentLevel.GridEntities = make([][]*engi.Entity, len(m.currentLevel.Grid))
	for rowNumber, tileRow := range m.currentLevel.Grid {
		m.currentLevel.GridEntities[rowNumber] = make([]*engi.Entity, len(tileRow))
		for columnNumber, tile := range tileRow {
			e := engi.NewEntity([]string{"RenderSystem"})
			e.AddComponent(&engi.SpaceComponent{engi.Point{float32(columnNumber) * tileWidth, float32(rowNumber) * tileHeight}, tileWidth, tileHeight})

			switch tile {
			case TilePlayer:
				// set player location
				m.currentLevel.PlayerX, m.currentLevel.PlayerY = columnNumber, rowNumber
				fallthrough
			case TileBlank:
				e.AddComponent(tileBlank)
			case TileWall:
				e.AddComponent(tileWall)
			case TileGoal:
				e.AddComponent(tileGoal)
			case TileRoute:
				e.AddComponent(tileRoute)
			}

			m.currentLevel.GridEntities[rowNumber][columnNumber] = e
			m.World.AddEntity(e)
		}
	}

	// Draw the player
	m.playerEntity = engi.NewEntity([]string{"RenderSystem", "MovementSystem", m.Type()})
	m.playerEntity.AddComponent(tilePlayer)
	m.playerEntity.AddComponent(&engi.SpaceComponent{engi.Point{float32(m.currentLevel.PlayerX) * tileWidth, float32(m.currentLevel.PlayerY) * tileHeight}, tileWidth, tileHeight})
	m.World.AddEntity(m.playerEntity)
}

func (m *Maze) Update(entity *engi.Entity, dt float32) {
	if entity.ID() != m.playerEntity.ID() {
		return
	}

	var move *MovementComponent
	if entity.GetComponent(&move) {
		return // because we're still moving!
	}

	oldX, oldY := m.currentLevel.PlayerX, m.currentLevel.PlayerY

	switch m.Controller(*m.currentLevel) {
	case ActionUp:
		m.currentLevel.PlayerY--
	case ActionDown:
		m.currentLevel.PlayerY++
	case ActionLeft:
		m.currentLevel.PlayerX--
	case ActionRight:
		m.currentLevel.PlayerX++
	case ActionStop:
		return // so don't animate
	}

	entity.AddComponent(&MovementComponent{
		From: engi.Point{float32(oldX) * tileWidth, float32(oldY) * tileHeight},
		To:   engi.Point{float32(m.currentLevel.PlayerX) * tileWidth, float32(m.currentLevel.PlayerY) * tileHeight},
		In:   time.Millisecond * 150,
		Callback: func() {
			if m.currentLevel.Grid[m.currentLevel.PlayerY][m.currentLevel.PlayerX] == TileRoute {
				m.currentLevel.Grid[m.currentLevel.PlayerY][m.currentLevel.PlayerX] = TileBlank
				m.currentLevel.GridEntities[m.currentLevel.PlayerY][m.currentLevel.PlayerX].AddComponent(tileBlank)
			}
		},
	})
}

func ControllerKeyboard(l Level) Action {
	if engi.Keys.Get(engi.D).Down() && l.IsAvailable(l.PlayerX+1, l.PlayerY) {
		return ActionRight
	} else if engi.Keys.Get(engi.A).Down() && l.IsAvailable(l.PlayerX-1, l.PlayerY) {
		return ActionLeft
	} else if engi.Keys.Get(engi.S).Down() && l.IsAvailable(l.PlayerX, l.PlayerY+1) {
		return ActionDown
	} else if engi.Keys.Get(engi.W).Down() && l.IsAvailable(l.PlayerX, l.PlayerY-1) {
		return ActionUp
	}

	return ActionStop
}

func ControllerAutoPilot(l Level) Action {
	priority := []Tile{TileGoal, TileRoute}

	for _, p := range priority {
		if l.Grid[l.PlayerY][l.PlayerX-1] == p {
			return ActionLeft
		} else if l.Grid[l.PlayerY][l.PlayerX+1] == p {
			return ActionRight
		} else if l.Grid[l.PlayerY-1][l.PlayerX] == p {
			return ActionUp
		} else if l.Grid[l.PlayerY+1][l.PlayerX] == p {
			return ActionDown
		}
	}

	return ActionStop
}

type MazeMessage struct{}

func (MazeMessage) Type() string { return "MazeMessage" }
