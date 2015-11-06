package systems

import (
	"image/color"

	"bitbucket.org/etiennebruines/bcigame/helpers"
	"github.com/paked/engi"
)

var (
	MenuColorBackground = color.RGBA{255, 255, 255, 125}
	MenuColorBox        = color.RGBA{180, 180, 180, 255}
	menuColor1          = color.RGBA{102, 153, 0, 255}
	menuColor2          = color.RGBA{153, 102, 0, 255}
	menuPadding         = float32(50)
)

type Menu struct {
	*engi.System
	coreID string

	menuActive bool
}

func (*Menu) Type() string {
	return "MenuSystem"
}

func (m *Menu) New() {
	m.System = engi.NewSystem()

	e := engi.NewEntity([]string{m.Type()})
	e.AddComponent(&engi.UnpauseComponent{})
	m.AddEntity(e)
	m.coreID = e.ID()
}

func (m *Menu) Update(entity *engi.Entity, dt float32) {
	if engi.Keys.KEY_ESCAPE.JustPressed() {
		if m.menuActive {
			m.closeMenu()
		} else {
			m.openMenu()
		}
		m.menuActive = !m.menuActive
	}
}

func (m *Menu) closeMenu() {
	// Unpause everything
	engi.Mailbox.Dispatch(engi.PauseMessage{false})

	// Remove all entities
	for _, e := range m.Entities() {
		if e.ID() != m.coreID {
			m.World.RemoveEntity(e)
		}
	}
}

func (m *Menu) openMenu() {
	// Pause everything
	engi.Mailbox.Dispatch(engi.PauseMessage{true})

	// Create the visual menu
	// - background
	backgroundWidth := engi.Width() / 2
	backgroundHeight := engi.Height() / 2

	menuBackground := helpers.GenerateSquare(
		MenuColorBackground, MenuColorBackground,
		backgroundWidth, backgroundHeight,
		0, 0,
		engi.HUDGround, m.Type(),
	)
	menuBackground.AddComponent(&engi.UnpauseComponent{})
	m.World.AddEntity(menuBackground)

	// - box
	menuWidth := (engi.Width() - 2*2*menuPadding) / 2
	menuHeight := (engi.Height() - 2*2*menuPadding) / 2

	menuEntity := helpers.GenerateSquare(
		MenuColorBox, MenuColorBox,
		menuWidth, menuHeight,
		menuPadding, menuPadding,
		engi.HUDGround+1, m.Type(),
	)
	menuEntity.AddComponent(&engi.UnpauseComponent{})
	m.World.AddEntity(menuEntity)

}
