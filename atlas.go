package main

import (
	"sort"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type fixedGlyph struct {
	dot     fixed.Point26_6
	frame   fixed.Rectangle26_6
	advance fixed.Int26_6
}

// makeSquareMapping finds an optimal glyph arrangement of the given runes, so that their common
// bounding box is as square as possible.
func makeSquareMapping(face font.Face, runes []rune, padding fixed.Int26_6) (map[rune]fixedGlyph, fixed.Rectangle26_6) {
	width := sort.Search(int(fixed.I(1024*1024)), func(i int) bool {
		width := fixed.Int26_6(i)
		_, bounds := makeMapping(face, runes, padding, width)
		return bounds.Max.X-bounds.Min.X >= bounds.Max.Y-bounds.Min.Y
	})
	return makeMapping(face, runes, padding, fixed.Int26_6(width))
}

// makeMapping arranges glyphs of the given runes into rows in such a way, that no glyph is located
// fully to the right of the specified width. Specifically, it places glyphs in a row one by one and
// once it reaches the specified width, it starts a new row.
func makeMapping(face font.Face, runes []rune, padding, width fixed.Int26_6) (map[rune]fixedGlyph, fixed.Rectangle26_6) {
	mapping := make(map[rune]fixedGlyph)
	bounds := fixed.Rectangle26_6{}

	dot := fixed.P(0, 0)

	for _, r := range runes {
		b, advance, ok := face.GlyphBounds(r)
		if !ok {
			continue
		}

		// this is important for drawing, artifacts arise otherwise
		frame := fixed.Rectangle26_6{
			Min: fixed.P(b.Min.X.Floor(), b.Min.Y.Floor()),
			Max: fixed.P(b.Max.X.Ceil(), b.Max.Y.Ceil()),
		}

		dot.X -= frame.Min.X
		frame = frame.Add(dot)

		mapping[r] = fixedGlyph{
			dot:     dot,
			frame:   frame,
			advance: advance,
		}
		bounds = bounds.Union(frame)

		dot.X = frame.Max.X

		// padding + align to integer
		dot.X += padding
		dot.X = fixed.I(dot.X.Ceil())

		// width exceeded, new row
		if frame.Max.X >= width {
			dot.X = 0
			dot.Y += face.Metrics().Ascent + face.Metrics().Descent

			// padding + align to integer
			dot.Y += padding
			dot.Y = fixed.I(dot.Y.Ceil())
		}
	}

	return mapping, bounds
}

func i2f(i fixed.Int26_6) float64 {
	return float64(i) / (1 << 6)
}
