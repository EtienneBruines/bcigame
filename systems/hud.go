package systems

import "github.com/paked/engi"

type Hud struct {
	*engi.System
}

func (h *Hud) New() {
	h.System = engi.NewSystem()
}

func (h *Hud) Update(entity *engi.Entity, dt float32) {

}
