// Render the WebGL morph study as numbered PNG frames with local Chromium.
// Usage: node capture.cjs OUTPUT_DIR [FPS=24] [SECONDS=12]
const fs = require('fs');
const path = require('path');
const { pathToFileURL } = require('url');
const { chromium } = require('playwright');

const outputDir = process.argv[2];
const fps = Number(process.argv[3] || 24);
const seconds = Number(process.argv[4] || 12);
if (!outputDir || !Number.isInteger(fps) || !Number.isInteger(seconds) || fps < 1 || seconds < 1) {
  throw new Error('usage: node capture.cjs OUTPUT_DIR [FPS=24] [SECONDS=12]');
}

(async () => {
  fs.mkdirSync(outputDir, { recursive: true });
  const browser = await chromium.launch({
    headless: true,
    args: ['--use-gl=angle', '--use-angle=swiftshader', '--enable-unsafe-swiftshader'],
  });
  try {
    const page = await browser.newPage({ viewport: { width: 1080, height: 420 } });
    page.on('pageerror', error => console.error(error));
    await page.goto(pathToFileURL(path.join(__dirname, 'morph.html')).href);
    for (let frame = 0; frame < fps * seconds; frame++) {
      await page.evaluate(time => window.renderAt(time), frame / fps);
      const name = `frame_${String(frame).padStart(4, '0')}.png`;
      await page.screenshot({ path: path.join(outputDir, name) });
    }
    console.log(`Rendered ${fps * seconds} frames to ${outputDir}`);
  } finally {
    await browser.close();
  }
})().catch(error => { console.error(error); process.exitCode = 1; });
