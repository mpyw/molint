# Assets

| File | Size | Used by |
| --- | --- | --- |
| `icon.png` | 512×512, transparent | The top of the README |
| `social-preview.png` | 1280×640 | The repository's social preview, set in its settings |

Both were drawn with ChatGPT, then scaled and reduced to 256 colors. To draw them again, give it these prompts in one conversation, in this order.

## Logo

```text
Create an original app icon for a developer tool called "molint", a Go linter
that enforces the samber/mo library, whose mascot is a unicorn.

Subject: a friendly, confident unicorn head, facing three-quarters to the left,
wearing sleek black sunglasses with a subtle reflection. A short rainbow mane
flows back from the forehead. The horn is golden and simple.

Style: flat vector illustration, bold clean shapes, thick smooth outlines,
no gradients except a soft highlight on the sunglasses. It must stay readable
at 32×32 pixels, so keep details few and large.

Colors: white or very light lavender unicorn, a pastel rainbow mane
(pink, orange, yellow, mint, sky blue, violet), black sunglasses.

Composition: centered, the head fills about 80% of the canvas, square 1024×1024,
transparent background. No text, no letters, no border, no drop shadow.

Make it an original design. Do not copy any existing emoji artwork.
```

## Social preview

```text
Now create a social preview banner for the same project, using this exact
unicorn. Keep the unicorn identical: same shape, colors, outline, sunglasses
and expression. Do not redraw or restyle it.

Size: 1280×640 pixels.

Layout:
- Right side: the unicorn, about 480 pixels tall, placed so that it faces
  left, toward the text.
- Left side: the word "molint" in a large, bold, rounded sans-serif, in white.
  Below it, smaller, in light lavender: "Absence is mo.Option, not nil".
  Left-align both lines.

Background: deep navy (#1B1F3B), with a very subtle pattern of faint curly
braces and angle brackets, like source code, at about 5% opacity. Add a soft
glow behind the unicorn so its dark outline stands out from the background.

Keep the text and the unicorn inside the central 1100×540 area, since the
edges may be cropped. Spell the text exactly as given. No other text,
no logos, no border.
```

## Scaling

The images came out larger than their sizes above. They were scaled with `sips`, then reduced with Pillow:

```bash
sips -z 512 512 icon-original.png --out icon.png
sips -z 640 1280 preview-original.png --out social-preview.png
python3 -c '
from PIL import Image
Image.open("icon.png").quantize(256, method=Image.Quantize.FASTOCTREE, dither=Image.Dither.NONE).save("icon.png", optimize=True)
Image.open("social-preview.png").convert("RGB").quantize(256, method=Image.Quantize.MEDIANCUT, dither=Image.Dither.NONE).save("social-preview.png", optimize=True)
'
```
