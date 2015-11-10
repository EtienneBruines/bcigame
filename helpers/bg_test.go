package helpers_test

import (
	"fmt"
	"github.com/EtienneBruines/bcigame/helpers"
	"github.com/paked/engi"
	"image/color"
	"testing"
)

func BenchmarkSolidOne(b *testing.B) {
	engi.OpenHeadlessNoRun()

	c1 := color.NRGBA{255, 255, 255, 255}
	c2 := color.NRGBA{255, 255, 255, 255}
	w := float32(100)
	h := float32(100)
	priority := engi.Background
	var r *engi.RenderComponent

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		helpers.GenerateSquareComonent(c1, c2, w, h, priority)
	}
	fmt.Sprint(r)
}
