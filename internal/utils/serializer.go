package utils

import (
	"encoding/hex"
	"errors"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkb"
	"github.com/paulmach/orb/geojson"
)

func EWKBHexToGeoJSON(hexString string) (string, error) {
	if hexString == "" {
		return "", errors.New("geometry vazia")
	}

	bin, err := hex.DecodeString(hexString)
	if err != nil {
		return "", err
	}

	geom, err := wkb.Unmarshal(bin)
	if err != nil {
		return "", err
	}

	feature := geojson.NewFeature(geom.(orb.Geometry))
	js, err := feature.MarshalJSON()
	if err != nil {
		return "", err
	}

	return string(js), nil
}
