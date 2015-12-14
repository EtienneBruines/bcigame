package scenes

import (
	"github.com/EtienneBruines/bcigame/systems"
	"github.com/paked/engi"
	"github.com/paked/engi/ecs"
)

type Calibrate struct{}

func (*Calibrate) Preload() {}
func (*Calibrate) Setup(w *ecs.World) {
	w.AddSystem(&engi.RenderSystem{})
	w.AddSystem(&systems.FPS{})
	w.AddSystem(&systems.MenuListener{})
	w.AddSystem(&systems.Calibrate{Visualize: true})
}

func (*Calibrate) Show()        {}
func (*Calibrate) Hide()        {}
func (*Calibrate) Type() string { return "CalibrateScene" }
