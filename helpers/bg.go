package helpers

import (
	"github.com/paked/engi"
	"github.com/paked/engi/ecs"
	"image"
	"image/color"
)

// GenerateSquare creates a square, alternating two colors, with given size and priority level
func GenerateSquare(c1, c2 color.NRGBA, w, h float32, offX, offY float32, priority engi.PriorityLevel, requirements ...string) *ecs.Entity {
	field := ecs.NewEntity(append([]string{"RenderSystem"}, requirements...))
	field.AddComponent(GenerateSquareComonent(c1, c2, w, h, priority))
	field.AddComponent(&engi.SpaceComponent{engi.Point{offX, offY}, w, h})
	return field
}

// GenerateSquareComonent creates a square, alternating two colors, with given size and priority levl
func GenerateSquareComonent(c1, c2 color.NRGBA, w, h float32, priority engi.PriorityLevel) *engi.RenderComponent {
	rect := image.Rect(0, 0, int(w), int(h))
	img := image.NewNRGBA(rect)
	if c1 == c2 {
		// Solid color
		for i := 0; i < len(img.Pix); i += 4 {
			img.Pix[i] = c1.R
			img.Pix[i+1] = c1.G
			img.Pix[i+2] = c1.B
			img.Pix[i+3] = c1.A
		}
	} else {
		// TODO: we can optimize this as well
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
	}

	bgTexture := engi.NewImageObject(img)
	fieldRender := engi.NewRenderComponent(engi.NewRegion(engi.NewTexture(bgTexture), 0, 0, int(w), int(h)), engi.Point{1, 1}, "")
	fieldRender.SetPriority(priority)

	return fieldRender
}
