package systems

import (
	"image/color"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/EtienneBruines/bcigame/helpers"
	"github.com/paked/engi"
	"github.com/paked/engi/ecs"
)

const (
	tileWidth  float32 = 80
	tileHeight float32 = 80

	moveSpeed = 3.0

	randomMinWidth  = 15
	randomMaxWidth  = 35
	randomMinHeight = 5
	randomMaxHeight = 25
)

var (
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

var ActiveMazeSystem *Maze

type Maze struct {
	*ecs.System
	World *ecs.World

	LevelDirectory string
	Controller     Controller

	active bool

	levels []Level

	currentLevel Level
	playerEntity *ecs.Entity
}

func (Maze) Type() string { return "MazeSystem" }

func (m *Maze) New(w *ecs.World) {
	ActiveMazeSystem = m
	m.System = ecs.NewSystem()
	m.World = w

	tilePlayer = helpers.GenerateSquareComonent(tilePlayerColor, tilePlayerColor, tileWidth, tileHeight, engi.MiddleGround)
	tileWall = helpers.GenerateSquareComonent(tileWallColor, tileWallColor, tileWidth, tileHeight, engi.ScenicGround+1)
	tileBlank = helpers.GenerateSquareComonent(tileBlankColor, tileBlankColor, tileWidth, tileHeight, engi.ScenicGround+2)
	tileGoal = helpers.GenerateSquareComonent(tileGoalColor, tileGoalColor, tileWidth, tileHeight, engi.ScenicGround+3)
	tileRoute = helpers.GenerateSquareComonent(tileRouteColor, tileRouteColor, tileWidth, tileHeight, engi.ScenicGround+4)

	m.levels = LoadLevels(m.LevelDirectory)

	engi.Mailbox.Listen("MazeMessage", func(msg engi.Message) {
		mazeMsg, ok := msg.(MazeMessage)
		if !ok {
			return
		}
		if _, ok = m.Controller.(*AIController); ok && m.active {
			m.currentLevel.Save("assets/levels/random-" + strconv.Itoa(int(time.Now().Unix())) + ".maze")
		}
		m.cleanup()
		m.initialize(mazeMsg.LevelName)
	})
}

func (m *Maze) cleanup() {
	m.active = false

	for _, row := range m.currentLevel.GridEntities {
		for _, cell := range row {
			m.World.RemoveEntity(cell)
		}
	}

	if m.playerEntity != nil {
		m.World.RemoveEntity(m.playerEntity)
		m.playerEntity = nil
	}

	m.currentLevel = emptyLevel

	for _, entity := range m.Entities() {
		m.World.RemoveEntity(entity)
	}
}

func (m *Maze) initialize(level string) {
	m.active = true

	if len(level) == 0 {
		m.currentLevel = NewRandomLevel(randomMinWidth, randomMaxWidth, randomMinHeight, randomMaxHeight)
	} else {
		for lvlId := range m.levels {
			if m.levels[lvlId].Name == level {
				m.currentLevel = m.levels[lvlId].Copy()
				break
			}
		}
	}

	if m.currentLevel.ID == emptyLevel.ID {
		if len(m.levels) > 0 {
			m.currentLevel = m.levels[rand.Intn(len(m.levels))].Copy()
		} else {
			return
		}
	}

	ActiveCalibrateSystem.Connection.PutEvent("Started Level", m.currentLevel.Name)

	// Create world
	engi.WorldBounds.Max = engi.Point{float32(m.currentLevel.Width) * tileWidth, float32(m.currentLevel.Height) * tileHeight}

	engi.Mailbox.Dispatch(engi.CameraMessage{engi.XAxis, float32(m.currentLevel.Width) * tileWidth / 2, false})
	engi.Mailbox.Dispatch(engi.CameraMessage{engi.YAxis, float32(m.currentLevel.Height) * tileHeight / 2, false})

	// Initialize the tiles
	m.currentLevel.GridEntities = make([][]*ecs.Entity, len(m.currentLevel.Grid))
	for rowNumber, tileRow := range m.currentLevel.Grid {
		m.currentLevel.GridEntities[rowNumber] = make([]*ecs.Entity, len(tileRow))
		for columnNumber, tile := range tileRow {
			e := ecs.NewEntity([]string{"RenderSystem"})
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
			case TileError:
				e.AddComponent(tileRoute)
			case TileHiddenError:
				e.AddComponent(tileBlank)
			}

			m.currentLevel.GridEntities[rowNumber][columnNumber] = e
			m.World.AddEntity(e)
		}
	}

	// Draw the player
	m.playerEntity = ecs.NewEntity([]string{"RenderSystem", "MovementSystem", m.Type()})
	m.playerEntity.AddComponent(tilePlayer)
	m.playerEntity.AddComponent(&engi.SpaceComponent{engi.Point{float32(m.currentLevel.PlayerX) * tileWidth, float32(m.currentLevel.PlayerY) * tileHeight}, tileWidth, tileHeight})
	m.World.AddEntity(m.playerEntity)

	// Initialize the controller
	m.Controller.New()
}

func (m *Maze) Update(entity *ecs.Entity, dt float32) {
	if entity.ID() != m.playerEntity.ID() {
		return
	}

	var (
		move *MovementComponent
		ok   bool
	)

	if move, ok = entity.ComponentFast(move).(*MovementComponent); ok {
		return // because we're still moving!
	}

	if m.currentLevel.Width == 0 || m.currentLevel.Height == 0 {
		return // because there's no maze
	}

	oldX, oldY := m.currentLevel.PlayerX, m.currentLevel.PlayerY

	if m.currentLevel.Grid[oldY][oldX] == TileGoal {
		// Goal achieved!
		if strings.HasPrefix(m.currentLevel.Name, "Random ") {
			engi.Mailbox.Dispatch(MazeMessage{})
			return
		}
	}

	switch m.Controller.Action(m.currentLevel) {
	case ActionUp:
		m.currentLevel.PlayerY--
	case ActionDown:
		m.currentLevel.PlayerY++
	case ActionLeft:
		m.currentLevel.PlayerX--
	case ActionRight:
		m.currentLevel.PlayerX++
	case ActionStop:
		return // so don't move
	}

	if !m.currentLevel.IsAvailable(m.currentLevel.PlayerX, m.currentLevel.PlayerY) {
		m.currentLevel.PlayerX, m.currentLevel.PlayerY = oldX, oldY
		return // because it's an invalid move
	}

	entity.AddComponent(&MovementComponent{
		From: engi.Point{float32(oldX) * tileWidth, float32(oldY) * tileHeight},
		To:   engi.Point{float32(m.currentLevel.PlayerX) * tileWidth, float32(m.currentLevel.PlayerY) * tileHeight},
		In:   time.Second / moveSpeed,
		Callback: func() {
			if m.currentLevel.Grid[m.currentLevel.PlayerY][m.currentLevel.PlayerX] == TileRoute {
				m.currentLevel.Grid[m.currentLevel.PlayerY][m.currentLevel.PlayerX] = TileBlank
				m.currentLevel.GridEntities[m.currentLevel.PlayerY][m.currentLevel.PlayerX].AddComponent(tileBlank)
			} else if m.currentLevel.Grid[oldY][oldX] == TileError {
				m.currentLevel.Grid[oldY][oldX] = TileRoute

				if m.currentLevel.Grid[m.currentLevel.PlayerY][m.currentLevel.PlayerX] == TileHiddenError {
					m.currentLevel.Grid[m.currentLevel.PlayerY][m.currentLevel.PlayerX] = TileError
				}
			}
		},
	})
}

type MazeMessage struct {
	LevelName string
}

func (MazeMessage) Type() string { return "MazeMessage" }
