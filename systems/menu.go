package systems

import (
	"github.com/EtienneBruines/bcigame/helpers"
	"github.com/paked/engi"
	"image/color"
	"log"
	"os"
)

var (
	MenuColorBackground          = color.NRGBA{255, 255, 255, 125}
	MenuColorBox                 = color.NRGBA{180, 180, 180, 255}
	MenuColorItemBackground      = color.NRGBA{0, 0, 0, 125}
	MenuColorItemBackgroundFocus = color.NRGBA{64, 96, 0, 200}
	MenuColorItemForeground      = engi.Color{255, 255, 255, 255}

	menuItemHeight      = float32(50)
	menuItemOffsetX     = float32(menuPadding + menuItemPadding)
	menuItemFontPadding = float32(2)
	menuItemPadding     = float32(5)
	menuPadding         = float32(100)
)

type MenuItem struct {
	Text     string
	Callback func()
	SubItems []*MenuItem
	Parent   *MenuItem

	menuBackground *engi.Entity
	menuLabel      *engi.Entity
}

type Menu struct {
	*engi.System

	defaultBackground *engi.RenderComponent
	focusBackground   *engi.RenderComponent

	menuActive       bool
	menuEntities     []*engi.Entity
	menuItemEntities []*engi.Entity
	menuFocus        int
	items            []*MenuItem
	itemSelected     *MenuItem
}

func (*Menu) Type() string {
	return "MenuSystem"
}

func (m *Menu) New() {
	m.System = engi.NewSystem()

	specificLevel := &MenuItem{Text: "Play specific level ..."}

	callbackGenerator := func(l *Level) func() {
		msg := MazeMessage{l.Name}
		return func() {
			m.closeMenu()
			engi.Mailbox.Dispatch(msg)
		}
	}

	specificLevel.Callback = func() {
		specificLevel.SubItems = make([]*MenuItem, 0)
		for _, l := range ActiveMazeSystem.levels {
			specificLevel.SubItems = append(specificLevel.SubItems, &MenuItem{Text: l.Name,
				Callback: callbackGenerator(&l)})
		}
	}

	e := engi.NewEntity([]string{m.Type()})
	e.AddComponent(&engi.UnpauseComponent{})
	m.AddEntity(e)
	m.items = []*MenuItem{
		{Text: "Random Level", Callback: func() {
			m.closeMenu()
			engi.Mailbox.Dispatch(MazeMessage{})
		}},
		specificLevel,
		{Text: "Calibrate", Callback: func() {
			m.closeMenu()
			engi.Mailbox.Dispatch(CalibrateMessage{true})
		}},
		{Text: "Exit", Callback: func() {
			os.Exit(0)
		}},
	}

	// TODO: handle resizing of window
	menuWidth := (engi.Width() - 2*menuPadding)

	m.focusBackground = helpers.GenerateSquareComonent(
		MenuColorItemBackgroundFocus, MenuColorItemBackgroundFocus,
		menuWidth-2*menuItemPadding, menuItemHeight,
		engi.HUDGround+2,
	)

	m.defaultBackground = helpers.GenerateSquareComonent(
		MenuColorItemBackground, MenuColorItemBackground,
		menuWidth-2*menuItemPadding, menuItemHeight,
		engi.HUDGround+2,
	)
}

func (m *Menu) Update(entity *engi.Entity, dt float32) {
	// Check for ESCAPE
	if engi.Keys.Get(engi.Escape).JustPressed() {
		if m.menuActive {
			if m.itemSelected == nil {
				m.closeMenu()
			} else {
				selected := m.itemSelected
				m.closeMenu()
				m.itemSelected = selected.Parent
				m.openMenu()
			}
		} else {
			m.openMenu()
			return // so wait one frame before the menu gets to be used
		}
	}

	if m.menuActive {
		var updated bool
		var oldFocus = m.menuFocus
		var itemList []*MenuItem
		if m.itemSelected == nil {
			itemList = m.items
		} else {
			itemList = m.itemSelected.SubItems
		}

		if engi.Keys.Get(engi.ArrowDown).JustPressed() {
			m.menuFocus++
			if m.menuFocus >= len(itemList) {
				m.menuFocus = 0
			}
			updated = true
		} else if engi.Keys.Get(engi.ArrowUp).JustPressed() {
			m.menuFocus--
			if m.menuFocus < 0 {
				m.menuFocus = len(itemList) - 1
			}
			updated = true
		}

		if updated {
			// note that these replace the old RenderComponents
			itemList[oldFocus].menuBackground.AddComponent(m.defaultBackground)
			itemList[m.menuFocus].menuBackground.AddComponent(m.focusBackground)
		}

		if engi.Keys.Get(engi.Space).JustPressed() || engi.Keys.Get(engi.Enter).JustPressed() {
			itemList[m.menuFocus].Callback()
			if len(itemList[m.menuFocus].SubItems) > 0 {
				m.closeMenu()
				m.itemSelected = m.items[m.menuFocus]
				m.openMenu()
			}
		}
	}
}

func (m *Menu) closeMenu() {
	// Unpause everything
	engi.Mailbox.Dispatch(engi.PauseMessage{false})

	// Remove all entities
	for _, e := range m.menuEntities {
		m.World.RemoveEntity(e)
	}

	m.itemSelected = nil
	m.menuActive = false
}

func (m *Menu) openMenu() {
	// Pause everything
	engi.Mailbox.Dispatch(engi.PauseMessage{true})

	m.menuFocus = 0

	// Create the visual menu
	// - background
	backgroundWidth := engi.Width()
	backgroundHeight := engi.Height()

	menuBackground := helpers.GenerateSquare(
		MenuColorBackground, MenuColorBackground,
		backgroundWidth, backgroundHeight,
		0, 0,
		engi.HUDGround,
		"AudioSystem",
	)
	menuBackground.AddComponent(&engi.UnpauseComponent{})
	menuBackground.AddComponent(&engi.AudioComponent{File: "click_x.wav", Repeat: false, Background: true})
	m.menuEntities = append(m.menuEntities, menuBackground)
	m.World.AddEntity(menuBackground)

	// - box
	menuWidth := (engi.Width() - 2*menuPadding)
	menuHeight := (engi.Height() - 2*menuPadding)

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
	labelFontScale := float32(36 / itemFont.Size)

	// - items - entities
	offsetY := float32(menuPadding + menuItemPadding)

	var itemList []*MenuItem
	if m.itemSelected == nil {
		itemList = m.items
	} else {
		itemList = m.itemSelected.SubItems
	}

	for itemID, item := range itemList {
		item.menuBackground = engi.NewEntity([]string{"RenderSystem"})
		if itemID == m.menuFocus {
			item.menuBackground.AddComponent(m.focusBackground)
		} else {
			item.menuBackground.AddComponent(m.defaultBackground)
		}
		item.menuBackground.AddComponent(&engi.SpaceComponent{
			engi.Point{menuItemOffsetX, offsetY}, menuWidth - 2*menuItemPadding, menuItemHeight})
		item.menuBackground.AddComponent(&engi.UnpauseComponent{})
		m.menuEntities = append(m.menuEntities, item.menuBackground)
		m.World.AddEntity(item.menuBackground)

		item.menuLabel = engi.NewEntity([]string{"RenderSystem"})
		menuItemLabelRender := &engi.RenderComponent{
			Display:      itemFont.Render(item.Text),
			Scale:        engi.Point{labelFontScale, labelFontScale},
			Transparency: 1,
			Color:        0xffffff,
		}
		menuItemLabelRender.SetPriority(engi.HUDGround + 3)
		item.menuLabel.AddComponent(menuItemLabelRender)
		item.menuLabel.AddComponent(&engi.SpaceComponent{
			Position: engi.Point{
				menuItemOffsetX + (menuItemHeight-float32(itemFont.Size)*labelFontScale)/2,
				offsetY + menuItemFontPadding + (menuItemHeight-float32(itemFont.Size)*labelFontScale)/2,
			}})
		item.menuLabel.AddComponent(&engi.UnpauseComponent{})
		m.menuEntities = append(m.menuEntities, item.menuLabel)
		m.World.AddEntity(item.menuLabel)

		offsetY += menuItemHeight + menuItemPadding
	}

	m.menuActive = true
}
