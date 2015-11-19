package systems

import (
	"container/heap"
	"github.com/paked/engi"
	"log"
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

type AutoPilotController struct{}

func (ac *AutoPilotController) New() {}

func (ac *AutoPilotController) Action(l Level) Action {
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

		if l.PlayerX == goalX && l.PlayerY == goalY {
			return ActionStop // we already achieved goal
		}

		// Compute the route
		pq := &actionPriorityQueue{}
		heap.Init(pq)
		pq.Push(&priorityQueItem{value: State{nil, l.PlayerX, l.PlayerY}})

		// Keep track of what we have visited
		visited := make([][]bool, l.Height)
		for visIndex := range visited {
			visited[visIndex] = make([]bool, l.Width)
		}

		for i := 0; pq.Len() > 0 && i < maxIterations; i++ {
			pqitem := pq.Pop().(*priorityQueItem)
			state := pqitem.value
			visited[state.Y][state.X] = true

			for _, action := range possibleActions(&l, state.X, state.Y) {
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
					ai.Route = newActionList
					break
				}

				// to add something
				pq.Push(&priorityQueItem{
					value:    State{newActionList, x2, y2},
					priority: -manhattanDistance(x2, y2, goalX, goalY) - len(newActionList),
				})
			}
		}
	}

	if len(ai.Route) == 0 {
		log.Println("No solution found after", maxIterations, "iteractions")
		return ActionStop
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
