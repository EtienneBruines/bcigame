package scenes

import (
	"github.com/EtienneBruines/bcigame/systems"
	"github.com/paked/engi"
	"github.com/paked/engi/ecs"
)

type Menu struct{}

func (*Menu) Hide()        {}
func (*Menu) Show()        {}
func (*Menu) Type() string { return "MenuScene" }

func (m *Menu) Preload() {}
func (m *Menu) Setup(w *ecs.World) {
	w.AddSystem(&engi.AudioSystem{})
	w.AddSystem(&engi.RenderSystem{})
	w.AddSystem(&systems.FPS{})
	w.AddSystem(&systems.Menu{})
}
