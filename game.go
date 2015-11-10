package main

import (
	"github.com/EtienneBruines/bcigame/systems"
	"github.com/paked/engi"
	"path/filepath"
)

const (
	gameTitle = "BCI Game"
	assetsDir = "assets"
	levelsDir = "levels"
)

type BCIGame struct{}

func (b *BCIGame) Preload() {
	engi.Files.AddFromDir(assetsDir, true)
}

func (b *BCIGame) Setup(w *engi.World) {
	engi.SetBg(0x444444)

	w.AddSystem(&systems.Menu{})
	w.AddSystem(&systems.Maze{LevelDirectory: filepath.Join(assetsDir, levelsDir)})
	w.AddSystem(&engi.PauseSystem{})
	w.AddSystem(&systems.FPS{BaseTitle: gameTitle})
	w.AddSystem(&engi.RenderSystem{})
	//w.AddSystem(&systems.Calibrate{})
}

func main() {
	// TODO: don't hardcode this
	engi.Open(gameTitle, 800, 870, false, &BCIGame{})
}
