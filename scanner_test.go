package imaging

import (
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"testing"
)

func TestScanner(t *testing.T) {
	rect := image.Rect(-1, -1, 15, 15)
	colors := palette.Plan9
	testCases := []struct {
		name string
		img  image.Image
	}{
		{
			name: "NRGBA",
			img:  makeNRGBAImage(rect, colors),
		},
		{
			name: "NRGBA64",
			img:  makeNRGBA64Image(rect, colors),
		},
		{
			name: "RGBA",
			img:  makeRGBAImage(rect, colors),
		},
		{
			name: "RGBA64",
			img:  makeRGBA64Image(rect, colors),
		},
		{
			name: "Gray",
			img:  makeGrayImage(rect, colors),
		},
		{
			name: "Gray16",
			img:  makeGray16Image(rect, colors),
		},
		{
			name: "YCbCr-444",
			img:  makeYCbCrImage(rect, colors, image.YCbCrSubsampleRatio444),
		},
		{
			name: "YCbCr-422",
			img:  makeYCbCrImage(rect, colors, image.YCbCrSubsampleRatio422),
		},
		{
			name: "YCbCr-420",
			img:  makeYCbCrImage(rect, colors, image.YCbCrSubsampleRatio420),
		},
		{
			name: "YCbCr-440",
			img:  makeYCbCrImage(rect, colors, image.YCbCrSubsampleRatio440),
		},
		{
			name: "YCbCr-410",
			img:  makeYCbCrImage(rect, colors, image.YCbCrSubsampleRatio410),
		},
		{
			name: "YCbCr-411",
			img:  makeYCbCrImage(rect, colors, image.YCbCrSubsampleRatio411),
		},
		{
			name: "Paletted",
			img:  makePalettedImage(rect, colors),
		},
		{
			name: "Alpha",
			img:  makeAlphaImage(rect, colors),
		},
		{
			name: "Alpha16",
			img:  makeAlpha16Image(rect, colors),
		},
		{
			name: "Generic",
			img:  makeGenericImage(rect, colors),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.img.Bounds()
			s := newScanner(tc.img)
			for y := r.Min.Y; y < r.Max.Y; y++ {
				buf := make([]byte, r.Dx()*4)
				s.scan(0, y-r.Min.Y, r.Dx(), y+1-r.Min.Y, buf)
				wantBuf := readRow(tc.img, y)
				if !compareBytes(buf, wantBuf, 1) {
					fmt.Println(tc.img)
					t.Fatalf("scan horizontal line (y=%d): got %v want %v", y, buf, wantBuf)
				}
			}
			for x := r.Min.X; x < r.Max.X; x++ {
				buf := make([]byte, r.Dy()*4)
				s.scan(x-r.Min.X, 0, x+1-r.Min.X, r.Dy(), buf)
				wantBuf := readColumn(tc.img, x)
				if !compareBytes(buf, wantBuf, 1) {
					t.Fatalf("scan vertical line (x=%d): got %v want %v", x, buf, wantBuf)
				}
			}
		})
	}
}

func makeYCbCrImage(rect image.Rectangle, colors []color.Color, sr image.YCbCrSubsampleRatio) *image.YCbCr {
	img := image.NewYCbCr(rect, sr)
	j := 0
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			iy := img.YOffset(x, y)
			ic := img.COffset(x, y)
			c := color.NRGBAModel.Convert(colors[j]).(color.NRGBA)
			img.Y[iy], img.Cb[ic], img.Cr[ic] = color.RGBToYCbCr(c.R, c.G, c.B)
			j++
		}
	}
	return img
}

func makeNRGBAImage(rect image.Rectangle, colors []color.Color) *image.NRGBA {
	img := image.NewNRGBA(rect)
	fillDrawImage(img, colors)
	return img
}

func makeNRGBA64Image(rect image.Rectangle, colors []color.Color) *image.NRGBA64 {
	img := image.NewNRGBA64(rect)
	fillDrawImage(img, colors)
	return img
}

func makeRGBAImage(rect image.Rectangle, colors []color.Color) *image.RGBA {
	img := image.NewRGBA(rect)
	fillDrawImage(img, colors)
	return img
}

func makeRGBA64Image(rect image.Rectangle, colors []color.Color) *image.RGBA64 {
	img := image.NewRGBA64(rect)
	fillDrawImage(img, colors)
	return img
}

func makeGrayImage(rect image.Rectangle, colors []color.Color) *image.Gray {
	img := image.NewGray(rect)
	fillDrawImage(img, colors)
	return img
}

func makeGray16Image(rect image.Rectangle, colors []color.Color) *image.Gray16 {
	img := image.NewGray16(rect)
	fillDrawImage(img, colors)
	return img
}

func makePalettedImage(rect image.Rectangle, colors []color.Color) *image.Paletted {
	img := image.NewPaletted(rect, colors)
	fillDrawImage(img, colors)
	return img
}

func makeAlphaImage(rect image.Rectangle, colors []color.Color) *image.Alpha {
	img := image.NewAlpha(rect)
	fillDrawImage(img, colors)
	return img
}

func makeAlpha16Image(rect image.Rectangle, colors []color.Color) *image.Alpha16 {
	img := image.NewAlpha16(rect)
	fillDrawImage(img, colors)
	return img
}

func makeGenericImage(rect image.Rectangle, colors []color.Color) image.Image {
	img := image.NewRGBA(rect)
	fillDrawImage(img, colors)
	type genericImage struct{ *image.RGBA }
	return &genericImage{img}
}

func fillDrawImage(img draw.Image, colors []color.Color) {
	colorsNRGBA := make([]color.NRGBA, len(colors))
	for i, c := range colors {
		nrgba := color.NRGBAModel.Convert(c).(color.NRGBA)
		nrgba.A = uint8(i % 256)
		colorsNRGBA[i] = nrgba
	}
	rect := img.Bounds()
	i := 0
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			img.Set(x, y, colorsNRGBA[i])
			i++
		}
	}
}

func readRow(img image.Image, y int) []uint8 {
	row := make([]byte, img.Bounds().Dx()*4)
	i := 0
	for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
		c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
		row[i+0] = c.R
		row[i+1] = c.G
		row[i+2] = c.B
		row[i+3] = c.A
		i += 4
	}
	return row
}

func readColumn(img image.Image, x int) []uint8 {
	column := make([]byte, img.Bounds().Dy()*4)
	i := 0
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
		column[i+0] = c.R
		column[i+1] = c.G
		column[i+2] = c.B
		column[i+3] = c.A
		i += 4
	}
	return column
}

func TestScannerPalettedIndexOutOfPalette(t *testing.T) {
	// Paletted image whose pixels reference indexes beyond the palette length.
	img := image.NewPaletted(image.Rect(0, 0, 3, 2), color.Palette{
		color.NRGBA{0x10, 0x20, 0x30, 0xff},
		color.NRGBA{0x40, 0x50, 0x60, 0x80},
	})
	img.Pix = []uint8{0, 1, 2, 255, 1, 0}

	s := newScanner(img)
	dst := make([]uint8, 3*2*4)
	s.scan(0, 0, 3, 2, dst)

	want := []uint8{
		0x10, 0x20, 0x30, 0xff, 0x40, 0x50, 0x60, 0x80, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x40, 0x50, 0x60, 0x80, 0x10, 0x20, 0x30, 0xff,
	}
	if !compareBytes(dst, want, 0) {
		t.Fatalf("got %v want %v", dst, want)
	}
}

func processAllNoPanic(t *testing.T, img image.Image) {
	for name, fn := range map[string]func(image.Image) *image.NRGBA{
		"Grayscale": Grayscale,
		"Clone":     Clone,
		"Rotate90":  Rotate90,
		"Fill":      func(i image.Image) *image.NRGBA { return Fill(i, 50, 40, Center, Lanczos) },
		"Resize":    func(i image.Image) *image.NRGBA { return Resize(i, 70, 0, Lanczos) },
		"Crop":      func(i image.Image) *image.NRGBA { return Crop(i, image.Rect(1, 1, 30, 20)) },
	} {
		if out := fn(img); out == nil || out.Bounds().Empty() {
			t.Errorf("%s: unexpected empty result", name)
		}
	}
}

func TestMalformedPalettedTIFF(t *testing.T) {
	// testdata/cve_2023_36308_poc.tiff is a TIFF whose color map has fewer
	// entries than the pixel values reference (CVE-2023-36308). Depending on
	// the golang.org/x/image version, decoding either fails or returns a
	// paletted image with a short palette; neither may panic.
	img, err := Open("testdata/cve_2023_36308_poc.tiff")
	if err != nil {
		t.Logf("decode rejected the malformed TIFF: %v", err)
	} else {
		processAllNoPanic(t, img)
	}

	// The same malformed image built in memory, independent of the decoder.
	pal := image.NewPaletted(image.Rect(0, 0, 150, 103), make(color.Palette, 65))
	for i := range pal.Palette {
		pal.Palette[i] = color.NRGBA{uint8(i), uint8(i * 2), uint8(i * 3), 0xff}
	}
	for i := range pal.Pix {
		pal.Pix[i] = uint8(i)
	}
	processAllNoPanic(t, pal)
}
