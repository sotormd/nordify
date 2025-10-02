package nord

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"runtime"
	"sync"
)

type ImageNotFoundError struct {
	Name string
}

func (e ImageNotFoundError) Error() string {
	return fmt.Sprintf("image %s not found", e.Name)
}

type ImageExistsError struct {
	Name string
}

func (e ImageExistsError) Error() string {
	return fmt.Sprintf("image %s already exists", e.Name)
}

type ImageOpenError struct {
	Name string
}

func (e ImageOpenError) Error() string {
	return fmt.Sprintf("unable to open image %s", e.Name)
}

type ImageReadError struct {
	Name string
}

func (e ImageReadError) Error() string {
	return fmt.Sprintf("unable to read image %s", e.Name)
}

type ImageCreateError struct {
	Name string
}

func (e ImageCreateError) Error() string {
	return fmt.Sprintf("unable to create image %s", e.Name)
}

func normalize(a RGB) (float64, float64, float64) {
	r, g, b := a[0], a[1], a[2]

	return float64(r) / 255.0, float64(g) / 255.0, float64(b) / 255.0
}

func linearize(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}

	return math.Pow((c+0.055)/1.055, 2.4)
}

func toXYZ(rLin, gLin, bLin float64) (float64, float64, float64) {
	x := rLin*0.4124564 + gLin*0.3575761 + bLin*0.1804375
	y := rLin*0.2126729 + gLin*0.7151522 + bLin*0.0721750
	z := rLin*0.0193339 + gLin*0.1191920 + bLin*0.9503041

	return x, y, z
}

func toD65(x, y, z float64) (float64, float64, float64) {
	x = x / 0.95047
	y = y / 1.00000
	z = z / 1.08883

	return x, y, z
}

func f(c float64) float64 {
	if c > 0.008856 {
		return math.Pow(c, 1.0/3.0)
	}

	return (7.787 * c) + (16.0 / 116.0)
}

func toLAB(x, y, z float64) (float64, float64, float64) {
	fx := f(x)
	fy := f(y)
	fz := f(z)

	L := (116.0 * fy) - 16.0
	a := 500.0 * (fx - fy)
	b := 200.0 * (fy - fz)

	return L, a, b
}

func colorDistanceLAB(a, b RGB) float64 {
	r1, g1, b1 := normalize(a)
	r2, g2, b2 := normalize(b)

	r1 = linearize(r1)
	g1 = linearize(g1)
	b1 = linearize(b1)

	r2 = linearize(r2)
	g2 = linearize(g2)
	b2 = linearize(b2)

	L1, a1, b1 := toLAB(toD65(toXYZ(r1, g1, b1)))
	L2, a2, b2 := toLAB(toD65(toXYZ(r2, g2, b2)))

	dL := float64(L1 - L2)
	da := float64(a1 - a2)
	db := float64(b1 - b2)

	return dL*dL + da*da + db*db
}

func colorDistance(a, b RGB) float64 {
	dr := float64(a[0] - b[0])
	dg := float64(a[1] - b[1])
	db := float64(a[2] - b[2])

	return 0.3*dr*dr + 0.59*dg*dg + 0.11*db*db
}

func nearestNord(a RGB, palette Palette) RGB {
	best := palette[0]
	d := colorDistanceLAB(a, best)

	for _, b := range palette {
		n := colorDistanceLAB(a, b)
		if n < d {
			best = b
			d = n
		}
	}

	return best
}

func RecolorImage(input string, output string, palette Palette) error {
	if !fileExists(input) {
		return ImageNotFoundError{Name: input}
	}

	if fileExists(output) {
		return ImageExistsError{Name: output}
	}

	inFile, err := os.Open(input)
	if err != nil {
		return ImageOpenError{Name: input}
	}
	defer inFile.Close()

	img, _, err := image.Decode(inFile)
	if err != nil {
		return ImageReadError{Name: input}
	}

	//    bounds := img.Bounds()
	//    outImg := image.NewRGBA(bounds)

	//    for y := bounds.Min.Y; y < bounds.Max.Y; y ++ {
	//    	for x := bounds.Min.X; x < bounds.Max.X; x ++ {
	//    		c := img.At(x, y)

	//    		r, g, b, a := c.RGBA()

	//    		r = r>>8
	//    		g = g>>8
	//    		b = b>>8
	//    		a = a>>8

	//    		c1 := RGB{uint8(r), uint8(g), uint8(b)}

	//    		c2 := nearestNord(c1, palette)

	//    		cout := color.RGBA{c2[0], c2[1], c2[2], uint8(a)}

	//    		outImg.Set(x, y, cout)
	//    	}
	//    }

	// since each pixel is independent,
	// we can use goroutines to speed up the
	// process of cin -> cout

	bounds := img.Bounds()
	outImg := image.NewRGBA(bounds)

	// limit the number of goroutines
	// to the number of available cpus
	sem := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup

	// loop through all pixels
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		wg.Add(1)
		sem <- struct{}{} // acquire a slot

		go func(y int) {
			defer wg.Done()
			defer func() { <-sem }() // release a slot
			// this ensures that sem <= runtime.NumCPU() at all time
			// so, the number of goroutines is also <= runtime.NumCPU() at all times

			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				c := img.At(x, y)

				r, g, b, a := c.RGBA()

				r = r >> 8
				g = g >> 8
				b = b >> 8
				a = a >> 8

				c1 := RGB{uint8(r), uint8(g), uint8(b)}

				c2 := nearestNord(c1, palette)

				cout := color.RGBA{c2[0], c2[1], c2[2], uint8(a)}

				outImg.Set(x, y, cout)
			}
		}(y)
	}

	wg.Wait()

	outFile, err := os.Create(output)
	if err != nil {
		return ImageCreateError{Name: output}
	}
	defer outFile.Close()

	png.Encode(outFile, outImg)

	return nil
}
