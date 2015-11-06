package helpers

import (
	"github.com/paked/engi"
	"image"
	"image/color"
)

// GenerateSquare creates a square, alternating two colors, with given size and priority level
func GenerateSquare(c1, c2 color.Color, w, h float32, offX, offY float32, priority engi.PriorityLevel, requirements ...string) *engi.Entity {
	rect := image.Rect(0, 0, int(w), int(h))
	img := image.NewNRGBA(rect)
	for i := rect.Min.X; i < rect.Max.X; i++ {
		for j := rect.Min.Y; j < rect.Max.Y; j++ {
			if i%40 > 20 {
				if j%40 > 20 {
					img.Set(i, j, c1)
				} else {
					img.Set(i, j, c2)
				}
			} else {
				if j%40 > 20 {
					img.Set(i, j, c2)
				} else {
					img.Set(i, j, c1)
				}
			}
		}
	}
	bgTexture := engi.NewImageObject(img)
	field := engi.NewEntity(append([]string{"RenderSystem"}, requirements...))
	fieldRender := engi.NewRenderComponent(engi.NewRegion(engi.NewTexture(bgTexture), 0, 0, int(w), int(h)), engi.Point{1, 1}, "")
	fieldRender.Priority = priority
	fieldSpace := &engi.SpaceComponent{engi.Point{offX, offY}, w, h}
	field.AddComponent(fieldRender)
	field.AddComponent(fieldSpace)
	return field
}
