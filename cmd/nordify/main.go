package main

import (
	"errors"
	"flag"
	"fmt"
	"nordify/nord"
	"os"
)

func main() {
	parg := flag.String("p", "nord", "Palette to use")

	flag.Parse()

	args := flag.Args()

	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: nordify input.png output.png")
		os.Exit(1)
	}

	input := args[0]
	output := args[1]

	palette, err := nord.GetPalette(*parg)

	var pnf nord.PaletteNotFoundError
	var prd nord.PaletteReadError
	var phx nord.InvalidHexError
	if err != nil {
		if errors.As(err, &pnf) {
			fmt.Fprintf(os.Stderr, "Invalid palette %s", pnf.Name)
		} else if errors.As(err, &prd) {
			fmt.Fprintf(os.Stderr, "Unable to read palette file %s", prd.Name)
		} else if errors.As(err, &phx) {
			fmt.Fprintf(os.Stderr, "Invalid hex code %s", phx.Hex)
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
			fmt.Fprintf(os.Stderr, "Image %s not found", inf.Name)
		} else if errors.As(err, &ird) {
			fmt.Fprintf(os.Stderr, "Unable to read image %s", ird.Name)
		} else if errors.As(err, &ict) {
			fmt.Fprintf(os.Stderr, "Unable to create image %s", ict.Name)
		} else if errors.As(err, &iex) {
			fmt.Fprintf(os.Stderr, "Image %s exists", iex.Name)
		}
		os.Exit(1)
	}
}
