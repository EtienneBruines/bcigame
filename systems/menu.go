package systems

import (
	"github.com/EtienneBruines/bcigame/helpers"
	"github.com/paked/engi"
	"image/color"
	"log"
	"os"
)

var (
	MenuColorBackground          = color.RGBA{255, 255, 255, 125}
	MenuColorBox                 = color.RGBA{180, 180, 180, 255}
	MenuColorItemBackground      = color.RGBA{0, 0, 0, 125}
	MenuColorItemBackgroundFocus = color.RGBA{64, 96, 0, 200}
	MenuColorItemForeground      = engi.Color{255, 255, 255, 255}
	MenuColorItemBox             = color.RGBA{230, 230, 230, 255}

	menuItemHeight      = float32(25)
	menuItemOffsetX     = float32(menuPadding + menuItemPadding)
	menuItemFontPadding = float32(1)
	menuItemPadding     = float32(2.5)
	menuPadding         = float32(50)
)

type MenuItem struct {
	Text     string
	Callback func()
}

type Menu struct {
	*engi.System

	menuActive   bool
	menuEntities []*engi.Entity
	menuFocus    int
	items        []MenuItem
}

func (*Menu) Type() string {
	return "MenuSystem"
}

func (m *Menu) New() {
	m.System = engi.NewSystem()

	e := engi.NewEntity([]string{m.Type()})
	e.AddComponent(&engi.UnpauseComponent{})
	m.AddEntity(e)
	m.items = []MenuItem{
		{"New Game", func() {
			log.Println("New game")
			m.closeMenu()
		}},
		{"Calibrate", func() {
			log.Println("Calibrate")
			m.closeMenu()
		}},
		{"Exit", func() {
			os.Exit(0)
		}},
	}
}

func (m *Menu) Update(entity *engi.Entity, dt float32) {
	// Check for ESCAPE
	if engi.Keys.KEY_ESCAPE.JustReleased() {
		if m.menuActive {
			m.closeMenu()
		} else {
			m.openMenu()
		}
		m.menuActive = !m.menuActive
	}

	if m.menuActive {

	}

	// Check if any button/item is being hovered

	// Check if any button/item is clicked

}

func (m *Menu) closeMenu() {
	// Unpause everything
	engi.Mailbox.Dispatch(engi.PauseMessage{false})

	// Remove all entities
	for _, e := range m.menuEntities {
		m.World.RemoveEntity(e)
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
		engi.HUDGround,
	)
	menuBackground.AddComponent(&engi.UnpauseComponent{})
	m.menuEntities = append(m.menuEntities, menuBackground)
	m.World.AddEntity(menuBackground)

	// - box
	menuWidth := (engi.Width() - 2*2*menuPadding) / 2
	menuHeight := (engi.Height() - 2*2*menuPadding) / 2

	menuEntity := helpers.GenerateSquare(
		MenuColorBox, MenuColorBox,
		menuWidth, menuHeight,
		menuPadding, menuPadding,
		engi.HUDGround+1,
	)
	menuEntity.AddComponent(&engi.UnpauseComponent{})
	m.menuEntities = append(m.menuEntities, menuEntity)
	m.World.AddEntity(menuEntity)

	// - items - font
	itemFont := (&engi.Font{URL: "Roboto-Regular.ttf", Size: 64, FG: MenuColorItemForeground})
	if err := itemFont.CreatePreloaded(); err != nil {
		log.Fatalln("Could not load font:", err)
	}
	labelFontScale := float32(18 / itemFont.Size)

	// - items - entities
	offsetY := float32(menuPadding + menuItemPadding)
	for itemID, item := range m.items {
		var menuColorItemBackground = MenuColorItemBackground
		if itemID == m.menuFocus {
			menuColorItemBackground = MenuColorItemBackgroundFocus
		}

		menuItemBackground := helpers.GenerateSquare(
			menuColorItemBackground, menuColorItemBackground,
			menuWidth-2*menuItemPadding, menuItemHeight,
			menuItemOffsetX, offsetY,
			engi.HUDGround+2,
		)
		menuItemBackground.AddComponent(&engi.UnpauseComponent{})
		m.menuEntities = append(m.menuEntities, menuItemBackground)
		m.World.AddEntity(menuItemBackground)

		menuItemLabel := engi.NewEntity([]string{"RenderSystem"})
		menuItemLabelRender := &engi.RenderComponent{
			Display:      itemFont.Render(item.Text),
			Scale:        engi.Point{labelFontScale, labelFontScale},
			Label:        "",                 // TODO: unused?
			Priority:     engi.HUDGround + 3, //
			Transparency: 1,                  //
			Color:        0xffffff,           // TODO: unused?
		}
		menuItemLabel.AddComponent(menuItemLabelRender)
		menuItemLabel.AddComponent(&engi.SpaceComponent{
			engi.Point{
				menuItemOffsetX + (menuItemHeight-float32(itemFont.Size)*labelFontScale)/2,
				offsetY + menuItemFontPadding + (menuItemHeight-float32(itemFont.Size)*labelFontScale)/2,
			}, 20, 20}) // TODO: unused ??
		menuItemLabel.AddComponent(&engi.UnpauseComponent{})
		m.menuEntities = append(m.menuEntities, menuItemLabel)
		m.World.AddEntity(menuItemLabel)

		offsetY += menuItemHeight + menuItemPadding
	}
}
