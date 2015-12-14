package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"

	"github.com/EtienneBruines/bcigame/scenes"
	"github.com/EtienneBruines/bcigame/systems"
	"github.com/paked/engi"
	"github.com/paked/engi/ecs"
)

const (
	gameTitle  = "BCI Game"
	assetsDir  = "assets"
	levelsDir  = "levels"
	cpuprofile = "cpu.out"
)

type BCIGame struct{}

func (b *BCIGame) Preload() {
	engi.Files.AddFromDir(assetsDir, true)
}

func (b *BCIGame) Setup(w *ecs.World) {
	engi.SetBg(0x444444)

	w.AddSystem(&systems.MenuListener{})
	w.AddSystem(&systems.Maze{LevelDirectory: filepath.Join(assetsDir, levelsDir), Controller: &systems.ErroneousKeyboardController{}})
	w.AddSystem(&systems.FPS{BaseTitle: gameTitle})
	w.AddSystem(&systems.MovementSystem{})
	w.AddSystem(&systems.Calibrate{})
	w.AddSystem(&engi.RenderSystem{})
}

func (*BCIGame) Show()        {}
func (*BCIGame) Hide()        {}
func (*BCIGame) Type() string { return "BCIGame" }

func main() {
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for range c {
				pprof.StopCPUProfile()
				os.Exit(0)
			}
		}()
	}

	engi.RegisterScene(&scenes.Menu{})
	engi.RegisterScene(&scenes.Calibrate{})

	// TODO: don't hardcode this
	engi.Open(gameTitle, 1600, 800, false, &BCIGame{})
}
