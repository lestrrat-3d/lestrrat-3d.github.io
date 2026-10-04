# Header promo

This command uses the site's `hero.png` and the pinned Kavad module to render
the animated header art. The site serves the committed WebM or MP4 file.

Run these commands from the repository root. The capture step needs Node,
Playwright with Chromium, and ffmpeg.

```sh
go -C tools/promo run . -out ../../.tmp/promo -fps 24 -seconds 6
npm install --prefix .tmp/promo-node --no-save playwright@1.56.1
.tmp/promo-node/node_modules/.bin/playwright install chromium
kavad_module_dir=$(go -C tools/promo list -m -f '{{.Dir}}' github.com/lestrrat-go/kavad)
NODE_PATH="$PWD/.tmp/promo-node/node_modules" node "$kavad_module_dir/capture/capture.cjs" .tmp/promo
ffmpeg -y -i .tmp/promo/video.mp4 -c copy -movflags +faststart hero-promo.mp4
ffmpeg -y -i hero-promo.mp4 -c:v libvpx-vp9 -b:v 0 -crf 34 -an hero-promo.webm
```

The video uses `hero.png` as its poster. Visitors who request reduced motion
see the static `hero.png` background instead.
