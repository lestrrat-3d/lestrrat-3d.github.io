// Command promo renders the site's CAD hero as a short Kavad animation.
// Run from the repository root with go -C tools/promo run . -out ../../.tmp/promo.
package main

import (
	"context"
	"math"
	"os"

	"github.com/lestrrat-go/kavad"
	"github.com/lestrrat-go/kavad/svg"
)

const (
	width  = 800
	height = 600
)

type show struct{}

func (show) Size() (int, int) { return width, height }

func (show) Start(context.Context) (*kavad.Run, error) {
	image, err := kavad.LoadPNG(os.DirFS("../.."), "hero.png")
	if err != nil {
		return nil, err
	}
	return kavad.NewRun(still{}, kavad.NewStage(), painter{image: image}, image), nil
}

type still struct{}

func (still) Send(context.Context, string) error { return nil }
func (still) Done() bool                         { return false }

type painter struct{ image *kavad.Image }

func (p painter) Paint(c kavad.Canvas, now int) {
	phase := 2 * math.Pi * float64(now) / 6000
	background := kavad.MustHex("#0b1020")
	c.FillPolygon([]kavad.Point{{X: 0, Y: 0}, {X: width, Y: 0},
		{X: width, Y: height}, {X: 0, Y: height}}, background)

	// The CAD render stays sharp while its position and scale move gently.
	scale := 1.02 + 0.025*math.Cos(phase)
	w, h := 760*scale, 555*scale
	x := (width-w)/2 + 11*math.Sin(phase)
	y := (height-h)/2 + 7*math.Sin(phase+0.7)
	c.DrawImage(p.image, x, y, w, h, 0.9)

	accent := kavad.MustHex("#5ad1e6")
	accent.A = 0.85
	c.StrokePolyline([]kavad.Point{{X: 465, Y: 542}, {X: 706, Y: 542}}, 1.5, false, accent)
	c.FillCircle(585+121*math.Sin(phase), 542, 4, accent)
	c.DrawText(kavad.Text{X: 706, Y: 530, S: "CODE  /  CAD  /  PIXELS",
		Font: kavad.Mono, Size: 15, Align: kavad.AlignEnd, Color: accent})
}

func main() { svg.Main(show{}) }
