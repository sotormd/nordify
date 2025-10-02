package nord

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type RGB [3]uint8
type Palette []RGB

type PaletteNotFoundError struct {
	Name string
}

func (e PaletteNotFoundError) Error() string {
	return fmt.Sprintf("palette %s not found", e.Name)
}

type PaletteReadError struct {
	Name string
}

func (e PaletteReadError) Error() string {
	return fmt.Sprintf("unable to read palette %s", e.Name)
}

type EmptyPaletteError struct {
	Name string
}

func (e EmptyPaletteError) Error() string {
	return fmt.Sprintf("palette %s is empty", e.Name)
}

type InvalidHexError struct {
	Hex string
}

func (e InvalidHexError) Error() string {
	return fmt.Sprintf("invalid hex code %s", e.Hex)
}

func constructPalettePath(palette string) string {
	path := filepath.Join("palettes", fmt.Sprintf("%s.json", palette))
	return path
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func paletteExists(palette string) bool {
	path := constructPalettePath(palette)
	if fileExists(path) {
		return true
	}
	return false
}

func isValidHex(hex string) bool {
	match, err := regexp.MatchString("^[0-9a-fA-F]", hex)

	if err != nil || !match {
		return false
	}

	return true
}

func hexToRGB(hex string) (RGB, error) {
	var color RGB

	for i := range []int{0, 1, 2} {
		v64, err := strconv.ParseInt(hex[i*2:i*2+2], 16, 16)

		if err != nil {
			return RGB{0, 0, 0}, err
		}

		v8 := uint8(v64)

		color[i] = v8
	}

	return color, nil
}

func readPalette(palette string) ([]string, error) {
	var rawPalette []string

	path := constructPalettePath(palette)
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(data, &rawPalette)

	return rawPalette, nil
}

func GetPalette(palette string) (Palette, error) {
	if !paletteExists(palette) {
		return nil, PaletteNotFoundError{Name: palette}
	}

	rawPalette, err := readPalette(palette)

	if err != nil {
		return nil, PaletteReadError{Name: palette}
	}

	var neatPalette Palette

	for _, hex := range rawPalette {
		hex = strings.TrimPrefix(strings.ToLower(hex), "#")
		if !isValidHex(hex) {
			return nil, InvalidHexError{Hex: hex}
		}
		rgb, err := hexToRGB(hex)
		if err != nil {
			return nil, InvalidHexError{Hex: hex}
		}

		neatPalette = append(neatPalette, rgb)
	}

	if len(neatPalette) < 1 {
		return nil, EmptyPaletteError{Name: palette}
	}

	return neatPalette, nil
}
