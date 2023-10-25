package main

import (
	"fmt"
	"image"
	"image/draw"

	"github.com/alacrity-engine/core/math/geometry"
	codec "github.com/alacrity-engine/resource-codec"
	"github.com/golang/freetype/truetype"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/image/math/fixed"
)

func runeRangesToRunes(ranges [][2]string) ([]rune, error) {
	runeSet := map[rune]struct{}{}

	for _, runeRange := range ranges {
		if len([]rune(runeRange[0])) > 1 || len([]rune(runeRange[1])) > 1 {
			return nil, fmt.Errorf("incorrect range value: '%s'", runeRange)
		}

		if len([]rune(runeRange[0])) <= 0 || len([]rune(runeRange[1])) <= 0 {
			return nil, fmt.Errorf("incorrect range value: '%s'", runeRange)
		}

		a := []rune(runeRange[0])[0]
		b := []rune(runeRange[1])[0]

		for i := a; i <= b; i++ {
			runeSet[i] = struct{}{}
		}
	}

	runeArr := make([]rune, 0, len(runeSet))

	for run := range runeSet {
		runeArr = append(runeArr, run)
	}

	return runeArr, nil
}

func (atlasMeta AtlasMeta) ToAtlasData(resourceFile *bolt.DB) (*codec.AtlasData, error) {
	var fontData []byte

	err := resourceFile.View(func(tx *bolt.Tx) error {
		buck := tx.Bucket([]byte("fonts"))

		if buck == nil {
			return fmt.Errorf("fonts bucket not found")
		}

		fontData = buck.Get([]byte(atlasMeta.Font))

		if fontData == nil {
			return fmt.Errorf(
				"the '%s' font not found", atlasMeta.Font)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	ttFont, err := truetype.Parse(fontData)

	if err != nil {
		return nil, err
	}

	face := truetype.NewFace(ttFont, &truetype.Options{
		Size:              float64(atlasMeta.Size),
		GlyphCacheEntries: 1,
	})

	runes, err := runeRangesToRunes(atlasMeta.CharacterRanges)

	if err != nil {
		return nil, err
	}

	runes = append(runes, ' ')
	fixedMapping, fixedBounds := makeSquareMapping(
		face, runes, fixed.I(2))
	atlasImg := image.NewRGBA(image.Rect(
		fixedBounds.Min.X.Floor(),
		fixedBounds.Min.Y.Floor(),
		fixedBounds.Max.X.Ceil(),
		fixedBounds.Max.Y.Ceil(),
	))

	for r, fg := range fixedMapping {
		if dr, mask, maskp, _, ok := face.Glyph(fg.dot, r); ok {
			draw.Draw(atlasImg, dr, mask, maskp, draw.Src)
		}
	}

	bounds := geometry.R(
		i2f(fixedBounds.Min.X),
		i2f(fixedBounds.Min.Y),
		i2f(fixedBounds.Max.X),
		i2f(fixedBounds.Max.Y),
	)

	mapping := make(map[rune]codec.GlyphData)
	var maxHeight float64

	for r, fg := range fixedMapping {
		glyph := codec.GlyphData{
			Dot: geometry.V(
				i2f(fg.dot.X),
				bounds.Max.Y-(i2f(fg.dot.Y)-bounds.Min.Y),
			),
			Frame: geometry.R(
				i2f(fg.frame.Min.X),
				bounds.Max.Y-(i2f(fg.frame.Min.Y)-bounds.Min.Y),
				i2f(fg.frame.Max.X),
				bounds.Max.Y-(i2f(fg.frame.Max.Y)-bounds.Min.Y),
			).Norm(),
			Advance: i2f(fg.advance),
		}

		if glyph.Frame.H() > maxHeight {
			maxHeight = glyph.Frame.H()
		}

		mapping[r] = glyph
	}

	atlasPic, err := codec.NewPictureFromImage(atlasImg)

	if err != nil {
		return nil, err
	}

	return &codec.AtlasData{
		Glyphs:    mapping,
		SymbolSet: atlasPic,
		Size:      int32(atlasMeta.Size),
		FontName:  atlasMeta.Font,
		MaxHeight: maxHeight,
	}, nil
}
