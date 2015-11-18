package main

import (
	"github.com/EtienneBruines/bcigame/systems"
	"github.com/paked/engi"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
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

func (b *BCIGame) Setup(w *engi.World) {
	engi.SetBg(0x444444)

	w.AddSystem(&systems.Menu{})
	w.AddSystem(&systems.Maze{LevelDirectory: filepath.Join(assetsDir, levelsDir), Controller: systems.ControllerAutoPilot})
	w.AddSystem(&systems.FPS{BaseTitle: gameTitle})
	w.AddSystem(&systems.MovementSystem{})
	w.AddSystem(&engi.PauseSystem{})
	w.AddSystem(&engi.RenderSystem{})
	//w.AddSystem(&systems.Calibrate{})

	engi.Mailbox.Dispatch(systems.MazeMessage{})
}

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

	// TODO: don't hardcode this
	engi.Open(gameTitle, 400, 800, false, &BCIGame{})
}
