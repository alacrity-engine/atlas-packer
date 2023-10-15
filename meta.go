package main

type AtlasMeta struct {
	Name            string      `yaml:"name"`
	Font            string      `yaml:"font"`
	Size            int16       `yaml:"size"`
	CharacterRanges [][2]string `yaml:"characterRanges"`
}
