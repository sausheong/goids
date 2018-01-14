package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"sort"

	"github.com/llgcode/draw2d/draw2dimg"
)

// parameters
var windowWidth, windowHeight = 800, 600
var goidSize = 3
var goidColor = color.RGBA{200, 200, 100, 255} // gray, 50% transparency
var populationSize = 150
var loops = 100
var numNeighbours = 7
var separationFactor = float64(goidSize * 5)
var coherenceFactor = 8

func main() {
	clearScreen()
	hideCursor()

	goids := make([]*Goid, 0)
	for i := 0; i < populationSize; i++ {
		g := createRandomGoid()
		goids = append(goids, &g)
	}

	for i := 0; i < loops; i++ {
		move(goids)
		frame := draw(goids)
		printImage(frame.SubImage(frame.Rect))
		fmt.Printf("\nLoop: %d", i)

	}
	showCursor()
}

// Goid represents a drawn goid
type Goid struct {
	X     int // position
	Y     int
	Vx    int // velocity
	Vy    int
	R     int // radius
	Color color.Color
}

func createRandomGoid() (g Goid) {
	g = Goid{
		X:     rand.Intn(windowWidth),
		Y:     rand.Intn(windowHeight),
		Vx:    rand.Intn(goidSize),
		Vy:    rand.Intn(goidSize),
		R:     goidSize,
		Color: goidColor,
	}
	return
}

// find the nearest neighbours
func (g *Goid) nearestNeighbours(goids []*Goid) (neighbours []Goid) {
	neighbours = make([]Goid, len(goids))
	for _, goid := range goids {
		neighbours = append(neighbours, *goid)
	}
	sort.SliceStable(neighbours, func(i, j int) bool {
		return g.distance(neighbours[i]) < g.distance(neighbours[j])
	})
	return
}

// distance between 2 goids
func (g *Goid) distance(n Goid) float64 {
	x := g.X - n.X
	y := g.Y - n.Y
	return math.Sqrt(float64(x*x + y*y))

}

// move the goids with the 3 classic boid rules
func move(goids []*Goid) {
	for _, goid := range goids {
		neighbours := goid.nearestNeighbours(goids)
		separate(goid, neighbours)
		align(goid, neighbours)
		cohere(goid, neighbours)

		stayInWindow(goid)
	}
}

// if goid goes out of the window frame it comes back on the other side
func stayInWindow(goid *Goid) {
	if goid.X < 0 {
		goid.X = windowWidth + goid.X
	} else if goid.X > windowWidth {
		goid.X = windowWidth - goid.X
	}
	if goid.Y < 0 {
		goid.Y = windowHeight + goid.Y
	} else if goid.Y > windowHeight {
		goid.Y = windowHeight - goid.Y
	}
}

// steer to avoid crowding local goids
func separate(g *Goid, neighbours []Goid) {
	x, y := 0, 0
	for _, n := range neighbours[0:numNeighbours] {
		if g.distance(n) < separationFactor {
			x += g.X - n.X
			y += g.Y - n.Y
		}
	}
	g.Vx = x
	g.Vy = y
	g.X += x
	g.Y += y
}

// steer towards the average heading of local goids
func align(g *Goid, neighbours []Goid) {
	x, y := 0, 0
	for _, n := range neighbours[0:numNeighbours] {
		x += n.Vx
		y += n.Vy
	}
	dx, dy := x/numNeighbours, y/numNeighbours
	g.Vx += dx
	g.Vy += dy
	g.X += dx
	g.Y += dy
}

// steer to move toward the average position of local goids
func cohere(g *Goid, neighbours []Goid) {
	x, y := 0, 0
	for _, n := range neighbours[0:numNeighbours] {
		x += n.X
		y += n.Y
	}
	dx, dy := ((x/numNeighbours)-g.X)/coherenceFactor, ((y/numNeighbours)-g.Y)/coherenceFactor
	g.Vx += dx
	g.Vy += dy
	g.X += dx
	g.Y += dy
}

// draw the goids
func draw(goids []*Goid) *image.RGBA {
	dest := image.NewRGBA(image.Rect(0, 0, windowWidth, windowHeight))
	gc := draw2dimg.NewGraphicContext(dest)
	for _, goid := range goids {
		gc.SetFillColor(goid.Color)
		gc.MoveTo(float64(goid.X), float64(goid.Y))
		gc.ArcTo(float64(goid.X), float64(goid.Y), float64(goid.R), float64(goid.R), 0, -math.Pi*2)
		gc.LineTo(float64(goid.X-goid.Vx), float64(goid.Y-goid.Vy))
		gc.Close()
		gc.Fill()
	}
	return dest
}

// ANSI escape sequence codes to perform action on terminal
func hideCursor() {
	fmt.Print("\033[?25l")
}

func showCursor() {
	fmt.Print("\x1b[?25h\n")
}

func clearScreen() {
	fmt.Print("\x1b[2J")
}

// this only works for iTerm!
func printImage(img image.Image) {
	var buf bytes.Buffer
	png.Encode(&buf, img)
	imgBase64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	fmt.Printf("\x1b[2;0H\x1b]1337;File=inline=1:%s\a", imgBase64Str)
}
