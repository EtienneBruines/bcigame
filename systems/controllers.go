package systems

import (
	"container/heap"
	"fmt"

	"github.com/paked/engi"
)

type Controller interface {
	New()
	Action(Level) Action
}

type Action uint8

const (
	ActionUp Action = iota
	ActionRight
	ActionDown
	ActionLeft
	ActionStop
)

type KeyboardController struct{}

func (kb *KeyboardController) New() {}

func (kb *KeyboardController) Action(l Level) Action {
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

type AutoPilotController struct {
	streak int
}

func (ac *AutoPilotController) New() {}

func (ac *AutoPilotController) Action(l Level) Action {
	priority := []Tile{TileGoal, TileHiddenError, TileRoute, TileError}

	action := ActionStop

	for _, p := range priority {
		if l.Grid[l.PlayerY][l.PlayerX-1] == p {
			if p == TileHiddenError {
				ac.streak++
			} else {
				ac.streak = 0
			}
			action = ActionLeft
			break
		} else if l.Grid[l.PlayerY][l.PlayerX+1] == p {
			if p == TileHiddenError {
				ac.streak++
			} else {
				ac.streak = 0
			}
			action = ActionRight
			break
		} else if l.Grid[l.PlayerY-1][l.PlayerX] == p {
			if p == TileHiddenError {
				ac.streak++
			} else {
				ac.streak = 0
			}
			action = ActionUp
			break
		} else if l.Grid[l.PlayerY+1][l.PlayerX] == p {
			if p == TileHiddenError {
				ac.streak++
			} else {
				ac.streak = 0
			}
			action = ActionDown
			break
		}
	}

	if action != ActionStop && ActiveCalibrateSystem != nil {
		ActiveCalibrateSystem.Connection.PutEvent("Tile", fmt.Sprintf("ErrorStreak: %d", ac.streak))
	}

	return action
}

type ErroneousKeyboardController struct {
	streak int
}

func (kb *ErroneousKeyboardController) New() {}

func (kb *ErroneousKeyboardController) Action(l Level) Action {
	action := ActionStop
	priority := []Tile{TileHiddenError, TileError}

	var hiddenPointOfError bool
	var userError bool

	if engi.Keys.Get(engi.D).Down() && l.IsAvailable(l.PlayerX+1, l.PlayerY) {
		action = ActionRight
	} else if engi.Keys.Get(engi.A).Down() && l.IsAvailable(l.PlayerX-1, l.PlayerY) {
		action = ActionLeft
	} else if engi.Keys.Get(engi.S).Down() && l.IsAvailable(l.PlayerX, l.PlayerY+1) {
		action = ActionDown
	} else if engi.Keys.Get(engi.W).Down() && l.IsAvailable(l.PlayerX, l.PlayerY-1) {
		action = ActionUp
	}

	// Check if the user made a mistake
	if action != ActionStop {
		switch action {
		case ActionRight:
			if l.Grid[l.PlayerY][l.PlayerX+1] != TileRoute &&
				l.Grid[l.PlayerY][l.PlayerX+1] != TileGoal &&
				l.Grid[l.PlayerY][l.PlayerX+1] != TileError {
				userError = true
			}
		case ActionLeft:
			if l.Grid[l.PlayerY][l.PlayerX-1] != TileRoute &&
				l.Grid[l.PlayerY][l.PlayerX-1] != TileGoal &&
				l.Grid[l.PlayerY][l.PlayerX-1] != TileError {
				userError = true
			}
		case ActionDown:
			if l.Grid[l.PlayerY+1][l.PlayerX] != TileRoute &&
				l.Grid[l.PlayerY+1][l.PlayerX] != TileGoal &&
				l.Grid[l.PlayerY+1][l.PlayerX] != TileError {
				userError = true
			}
		case ActionUp:
			if l.Grid[l.PlayerY-1][l.PlayerX] != TileRoute &&
				l.Grid[l.PlayerY-1][l.PlayerX] != TileGoal &&
				l.Grid[l.PlayerY-1][l.PlayerX] != TileError {
				userError = true
			}
		}
	}

	// If current tile is ErrorTile, then make a mistake
	if action != ActionStop &&
		l.Grid[l.PlayerY][l.PlayerX] == TileError || l.Grid[l.PlayerY][l.PlayerX] == TileHiddenError {
		for _, p := range priority {
			if l.Grid[l.PlayerY][l.PlayerX-1] == p {
				if action != ActionLeft {
					hiddenPointOfError = true
				}
				action = ActionLeft
				break
			} else if l.Grid[l.PlayerY][l.PlayerX+1] == p {
				if action != ActionRight {
					hiddenPointOfError = true
				}
				action = ActionRight
				break
			} else if l.Grid[l.PlayerY-1][l.PlayerX] == p {
				if action != ActionUp {
					hiddenPointOfError = true
				}
				action = ActionUp
				break
			} else if l.Grid[l.PlayerY+1][l.PlayerX] == p {
				if action != ActionDown {
					hiddenPointOfError = true
				}
				action = ActionDown
				break
			}
		}
	}

	if action != ActionStop && ActiveCalibrateSystem != nil {
		var distance int

		switch action {
		case ActionRight:
			distance = kb.distanceToRoute(&l, l.PlayerX+1, l.PlayerY)
		case ActionLeft:
			distance = kb.distanceToRoute(&l, l.PlayerX-1, l.PlayerY)
		case ActionDown:
			distance = kb.distanceToRoute(&l, l.PlayerX, l.PlayerY+1)
		case ActionUp:
			distance = kb.distanceToRoute(&l, l.PlayerX, l.PlayerY-1)
		}

		fmt.Sprintf("%d", distance)

		if hiddenPointOfError {
			kb.streak++
			ActiveCalibrateSystem.Connection.PutEvent("Tile", fmt.Sprintf("HiddenPointOfError: %d", kb.streak))
		} else if userError {
			kb.streak++
			ActiveCalibrateSystem.Connection.PutEvent("Tile", fmt.Sprintf("UserError: %d", kb.streak))
		} else {
			kb.streak = 0
			ActiveCalibrateSystem.Connection.PutEvent("Tile", fmt.Sprintf("NoError"))
		}
	}

	return action
}

func (kb *ErroneousKeyboardController) distanceToRoute(l *Level, x, y int) int {
	pq := &actionPriorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &priorityQueItem{value: State{nil, x, y}})

	for i := 0; pq.Len() > 0; i++ {
		pqitem := heap.Pop(pq).(*priorityQueItem)
		state := pqitem.value

		if l.Grid[state.Y][state.X] == TileGoal ||
			l.Grid[state.Y][state.X] == TileRoute ||
			l.Grid[state.Y][state.X] == TileError {
			return -pqitem.priority
		}

		var x2, y2 int
		for _, a := range possibleActions(l, state.X, state.Y) {
			switch a {
			case ActionUp:
				x2, y2 = state.X, state.Y-1
			case ActionDown:
				x2, y2 = state.X, state.Y+1
			case ActionLeft:
				x2, y2 = state.X-1, state.Y
			case ActionRight:
				x2, y2 = state.X+1, state.Y
			default:
				x2, y2 = state.X, state.Y
			}

			// to add something
			heap.Push(pq, &priorityQueItem{
				value:    State{nil, x2, y2},
				priority: pqitem.priority - 1,
			})
		}
	}

	return 10000000
}

type State struct {
	Route []Action
	X, Y  int
}

type priorityQueItem struct {
	value    State
	priority int // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

type actionPriorityQueue []*priorityQueItem

func (pq actionPriorityQueue) Len() int { return len(pq) }

func (pq actionPriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].priority > pq[j].priority
}

func (pq actionPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *actionPriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*priorityQueItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *actionPriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

type AIController struct {
	Route []Action
}

func (ai *AIController) New() {
	ai.Route = nil
}

func (ai *AIController) Action(l Level) Action {
	const maxIterations = 1000

	if len(ai.Route) == 0 {

		// Find goal state
		var goalX, goalY int
		for rowIndex, row := range l.Grid {
			for cellIndex, cell := range row {
				if cell == TileGoal {
					goalX, goalY = cellIndex, rowIndex
				}
			}
		}

		ai.Route = computeRoute(&l, l.PlayerX, l.PlayerY, goalX, goalY)
	}

	nextAction := ai.Route[0]
	ai.Route = ai.Route[1:]
	return nextAction
}

func possibleActions(l *Level, x, y int) []Action {
	var actions []Action
	if x > 0 {
		if l.IsAvailable(x-1, y) {
			actions = append(actions, ActionLeft)
		}
	}
	if y > 0 {
		if l.IsAvailable(x, y-1) {
			actions = append(actions, ActionUp)
		}
	}
	if x < l.Width-1 {
		if l.IsAvailable(x+1, y) {
			actions = append(actions, ActionRight)
		}
	}
	if y < l.Height-1 {
		if l.IsAvailable(x, y+1) {
			actions = append(actions, ActionDown)
		}
	}

	return actions
}

func manhattanDistance(x1, y1, x2, y2 int) int {
	diffX := x1 - x2
	diffY := y1 - y2
	if diffX < 0 {
		diffX *= -1
	}
	if diffY < 0 {
		diffY *= -1
	}
	return diffX + diffY
}

func computeRoute(l *Level, startX, startY, goalX, goalY int) []Action {
	const maxIterations = 1000

	if startX == goalX && startY == goalY {
		return []Action{ActionStop} // we already achieved goal
	}

	// Compute the route
	pq := &actionPriorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &priorityQueItem{value: State{nil, startX, startY}})

	// Keep track of what we have visited
	visited := make([][]bool, l.Height)
	for visIndex := range visited {
		visited[visIndex] = make([]bool, l.Width)
	}

	for i := 0; pq.Len() > 0 && i < maxIterations; i++ {
		pqitem := heap.Pop(pq).(*priorityQueItem)
		state := pqitem.value

		for _, action := range possibleActions(l, state.X, state.Y) {
			var x2, y2 int
			switch action {
			case ActionUp:
				x2, y2 = state.X, state.Y-1
			case ActionDown:
				x2, y2 = state.X, state.Y+1
			case ActionLeft:
				x2, y2 = state.X-1, state.Y
			case ActionRight:
				x2, y2 = state.X+1, state.Y
			default:
				x2, y2 = state.X, state.Y
			}

			if visited[y2][x2] {
				continue // with other actions
			}

			// New route list
			newActionList := make([]Action, len(state.Route)+1)
			for index, routeItem := range state.Route {
				newActionList[index] = routeItem
			}
			newActionList[len(state.Route)] = action

			if x2 == goalX && y2 == goalY {
				return newActionList
			}

			// to add something
			heap.Push(pq, &priorityQueItem{
				value:    State{newActionList, x2, y2},
				priority: -manhattanDistance(x2, y2, goalX, goalY),
			})

			visited[y2][x2] = true // because it's queued
		}
	}

	return []Action{ActionStop}
}
