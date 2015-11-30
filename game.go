package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"

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

	w.AddSystem(&systems.Menu{})
	w.AddSystem(&systems.Maze{LevelDirectory: filepath.Join(assetsDir, levelsDir), Controller: &systems.AutoPilotController{}})
	w.AddSystem(&systems.FPS{BaseTitle: gameTitle})
	w.AddSystem(&systems.MovementSystem{})
	w.AddSystem(&engi.RenderSystem{})
	w.AddSystem(&engi.AudioSystem{})
	w.AddSystem(&systems.Calibrate{})

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
	engi.Open(gameTitle, 1600, 800, false, &BCIGame{})
}
