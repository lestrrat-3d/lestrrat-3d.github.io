// Command hero renders the artwork behind the site's hero section.
//
// The parts are modeled the way the rest of this organization's stack is meant
// to be used: sketch solves the 2D profiles, decad extrudes and places them and
// proves the assembly sound, and solidlens rasterizes it. Nothing here is drawn
// by hand.
//
//	go run ./tools/hero -out hero.png
package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"

	"github.com/lestrrat-3d/decad"
	"github.com/lestrrat-3d/r3"
	"github.com/lestrrat-3d/sketch"
	"github.com/lestrrat-3d/solidlens"
	"github.com/lestrrat-3d/units"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "hero: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	out := flag.String("out", "hero.png", "PNG file to write")
	width := flag.Int("width", 1300, "output width in pixels")
	height := flag.Int("height", 950, "output height in pixels")
	flag.Parse()

	ctx := context.Background()

	doc := decad.New()
	parts, err := build(ctx, doc)
	if err != nil {
		return err
	}

	// Every piece is a separate body, so Verify's pairwise partition is the
	// check that the arrangement holds: pieces rest against each other with a
	// clearance, and nothing interpenetrates.
	report, err := doc.Verify(ctx, decad.WithTolerance(units.Scalar(2e-2)))
	if err != nil {
		return fmt.Errorf("failed to verify: %w", err)
	}
	name := make(map[*decad.Body]string, len(parts))
	for _, p := range parts {
		name[p.body] = p.name
	}
	fmt.Fprintf(os.Stderr, "pieces: %d, document bodies: %d, verify: %s (trustworthy: %v)\n",
		len(parts), len(report.Bodies), report.Status, report.Trustworthy())
	for _, p := range parts {
		box, err := p.body.Bounds()
		if err != nil {
			return fmt.Errorf("failed to read bounds of %s: %w", p.name, err)
		}
		fmt.Fprintf(os.Stderr, "  %-7s x %7.1f..%7.1f  y %7.1f..%7.1f  z %7.1f..%7.1f\n",
			p.name, box.Min.X, box.Max.X, box.Min.Y, box.Max.Y, box.Min.Z, box.Max.Z)
	}
	for _, d := range report.Diagnostics {
		who := ""
		if d.Pair != nil {
			who = fmt.Sprintf(" [%s + %s]", name[d.Pair.A], name[d.Pair.B])
		} else if d.Body != nil {
			who = fmt.Sprintf(" [%s]", name[d.Body])
		}
		fmt.Fprintf(os.Stderr, "  %s%s: %s\n", d.Status, who, d.Message)
	}

	return render(ctx, parts, *out, *width, *height)
}

// piece is one body, the name it is reported under, and the color it renders in.
type piece struct {
	name  string
	body  *decad.Body
	color solidlens.Color
}

// The site palette, so the render and the page agree.
var (
	cyan   = solidlens.RGB(0.19, 0.60, 0.75)
	violet = solidlens.RGB(0.42, 0.35, 0.78)
	coral  = solidlens.RGB(0.78, 0.34, 0.26)
	amber  = solidlens.RGB(0.80, 0.55, 0.12)
	green  = solidlens.RGB(0.22, 0.64, 0.46)
	steel  = solidlens.RGB(0.30, 0.38, 0.52)
)

// plateTop is the height of the pad the standing pieces rest on: the plate is
// seated on the ground and is this thick.
const (
	plateTop = 7
	// Pieces are set down just clear of what they rest on. A piece touching
	// another exactly is a contact the interference proof cannot classify, and
	// the gap is far too small to see.
	clearance = 0.4
	// cubeHeight is how tall the cube is, and cubeTop where its lid ends up
	// once it is seated on the plate.
	cubeHeight = 38
	cubeTop    = plateTop + clearance + cubeHeight
)

func build(ctx context.Context, doc *decad.Document) ([]piece, error) {
	w := sketch.NewWorld()

	// The pad everything else rests on.
	plate, err := extrude(ctx, w, doc, 7, func(s *sketch.Sketch) error {
		s.CreateRectangle(-75, -75, 75, 75)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build plate: %w", err)
	}
	plate, err = seat(ctx, plate, 0, 8, 0, 0, 0)
	if err != nil {
		return nil, err
	}

	// A cube, turned off the plate's axes.
	cube, err := extrude(ctx, w, doc, cubeHeight, func(s *sketch.Sketch) error {
		s.CreateRectangle(-19, -19, 19, 19)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build cube: %w", err)
	}
	cube, err = seat(ctx, cube, 0, 26, 34, 36, plateTop+clearance)
	if err != nil {
		return nil, err
	}

	// A hex nut standing on the plate.
	nut, err := extrude(ctx, w, doc, 15, func(s *sketch.Sketch) error {
		if _, err := s.CreatePolygon(0, 0, 6, 20); err != nil {
			return err
		}
		s.CreateCircle(s.CreatePoint(0, 0), 10)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build nut: %w", err)
	}
	nut, err = seat(ctx, nut, 0, 12, -54, 26, plateTop+clearance)
	if err != nil {
		return nil, err
	}

	// A ring stood on edge and leaned back against the cube.
	ring, err := extrude(ctx, w, doc, 7, func(s *sketch.Sketch) error {
		s.CreateCircle(s.CreatePoint(0, 0), 30)
		s.CreateCircle(s.CreatePoint(0, 0), 18)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build ring: %w", err)
	}
	ring, err = seat(ctx, ring, 66, 96, -16, 30, plateTop+clearance)
	if err != nil {
		return nil, err
	}

	// A rod lying across the plate, one end up on the cube.
	rod, err := extrude(ctx, w, doc, 84, func(s *sketch.Sketch) error {
		s.CreateCircle(s.CreatePoint(0, 0), 8)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build rod: %w", err)
	}
	rod, err = seat(ctx, rod, 90, 100, -58, -46, plateTop+clearance)
	if err != nil {
		return nil, err
	}

	// A slotted bar off the corner of the plate, flat on the ground.
	bar, err := extrude(ctx, w, doc, 9, func(s *sketch.Sketch) error {
		if _, err := s.CreateSlot(-34, 0, 34, 0, 13); err != nil {
			return err
		}
		s.CreateCircle(s.CreatePoint(-34, 0), 6)
		s.CreateCircle(s.CreatePoint(34, 0), 6)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build bar: %w", err)
	}
	bar, err = seat(ctx, bar, 3, 40, 10, -124, clearance)
	if err != nil {
		return nil, err
	}

	// A short spool standing on the plate.
	spool, err := extrude(ctx, w, doc, 26, func(s *sketch.Sketch) error {
		s.CreateCircle(s.CreatePoint(0, 0), 13)
		s.CreateCircle(s.CreatePoint(0, 0), 6)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build spool: %w", err)
	}
	spool, err = seat(ctx, spool, 0, 0, 46, -24, plateTop+clearance)
	if err != nil {
		return nil, err
	}

	// A washer left on top of the cube, dropped a little off square.
	washer, err := extrude(ctx, w, doc, 6, func(s *sketch.Sketch) error {
		s.CreateCircle(s.CreatePoint(0, 0), 16)
		s.CreateCircle(s.CreatePoint(0, 0), 7)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build washer: %w", err)
	}
	washer, err = seat(ctx, washer, 7, 40, 38, 32, cubeTop+clearance)
	if err != nil {
		return nil, err
	}

	return []piece{
		{"plate", plate, steel},
		{"washer", washer, green},
		{"cube", cube, violet},
		{"nut", nut, amber},
		{"ring", ring, cyan},
		{"rod", rod, coral},
		{"bar", bar, green},
		{"spool", spool, cyan},
	}, nil
}

// extrude solves one sketch and extrudes its outermost region. draw builds the
// profile; the outer boundary is the region carrying every hole drawn in it.
func extrude(ctx context.Context, w *sketch.World, doc *decad.Document, h float64, draw func(*sketch.Sketch) error) (*decad.Body, error) {
	s, err := w.CreateSketch(w.XY())
	if err != nil {
		return nil, fmt.Errorf("failed to create sketch: %w", err)
	}
	if err := draw(s); err != nil {
		return nil, fmt.Errorf("failed to draw profile: %w", err)
	}
	if _, err := s.Solve(ctx); err != nil {
		return nil, fmt.Errorf("failed to solve sketch: %w", err)
	}

	profile := widest(s.Profiles())
	if profile == nil {
		return nil, fmt.Errorf("sketch produced no closed region")
	}
	return doc.Extrude(s, profile, decad.Distance{D: units.Millimeters(h), Dir: decad.Along})
}

// widest picks the region that encloses the others: the one with the most
// holes, and among equals the one with the largest area.
func widest(profiles []*sketch.Profile) *sketch.Profile {
	var best *sketch.Profile
	for _, p := range profiles {
		switch {
		case best == nil:
			best = p
		case len(p.Holes) > len(best.Holes):
			best = p
		case len(p.Holes) == len(best.Holes) && p.Area > best.Area:
			best = p
		}
	}
	return best
}

// seat tips a body up by tilt degrees about +X, spins it by spin degrees about
// +Z, and sets it down at (x, y) so that its lowest point rests on rest. Both
// rotations run through the model origin, which is where every profile above is
// drawn. The resting height is measured on the rotated geometry rather than on
// a bounding box, so a tilted ring touches down on its rim and not on the
// corner of a box it never reaches.
func seat(ctx context.Context, b *decad.Body, tilt, spin, x, y, rest float64) (*decad.Body, error) {
	rot, err := rotation(tilt, spin)
	if err != nil {
		return nil, err
	}

	mesh, err := b.TessellateContext(ctx, units.Millimeters(0.02))
	if err != nil {
		return nil, fmt.Errorf("failed to tessellate for seating: %w", err)
	}
	low := math.Inf(1)
	for _, v := range mesh.Vertices() {
		if z := rot.Apply(v).Z; z < low {
			low = z
		}
	}

	move, err := r3.Translation(r3.Vec{X: x, Y: y, Z: rest - low})
	if err != nil {
		return nil, fmt.Errorf("failed to build translation: %w", err)
	}
	all, err := rot.Then(move)
	if err != nil {
		return nil, fmt.Errorf("failed to compose placement: %w", err)
	}
	placed, err := b.Placed(all)
	if err != nil {
		return nil, fmt.Errorf("failed to place body: %w", err)
	}
	return placed, nil
}

func rotation(tilt, spin float64) (r3.Transform, error) {
	t, err := r3.RotationAround(r3.Vec{}, r3.Vec{X: 1}, units.Degrees(tilt))
	if err != nil {
		return r3.Transform{}, fmt.Errorf("failed to build tilt: %w", err)
	}
	z, err := r3.RotationAround(r3.Vec{}, r3.Vec{Z: 1}, units.Degrees(spin))
	if err != nil {
		return r3.Transform{}, fmt.Errorf("failed to build spin: %w", err)
	}
	all, err := t.Then(z)
	if err != nil {
		return r3.Transform{}, fmt.Errorf("failed to compose rotations: %w", err)
	}
	return all, nil
}

func render(ctx context.Context, parts []piece, out string, width, height int) error {
	models := make([]solidlens.Model, 0, len(parts))
	for _, p := range parts {
		mesh, err := p.body.TessellateContext(ctx, units.Millimeters(0.1))
		if err != nil {
			return fmt.Errorf("failed to tessellate: %w", err)
		}
		models = append(models, solidlens.Model{
			Mesh:     mesh,
			Material: solidlens.Material{Color: p.color, Ambient: 0.24},
			Edges: solidlens.Edges{
				Enabled: true,
				Color:   solidlens.RGB(0.85, 0.94, 1.0),
				Width:   1.6,
				// Well above the angle two chords of a bore meet at, so a
				// drilled wall keeps its outline instead of showing facets.
				CreaseAngle: 40,
			},
		})
	}

	scene := solidlens.Scene{
		Camera: solidlens.Camera{
			Position: solidlens.Vec{X: 205, Y: -300, Z: 330},
			Target:   solidlens.Vec{X: -15, Y: -10, Z: -11},
			Up:       solidlens.Vec{Z: 1},
			FOV:      22,
		},
		Models: models,
		// Directional light only: a flat face then shades uniformly, which is
		// what makes the render read as a drawing rather than a photograph.
		DirectionalLights: []solidlens.DirectionalLight{
			{
				Direction: solidlens.Vec{X: -0.28, Y: 0.42, Z: -0.86},
				Color:     solidlens.RGB(1, 0.99, 0.96),
				Intensity: 1.05,
			},
			{
				Direction: solidlens.Vec{X: 0.72, Y: 0.34, Z: -0.12},
				Color:     solidlens.RGB(0.56, 0.5, 0.98),
				Intensity: 0.5,
			},
		},
		Background: solidlens.RGBA(0, 0, 0, 0),
	}

	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", out, err)
	}
	defer f.Close()

	if err := solidlens.RenderPNG(ctx, f, scene, solidlens.Settings{Width: width, Height: height}); err != nil {
		return fmt.Errorf("failed to render: %w", err)
	}
	return nil
}
