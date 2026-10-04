# Header promo

`morph.html` renders three 3D shape studies side by side. A block changes size,
gains chamfered and rounded edges, becomes a sphere, develops a hole, and becomes
a torus. A cylinder becomes a tapered, rounded vessel with a bore. A hexagonal
prism becomes a gear and then a ring. All three turn 180 degrees over the
12-second loop. Their starting shapes have half-turn symmetry, so the loop
returns to its first frame without a visible orientation jump.

The models are signed-distance shapes rendered with WebGL2. They illustrate
the geometry changes named in the caption; they are not output from Decad.
Kinetograph can rebuild Decad bodies at frame times, and Kavad can draw SVG
frames and place raster assets. Neither interpolates continuously through the
topology changes between a solid sphere and a torus. The library cards on the
page contain actual output from the tools.

Run these commands from the repository root. The capture step needs Node,
Playwright with Chromium, and ffmpeg. It writes intermediate frames to `.tmp/`.

```sh
npm install --prefix .tmp/promo-node --no-save playwright@1.56.1
.tmp/promo-node/node_modules/.bin/playwright install chromium
NODE_PATH="$PWD/.tmp/promo-node/node_modules" node tools/promo/capture.cjs .tmp/promo-frames 24 12
ffmpeg -y -framerate 24 -i .tmp/promo-frames/frame_%04d.png -c:v libx264 -crf 23 -pix_fmt yuv420p -movflags +faststart hero-promo.mp4
ffmpeg -y -i hero-promo.mp4 -c:v libvpx-vp9 -b:v 0 -crf 32 -an hero-promo.webm
ffmpeg -y -i .tmp/promo-frames/frame_0180.png -frames:v 1 hero-poster.png
```

The page serves WebM with MP4 fallback. The poster stays visible when a visitor
requests reduced motion. A small button pauses playback when JavaScript runs;
the browser's video controls remain available when it does not.
