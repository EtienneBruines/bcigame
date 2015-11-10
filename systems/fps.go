package systems

import (
	"fmt"
	"github.com/paked/engi"
)

// FPS shows the current frames per second in the window title
type FPS struct {
	*engi.System

	BaseTitle string
	lastFPS   int
}

func (*FPS) Type() string {
	return "FPSSystem"
}

func (f *FPS) New() {
	f.System = engi.NewSystem()

	f.AddEntity(engi.NewEntity([]string{f.Type()}))
}

func (f *FPS) Update(entity *engi.Entity, dt float32) {
	if fps := engi.Time.Fps(); f.lastFPS != int(fps) {
		f.lastFPS = int(fps)
		engi.SetTitle(fmt.Sprintf("%s - [FPS: %.0f] - %d", f.BaseTitle, fps, len(f.World.Entities())))
	}
}
