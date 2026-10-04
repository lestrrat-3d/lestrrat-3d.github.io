# Header promo

`morph.html` renders twelve copies of one 3D shape sequence across three rows.
Each cube changes size, gains chamfered and rounded edges, becomes a sphere,
develops a hole, and becomes a torus. The copies start at different times in
the sequence, so several shapes appear at once. Neighboring shapes use distinct
colors and turn in opposite directions. The upper and lower rows move right;
the middle row moves left. Each row advances by its full four-shape pattern over
the 24-second loop. The shape sequence repeats twice and each turn completes
one revolution, so the last frame joins the first.

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
NODE_PATH="$PWD/.tmp/promo-node/node_modules" node tools/promo/capture.cjs .tmp/promo-frames 24 24
ffmpeg -y -framerate 24 -i .tmp/promo-frames/frame_%04d.png -c:v libx264 -crf 23 -pix_fmt yuv420p -movflags +faststart hero-promo.mp4
ffmpeg -y -i hero-promo.mp4 -c:v libvpx-vp9 -b:v 0 -crf 32 -an hero-promo.webm
ffmpeg -y -i .tmp/promo-frames/frame_0000.png -frames:v 1 hero-poster.png
```

The page serves WebM with MP4 fallback. The poster stays visible when a visitor
requests reduced motion. A small button pauses playback when JavaScript runs;
the browser's video controls remain available when it does not.
