package systems

import (
	"github.com/paked/engi"
	"time"
)

type MovementSystem struct {
	*engi.System
}

func (MovementSystem) Type() string { return "MovementSystem" }

func (a *MovementSystem) New() {
	a.System = engi.NewSystem()
}

func (a *MovementSystem) Update(entity *engi.Entity, dt float32) {
	var move *MovementComponent
	if !entity.GetComponent(&move) {
		return
	}

	if !move.started {
		move.started = true
		move.timeLeft = float32(move.In.Seconds())
		move.speedX = (move.From.X - move.To.X) / move.timeLeft
		move.speedY = (move.From.Y - move.To.Y) / move.timeLeft
	}
	move.timeLeft -= dt

	var space *engi.SpaceComponent
	if !entity.GetComponent(&space) {
		return
	}

	space.Position.X -= move.speedX * dt
	space.Position.Y -= move.speedY * dt

	if move.timeLeft < 0 {
		// Because we might move more than needed
		space.Position.X -= move.speedX * move.timeLeft
		space.Position.Y -= move.speedY * move.timeLeft
		move.timeLeft = 0
	}

	if move.timeLeft == 0 {
		entity.RemoveComponent(move)
	}
}

type MovementComponent struct {
	From engi.Point
	To   engi.Point
	In   time.Duration

	started  bool
	timeLeft float32
	speedX   float32
	speedY   float32
}

func (MovementComponent) Type() string { return "MovementComponent" }