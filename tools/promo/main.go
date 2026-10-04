// Command promo renders a looping CAD wireframe for the site header with Kavad.
// Run from the repository root with go -C tools/promo run . -out ../../.tmp/promo.
package main

import (
	"context"
	"math"

	"github.com/lestrrat-go/kavad"
	"github.com/lestrrat-go/kavad/svg"
)

const (
	width    = 800
	height   = 600
	periodMS = 6000
)

type show struct{}

func (show) Size() (int, int) { return width, height }

func (show) Start(context.Context) (*kavad.Run, error) {
	return kavad.NewRun(still{}, kavad.NewStage(), painter{}), nil
}

type still struct{}

func (still) Send(context.Context, string) error { return nil }
func (still) Done() bool                         { return false }

type point3 struct{ x, y, z float64 }
type painter struct{}

func color(hex string, alpha float64) kavad.Color {
	c := kavad.MustHex(hex)
	c.A = alpha
	return c
}

// project turns a model point into a screen point. The camera returns to its
// starting angle at the end of the six-second loop.
func project(p point3, phase float64) kavad.Point {
	azimuth := 0.66 + 0.22*math.Sin(phase)
	elevation := 0.65
	rx := math.Cos(azimuth)*p.x - math.Sin(azimuth)*p.y
	ry := math.Sin(azimuth)*p.x + math.Cos(azimuth)*p.y
	return kavad.Point{
		X: 408 + rx*2.05,
		Y: 365 + (ry*math.Sin(elevation)-p.z*math.Cos(elevation))*2.05,
	}
}

func stroke(c kavad.Canvas, phase float64, pts []point3, col kavad.Color, width float64, closed bool) {
	out := make([]kavad.Point, len(pts))
	for i, p := range pts {
		out[i] = project(p, phase)
	}
	c.StrokePolyline(out, width, closed, col)
}

func circle(cx, cy, z, radius float64, segments int) []point3 {
	pts := make([]point3, segments)
	for i := range pts {
		t := 2 * math.Pi * float64(i) / float64(segments)
		pts[i] = point3{cx + radius*math.Cos(t), cy + radius*math.Sin(t), z}
	}
	return pts
}

func plate(z float64) []point3 {
	return []point3{{-130, -80, z}, {130, -80, z}, {130, 80, z}, {-130, 80, z}}
}

func (painter) Paint(c kavad.Canvas, now int) {
	phase := 2 * math.Pi * float64(now) / periodMS
	c.FillPolygon([]kavad.Point{{X: 0, Y: 0}, {X: width, Y: 0},
		{X: width, Y: height}, {X: 0, Y: height}}, kavad.MustHex("#0b1020"))

	base := color("#5ad1e6", 0.68)
	ghost := color("#5ad1e6", 0.2)
	highlight := color("#8ce4f4", 0.95)
	bottom := plate(-22)
	top := plate(0)
	stroke(c, phase, bottom, ghost, 1.1, true)
	for i, corner := range top {
		stroke(c, phase, []point3{corner, bottom[i]}, base, 1.4, false)
	}
	topFace := make([]kavad.Point, len(top))
	for i, corner := range top {
		topFace[i] = project(corner, phase)
	}
	c.FillPolygon(topFace, color("#5ad1e6", 0.055))
	stroke(c, phase, top, highlight, 2, true)
	stroke(c, phase, circle(0, 0, 0, 37, 48), base, 1.7, true)
	stroke(c, phase, circle(0, 0, -22, 37, 48), ghost, 1.1, true)
	for _, x := range []float64{-96, 96} {
		for _, y := range []float64{-47, 47} {
			stroke(c, phase, circle(x, y, 0, 8, 24), base, 1.4, true)
		}
	}

	// The annular part lifts while the two pins move in opposite directions.
	ringBottom := 38 + 23*(1-math.Cos(phase))/2
	ringTop := ringBottom + 22
	outer := color("#9b8cf0", 0.86)
	inner := color("#c6bbff", 0.65)
	stroke(c, phase, circle(0, 0, ringBottom, 57, 64), ghost, 1.2, true)
	stroke(c, phase, circle(0, 0, ringTop, 57, 64), outer, 2.2, true)
	stroke(c, phase, circle(0, 0, ringTop, 29, 64), inner, 1.8, true)
	stroke(c, phase, circle(0, 0, ringBottom, 29, 64), ghost, 1.1, true)
	for i := range 12 {
		a := 2 * math.Pi * float64(i) / 12
		stroke(c, phase, []point3{
			{57 * math.Cos(a), 57 * math.Sin(a), ringBottom},
			{57 * math.Cos(a), 57 * math.Sin(a), ringTop},
		}, outer, 1.1, false)
	}
	for j, x := range []float64{-96, 96} {
		z := 6 + 30*(1-math.Cos(phase+float64(j)*math.Pi))/2
		gold := color("#f0b429", 0.86)
		stroke(c, phase, circle(x, -47, z, 6, 24), gold, 1.8, true)
		stroke(c, phase, circle(x, -47, z+45, 6, 24), gold, 1.8, true)
		for _, side := range []float64{-6, 6} {
			stroke(c, phase, []point3{
				{x + side, -47, z},
				{x + side, -47, z + 45},
			}, gold, 1.4, false)
		}
	}
}

func main() { svg.Main(show{}) }
