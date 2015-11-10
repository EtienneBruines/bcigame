package main

import (
	"github.com/EtienneBruines/bcigame/helpers"
	"github.com/EtienneBruines/bcigame/systems"
	"github.com/paked/engi"
	"image/color"
)

const (
	gameTitle = "BCI Game"
)

var (
	worldColor1 = color.NRGBA{255, 0, 0, 255}
	worldColor2 = color.NRGBA{255, 255, 0, 255}
)

type BCIGame struct{}

func (b *BCIGame) Preload() {
	engi.Files.AddFromDir("assets", true)
}

func (b *BCIGame) Setup(w *engi.World) {
	engi.SetBg(0x444444)

	w.AddSystem(&systems.Menu{})
	//	w.AddSystem(&engi.PauseSystem{})
	//	w.AddSystem(&systems.FPS{BaseTitle: gameTitle})
	w.AddSystem(&engi.RenderSystem{})
	//w.AddSystem(&systems.Calibrate{})

	w.AddSystem(engi.NewKeyboardScroller(800, engi.W, engi.D, engi.S, engi.A))

	w.AddEntity(helpers.GenerateSquare(worldColor1, worldColor2, 300, 300, 0, 0, engi.Background))
}

func main() {
	// TODO: don't hardcode this
	engi.Open(gameTitle, 800, 870, false, &BCIGame{})
}
