package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
)

type imgObject struct {
	sY, eY, sX, eX int
}

// Object border color.
var blk = color.RGBA{0, 0, 0, 255}

func main() {
	imgFile, err := os.Open("go.png")
	if err != nil {
		log.Fatal(err)
	}
	defer imgFile.Close()

	m, _, err := image.Decode(imgFile)
	if err != nil {
		log.Fatal(err)
	}

	bounds := m.Bounds()
	imgNew := image.NewRGBA(m.Bounds())
	draw.Draw(imgNew, bounds, m, image.Point{}, draw.Over)

	backColor := color.RGBA{128, 255, 255, 255}
	// backColor := color.RGBA{106, 215, 229, 255}

	bgColor := getBGColor(imgNew)
	blkZone := getBlkZone(imgNew)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if imgNew.At(x, y) == bgColor {
				if inBlkZone(blkZone, x, y) {
					if !checkBlkBorders(imgNew, x, y) {
						imgNew.Set(x, y, backColor)
					}
				} else {
					imgNew.Set(x, y, backColor)
				}
			}
		}
	}

	imgFileNew, _ := os.Create("you_changed.png")
	defer imgFileNew.Close()
	png.Encode(imgFileNew, imgNew)

	// Remove small objects and crete new file from it.
	// you can specify width and height in px.
	removeObjFromImage(240, 30, imgNew)
}

func removeObjFromImage(w, h int, m *image.RGBA) {
	bounds := m.Bounds()
	bgColor := getBGColor(m)
	objZone := getObjZone(m, bgColor)
	filObjZone := filterObj(objZone, w, h)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if m.At(x, y) != bgColor {
				if inObjZone(filObjZone, x, y) {
					m.Set(x, y, bgColor)
				}
			}
		}
	}

	imgFileNew, _ := os.Create("you_changed2.png")
	defer imgFileNew.Close()
	png.Encode(imgFileNew, m)
}

// Gather all objects from image.
func getObjZone(m *image.RGBA, bg color.Color) []imgObject {
	bounds := m.Bounds()
	simj := []imgObject{}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if m.At(x, y) != bg && !inObjZone(simj, x, y) {
				obj := getObjborders(m, x, y, bg)
				simj = append(simj, obj)
				break
			}
		}
	}
	return simj
}

// Remove big or broken objects.
func filterObj(simj []imgObject, w, h int) []imgObject {
	i := 0
	for _, v := range simj {
		if v.eX-v.sX < w && v.eY-v.sY < h {
			simj[i] = v
			i++
		}
	}
	simj = simj[:i]
	i = 0
	for _, v := range simj {
		if v.eX-v.sX > 4 && v.eY-v.sY > 4 {
			simj[i] = v
			i++
		}
	}
	simj = simj[:i]
	return simj
}

// Check if background pixel inside object.
func inObjZone(simj []imgObject, x, y int) bool {
	for _, v := range simj {
		if x > v.sX && x < v.eX && y >= v.sY && y < v.eY {
			return true
		}
	}
	return false
}

// Gather object borders from coordinates.
func getObjborders(m *image.RGBA, x, y int, bg color.Color) imgObject {
	bounds := m.Bounds()
	imj := imgObject{
		sY: y,
		sX: x,
	}

	for k := x; k < bounds.Max.X; k++ {
		if m.At(k, y) == bg {
			imj.eX = k
			break
		}
	}

	ctr := imj.sX + (imj.eX-imj.sX)/2

	for k := y; k < bounds.Max.Y; k++ {
		if m.At(ctr, k) == bg {
			imj.eY = k
			break
		}
	}
	for k := imj.sY; k > bounds.Min.Y; k-- {
		if m.At(ctr, k) == bg {
			imj.sY = k
			break
		}
	}

	return imj
}

// Generate background color by checking top left and bottom left corners.
func getBGColor(m *image.RGBA) color.Color {
	mi := make(map[color.Color]int)
	bounds := m.Bounds()

	for y := bounds.Min.Y; y < 50; y++ {
		for x := bounds.Min.X; x < 50; x++ {
			mi[m.At(x, y)] += 1
		}
	}
	for y := bounds.Max.Y - 50; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < 50; x++ {
			mi[m.At(x, y)] += 1
		}
	}

	maxVal := 0
	var bg color.Color

	for i, k := range mi {
		if k > maxVal {
			maxVal = k
			bg = i
		}
	}

	return bg
}

// Gather all black lines from image on every Y axis.
func getBlkZone(m *image.RGBA) map[int][]int {
	msi := make(map[int][]int)
	bounds := m.Bounds()
	curEndPointX := bounds.Min.X

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if m.At(x, y) == blk && x > curEndPointX {
				blkLine := getBlkLine(m, x, y)
				curEndPointX = blkLine[1]
				msi[y] = append(msi[y], blkLine[0], blkLine[1])
			}
		}
		curEndPointX = bounds.Min.X
	}

	return msi
}

// Gather black line from coordinates.
func getBlkLine(m *image.RGBA, x int, y int) []int {
	bounds := m.Bounds()
	si := []int{}

	for k := x; k < bounds.Max.X; k++ {
		if m.At(k, y) != blk {
			return []int{x, k - 1}
		}
	}
	return si
}

// Check if background pixel between black borders on X axis.
func inBlkZone(msi map[int][]int, x int, y int) bool {
	si := msi[y]
	if len(si) > 2 {
		for i := 0; i < len(si)-1; i++ {
			if i%2 == 1 {
				if x > si[i] && x < si[i+1] {
					return true
				}
			}
		}
	}
	return false
}

// Check if background pixel surrounded by black borders.
func checkBlkBorders(m *image.RGBA, x int, y int) bool {
	bounds := m.Bounds()
	curPosX := x
	foundBorders := 0

	// Check bottom border.
	for curPosY := y; curPosY < bounds.Max.Y; curPosY++ {
		if m.At(x, curPosY) == blk {
			foundBorders += 1
			break
		}
	}
	// Check top border.
	for curPosY := y; curPosY > bounds.Min.Y; curPosY-- {
		if m.At(x, curPosY) == blk {
			foundBorders += 1
			break
		}
	}
	// Check bottom right diagonal border.
	for curPosY := y; curPosY < bounds.Max.Y; curPosY++ {
		if m.At(curPosX, curPosY) == blk {
			foundBorders += 1
			break
		}
		curPosX++
		if m.At(curPosX, curPosY) == blk {
			foundBorders += 1
			break
		}
	}
	// Check bottom left diagonal border.
	curPosX = x
	for curPosY := y; curPosY < bounds.Max.Y; curPosY++ {
		if m.At(curPosX, curPosY) == blk {
			foundBorders += 1
			break
		}
		curPosX--
		if m.At(curPosX, curPosY) == blk {
			foundBorders += 1
			break
		}
	}
	// Check top right diagonal border.
	curPosX = x
	for curPosY := y; curPosY > bounds.Min.Y; curPosY-- {
		if m.At(curPosX, curPosY) == blk {
			foundBorders += 1
			break
		}
		curPosX++
		if m.At(curPosX, curPosY) == blk {
			foundBorders += 1
			break
		}
	}
	// Check top left diagonal border.
	curPosX = x
	for curPosY := y; curPosY > bounds.Min.Y; curPosY-- {
		if m.At(curPosX, curPosY) == blk {
			foundBorders += 1
			break
		}
		curPosX--
		if m.At(curPosX, curPosY) == blk {
			foundBorders += 1
			break
		}
	}

	if foundBorders == 6 {
		return true
	}

	return false
}
