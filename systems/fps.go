package systems

import (
	"fmt"
	"github.com/paked/engi"
	"github.com/paked/engi/ecs"
)

// FPS shows the current frames per second in the window title
type FPS struct {
	*ecs.System
	World *ecs.World

	BaseTitle string
	lastFPS   int
}

func (*FPS) Type() string {
	return "FPSSystem"
}

func (f *FPS) New(w *ecs.World) {
	f.System = ecs.NewSystem()
	f.World = w

	f.AddEntity(ecs.NewEntity([]string{f.Type()}))
}

func (f *FPS) Update(entity *ecs.Entity, dt float32) {
	if fps := engi.Time.Fps(); f.lastFPS != int(fps) {
		f.lastFPS = int(fps)
		engi.SetTitle(fmt.Sprintf("%s - [FPS: %.0f] - %d", f.BaseTitle, fps, len(f.World.Entities())))
	}
}
