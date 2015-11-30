package systems

import (
	"github.com/paked/engi/ecs"
)

type Hud struct {
	*ecs.System
}

func (h *Hud) New(*ecs.World) {
	h.System = ecs.NewSystem()
}

func (h *Hud) Update(entity *ecs.Entity, dt float32) {

}
