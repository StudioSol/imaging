package imaging

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"sort"
	"testing"
)

// The golden hashes in testdata/golden_v1.5.0.json were generated with the
// unmodified upstream v1.5.0 code (go test -run TestGoldenV150 -update-golden-v150).
// They guarantee that the functions used by downstream projects keep
// producing exactly the same pixels as v1.5.0.

const goldenV150File = "testdata/golden_v1.5.0.json"

var updateGoldenV150 = flag.Bool("update-golden-v150", false, "regenerate "+goldenV150File)

func goldenHash(img *image.NRGBA) string {
	h := sha256.New()
	fmt.Fprintf(h, "%v|%d|", img.Rect, img.Stride)
	h.Write(img.Pix)
	return hex.EncodeToString(h.Sum(nil))
}

func goldenSources(t *testing.T) map[string]image.Image {
	srcs := map[string]image.Image{}
	for _, name := range []string{"branches.png", "branches.jpg", "flowers.png", "flowers_small.png", "orientation_6.jpg"} {
		img, err := Open("testdata/" + name)
		if err != nil {
			t.Fatalf("open %s: %v", name, err)
		}
		srcs[name] = img
	}

	// Synthetic images covering the other scanner code paths.
	rgba := image.NewRGBA(image.Rect(-3, 5, 97, 68))
	gray := image.NewGray(image.Rect(0, 0, 61, 89))
	pal := image.NewPaletted(image.Rect(2, 2, 72, 52), nil)
	for i := 0; i < 256; i++ {
		pal.Palette = append(pal.Palette, color.NRGBA{uint8(i), uint8(255 - i), uint8(i * 7), uint8(i*3 + 1)})
	}
	for i := range rgba.Pix {
		rgba.Pix[i] = uint8(i * 31)
		if i%4 == 3 {
			rgba.Pix[i] = 0xff
			if i%12 == 3 {
				rgba.Pix[i] = uint8(i)
			}
			// keep premultiplied values valid
			for k := 1; k <= 3; k++ {
				if rgba.Pix[i-k] > rgba.Pix[i] {
					rgba.Pix[i-k] = rgba.Pix[i]
				}
			}
		}
	}
	for i := range gray.Pix {
		gray.Pix[i] = uint8(i * 13)
	}
	for i := range pal.Pix {
		pal.Pix[i] = uint8(i * 17)
	}
	srcs["synthetic_rgba"] = rgba
	srcs["synthetic_gray"] = gray
	srcs["synthetic_paletted"] = pal
	return srcs
}

func goldenResults(t *testing.T) map[string]string {
	res := map[string]string{}
	for name, img := range goldenSources(t) {
		add := func(op string, out *image.NRGBA) { res[name+"/"+op] = goldenHash(out) }
		b := img.Bounds()

		// Functions used by PalcoMP3.
		add("Fill_100x60_Center_Lanczos", Fill(img, 100, 60, Center, Lanczos))
		add("Fill_640x335_Center_Lanczos", Fill(img, 640, 335, Center, Lanczos))
		add("Fill_50x50_Center_Lanczos", Fill(img, 50, 50, Center, Lanczos))
		add("Fill_33x71_Center_Lanczos", Fill(img, 33, 71, Center, Lanczos))
		add("Fill_1200x630_Center_Lanczos", Fill(img, 1200, 630, Center, Lanczos))
		add("Resize_77x0_Lanczos", Resize(img, 77, 0, Lanczos))
		add("Resize_600x0_Lanczos", Resize(img, 600, 0, Lanczos))
		add("Crop_inner", Crop(img, image.Rect(3, 5, b.Dx()/2, b.Dy()/2)))
		add("Crop_full_width", Crop(img, image.Rect(0, 7, b.Dx(), b.Dy()-3)))
		add("Rotate90", Rotate90(img))
		add("Rotate180", Rotate180(img))
		add("Rotate270", Rotate270(img))

		// Other common functions, as extra safety.
		add("Fill_100x60_TopLeft_Linear", Fill(img, 100, 60, TopLeft, Linear))
		add("Fit_80x80_Lanczos", Fit(img, 80, 80, Lanczos))
		add("Thumbnail_64x64_Lanczos", Thumbnail(img, 64, 64, Lanczos))
		add("Resize_40x30_NearestNeighbor", Resize(img, 40, 30, NearestNeighbor))
		add("Grayscale", Grayscale(img))
		add("Blur_1.5", Blur(img, 1.5))
		add("Sharpen_1.5", Sharpen(img, 1.5))
		add("AdjustContrast_20", AdjustContrast(img, 20))
	}
	return res
}

func TestGoldenV150(t *testing.T) {
	got := goldenResults(t)

	if *updateGoldenV150 {
		data, err := json.MarshalIndent(got, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := ioutil.WriteFile(goldenV150File, append(data, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
		return
	}

	data, err := ioutil.ReadFile(goldenV150File)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{}
	if err := json.Unmarshal(data, &want); err != nil {
		t.Fatal(err)
	}
	if len(want) != len(got) {
		t.Fatalf("golden has %d entries, got %d", len(want), len(got))
	}
	keys := make([]string, 0, len(want))
	for k := range want {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if got[k] != want[k] {
			t.Errorf("%s: result differs from v1.5.0", k)
		}
	}
}
