// Command workflow generates the sketch and verified model shown in the site's
// agent workflow infographic. Run it with:
//
//	go -C tools/workflow run . -out ../../workflow
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/lestrrat-3d/decad"
	"github.com/lestrrat-3d/sketch"
	"github.com/lestrrat-3d/solidlens"
	"github.com/lestrrat-3d/units"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "workflow: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	out := flag.String("out", "../../workflow", "output file prefix")
	flag.Parse()
	ctx := context.Background()

	w := sketch.NewWorld()
	s, err := w.CreateSketch(w.XY())
	if err != nil {
		return fmt.Errorf("create sketch: %w", err)
	}
	s.CreateRectangle(-48, -30, 48, 30)
	s.CreateCircle(s.CreatePoint(-26, 0), 8)
	s.CreateCircle(s.CreatePoint(26, 0), 8)
	if _, err := s.Solve(ctx); err != nil {
		return fmt.Errorf("solve sketch: %w", err)
	}

	var profile *sketch.Profile
	for _, p := range s.Profiles() {
		if len(p.Holes) == 2 {
			profile = p
			break
		}
	}
	if profile == nil || !profile.Valid {
		return fmt.Errorf("sketch did not produce a valid two-hole plate")
	}

	sketchImage, err := s.PNG(
		sketch.WithGrid(true),
		sketch.WithShowPoints(false),
		sketch.WithProfileFill(true),
		sketch.WithScale(5),
	)
	if err != nil {
		return fmt.Errorf("render sketch: %w", err)
	}

	doc := decad.New()
	body, err := doc.Extrude(s, profile, decad.Distance{D: units.Millimeters(9), Dir: decad.Along})
	if err != nil {
		return fmt.Errorf("extrude plate: %w", err)
	}
	report, err := doc.Verify(ctx)
	if err != nil {
		return fmt.Errorf("verify plate: %w", err)
	}
	if report.Status != decad.Sound {
		return fmt.Errorf("plate verification: %s", report.Status)
	}
	fmt.Fprintln(os.Stderr, "plate verification: sound")

	mesh, err := body.TessellateContext(ctx, units.Millimeters(0.15))
	if err != nil {
		return fmt.Errorf("tessellate plate: %w", err)
	}
	scene := solidlens.Scene{
		Camera: solidlens.Camera{
			Position: solidlens.Vec{X: 115, Y: -155, Z: 165},
			Target:   solidlens.Vec{Z: 4},
			Up:       solidlens.Vec{Z: 1},
			FOV:      25,
		},
		Models: []solidlens.Model{{
			Mesh:     mesh,
			Material: solidlens.Material{Color: solidlens.RGB(0.19, 0.60, 0.75), Ambient: 0.25},
			Edges: solidlens.Edges{
				Enabled: true, Color: solidlens.RGB(0.85, 0.94, 1), Width: 1.5,
				CreaseAngle: 40,
			},
		}},
		DirectionalLights: []solidlens.DirectionalLight{{
			Direction: solidlens.Vec{X: -0.4, Y: 0.5, Z: -0.7},
			Color:     solidlens.RGB(1, 0.99, 0.96),
			Intensity: 1.1,
		}},
		Background: solidlens.RGBA(0, 0, 0, 0),
	}
	var modelImage bytes.Buffer
	if err := solidlens.RenderPNG(ctx, &modelImage, scene, solidlens.Settings{Width: 700, Height: 460}); err != nil {
		return fmt.Errorf("render model: %w", err)
	}
	if err := os.WriteFile(*out+"-sketch.png", sketchImage, 0o644); err != nil {
		return fmt.Errorf("write sketch: %w", err)
	}
	if err := os.WriteFile(*out+"-model.png", modelImage.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write model: %w", err)
	}
	return nil
}
