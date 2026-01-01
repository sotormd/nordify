package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"nordify/nord"
	"os"
	"path/filepath"
)

//go:embed assets/palettes/*.json
var embeddedPalettes embed.FS

func ensurePalettesDir(dir string) error {
	if _, err := os.Stat(dir); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	entries, err := embeddedPalettes.ReadDir("assets/palettes")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		data, err := embeddedPalettes.ReadFile("assets/palettes/" + entry.Name())
		if err != nil {
			return err
		}

		outPath := filepath.Join(dir, entry.Name())
		if err := os.WriteFile(outPath, data, 0o644); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to find home directory")
		os.Exit(1)
	}

	default_palette_dir := filepath.Join(home, ".config", "nordify", "palettes")

	parg := flag.String("p", "nord", "Palette to use")
	darg := flag.String("d", default_palette_dir, "Directory to look for palettes")

	flag.Parse()

	args := flag.Args()

	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: nordify input.png output.png")
		os.Exit(1)
	}

	input := args[0]
	output := args[1]

	palettes_dir := *darg
	if err := ensurePalettesDir(palettes_dir); err != nil {
		fmt.Fprintln(os.Stderr, "Unable to install default palettes")
		os.Exit(1)
	}

	palette, err := nord.GetPalette(palettes_dir, *parg)

	var pnf nord.PaletteNotFoundError
	var prd nord.PaletteReadError
	var phx nord.InvalidHexError
	if err != nil {
		if errors.As(err, &pnf) {
			fmt.Fprintf(os.Stderr, "Unable to find palette %s\n", pnf.Name)
		} else if errors.As(err, &prd) {
			fmt.Fprintf(os.Stderr, "Unable to read palette %s\n", prd.Name)
		} else if errors.As(err, &phx) {
			fmt.Fprintf(os.Stderr, "Invalid hex code %s\n", phx.Hex)
		}
		os.Exit(1)
	}

	err = nord.RecolorImage(input, output, palette)

	var inf nord.ImageNotFoundError
	var ird nord.ImageReadError
	var ict nord.ImageCreateError
	var iex nord.ImageExistsError
	if err != nil {
		if errors.As(err, &inf) {
			fmt.Fprintf(os.Stderr, "Unable to find image %s\n", inf.Name)
		} else if errors.As(err, &ird) {
			fmt.Fprintf(os.Stderr, "Unable to read image %s\n", ird.Name)
		} else if errors.As(err, &ict) {
			fmt.Fprintf(os.Stderr, "Unable to create image %s\n", ict.Name)
		} else if errors.As(err, &iex) {
			fmt.Fprintf(os.Stderr, "Image %s exists\n", iex.Name)
		}
		os.Exit(1)
	}
}
