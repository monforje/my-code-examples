package avatar_test

import (
	"bytes"
	"image/png"
	"testing"

	"users/internal/generate/avatar"
)

func TestGenerate_Deterministic(t *testing.T) {
	id1 := avatar.Generate("same-seed")
	id2 := avatar.Generate("same-seed")
	if id1.Color != id2.Color {
		t.Error("Generate() not deterministic: different colors for same seed")
	}
	if id1.Grid != id2.Grid {
		t.Error("Generate() not deterministic: different grids for same seed")
	}
}

func TestGenerate_DifferentSeeds(t *testing.T) {
	id1 := avatar.Generate("seed-a")
	id2 := avatar.Generate("seed-b")
	if id1.Color == id2.Color && id1.Grid == id2.Grid {
		t.Error("Generate() same result for different seeds")
	}
}

func TestRender_ValidPNG(t *testing.T) {
	id := avatar.Generate("test")
	var buf bytes.Buffer
	if err := id.Render(&buf, 420); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	decoded, err := png.Decode(&buf)
	if err != nil {
		t.Fatalf("png.Decode() error = %v", err)
	}
	bounds := decoded.Bounds()
	if bounds.Dx() != 420 || bounds.Dy() != 420 {
		t.Errorf("Render() size = %dx%d, want 420x420", bounds.Dx(), bounds.Dy())
	}
}

func TestRender_DefaultSize(t *testing.T) {
	id := avatar.Generate("test")
	var buf bytes.Buffer
	if err := id.Render(&buf, 0); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	decoded, err := png.Decode(&buf)
	if err != nil {
		t.Fatalf("png.Decode() error = %v", err)
	}
	bounds := decoded.Bounds()
	if bounds.Dx() != 420 || bounds.Dy() != 420 {
		t.Errorf("Render(0) size = %dx%d, want 420x420 (default)", bounds.Dx(), bounds.Dy())
	}
}

func TestGrid_Symmetric(t *testing.T) {
	id := avatar.Generate("test")
	for row := 0; row < 5; row++ {
		for col := 0; col < 2; col++ {
			mirror := 4 - col
			if id.Grid[row][col] != id.Grid[row][mirror] {
				t.Errorf("Grid not symmetric: [%d][%d] = %v, [%d][%d] = %v",
					row, col, id.Grid[row][col], row, mirror, id.Grid[row][mirror])
			}
		}
	}
}
