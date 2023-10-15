package main

import "gopkg.in/yaml.v2"

func ReadAtlasesData(data []byte) ([]AtlasMeta, error) {
	atlases := []AtlasMeta{}
	err := yaml.Unmarshal(data, &atlases)

	if err != nil {
		return nil, err
	}

	return atlases, nil
}
