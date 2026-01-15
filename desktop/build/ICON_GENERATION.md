# Icon Generation

Commands used to generate app icons from the SVG source.

## Prerequisites

- Inkscape (`sudo apt install inkscape`)
- ImageMagick (`sudo apt install imagemagick`)

## Generate Icons

```bash
cd desktop/build

# Generate PNG (1024x1024) from SVG
inkscape aurora-icon.svg -w 1024 -h 1024 -o appicon.png

# Generate Windows ICO with multiple sizes
convert appicon.png -define icon:auto-resize=256,128,64,48,32,16 windows/icon.ico
```

## Icon Files

| File | Purpose |
|------|---------|
| `aurora-icon.svg` | Source vector icon |
| `appicon.png` | Linux & macOS icon (1024x1024) |
| `windows/icon.ico` | Windows executable icon (multi-size) |
