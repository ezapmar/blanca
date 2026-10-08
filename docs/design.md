<p align="center">
  <img src="../assets/blanca-banner.png" alt="Blanca" width="720">
</p>

# Design

Blanca has a terminal bezel, a launcher icon, a toolbar glyph and a README. This page is
the point of view for all of them, and for anything that comes later.

## The idea

A dog's face drawn in one confident black line. Head tilted, one ear up, the look a
terrier gives you when you have said its name and nothing else. The mark is Blanca
herself, an Anatolian sighthound and Russell terrier mix. Everything else is white paper
and black ink around that drawing.

What follows from that:

- **Ink on paper.** The mark is black line art with a white face. It is never recoloured,
  never outlined in another colour, never given a gradient or a shadow.
- **One face per surface.** A screen or an image carries the mark once, large, with
  room around it. The bezel shows one clipping, the icon shows one dog.
- **Quiet ground.** The home ground is graphite, a dark warm grey, so the white face is
  the brightest thing on it. On light grounds the mark stands on its own.
- **Capitalised.** The name is `Blanca`, capital B, in text and in the wordmark. It is a
  name, not a command.
- **One accent, rarely.** Vermilion, borrowed from the brainless family palette, for
  primary actions. The wordmark carries no dot.
- **Kilim at the edges.** On the web page, the brainless family's Anatolian motifs (the
  eye, the star, the ram's horn, the tree of life) sit beside the content as 15 by 15
  knot grids in bone, and a woven band of small eyes in vermilion, saffron and teal
  separates sections. Buttons and code blocks take stepped corners. The motifs never
  touch the mark and never appear in the bezel.

## Palette

| Token | Hex | Role |
|---|---|---|
| ink | `#111111` | The line. Text on paper |
| paper | `#FFFFFF` | The face. Light surfaces |
| graphite | `#262626` | Ground. Tile, banner, dark surfaces |
| bone | `#F9EBDB` | Wordmark and body text on graphite |
| ash | `#9B9B9B` | Secondary text on graphite |
| vermilion | `#F24B1E` | The one accent: primary action. A dye in the woven band |
| teal | `#0C9794` | Links, success |
| saffron | `#FBA335` | Warnings |

On graphite use bone for text and ash for secondary. On paper use ink. Never put bone or
ash text on paper.

## Logo

| File | Use |
|---|---|
| `assets/blanca.svg` | The mark, white face, transparent ground. Default wherever a logo fits |
| `assets/blanca-tile.svg` | The mark on a graphite rounded square. App launcher, avatars |
| `assets/blanca-symbolic.svg` | Solid silhouette in `currentColor`, eyes and nose knocked out. Bars, trays, anything 16 to 32 px |
| `assets/png/` | Raster sizes of all three |
| `assets/blanca-banner.png` | Mark, wordmark and tagline on graphite. README header |
| `assets/blanca-social.png` | 1280 by 640 GitHub social preview |
| `assets/source/blanca-original.svg` | The supplied artwork. Not used directly |

Rules:

- The mark is the supplied artwork (08/10/2026) reframed, never redrawn.
- Clear space around the mark of at least the height of one ear.
- Minimum size: 48 px for the mark and tile, 16 px for the symbolic glyph. Below 48 px
  the line art closes up, so switch to the symbolic.
- On dark bars the symbolic glyph takes the bar's own foreground colour. Do not force white.
- Wordmark: a heavy grotesque, capital B, bone on graphite, no full stop.
  The banner is the only typeset wordmark; do not set it in running text.

## The bezel

The bezel is a terminal, so it inherits whatever Omarchy theme is running and adds
nothing of its own. One clipping, its position, a line of key hints. Bold for the title,
dim for the hints, the terminal's foreground for the text. No colours, no boxes, no
emoji. It must read the same on a light theme, a dark theme, and over `ssh`.

## Where the logo is set

- **GitHub social preview.** Repository Settings, Social preview, upload
  `assets/blanca-social.png`. There is no API for it.
- **Omarchy launcher.** `blanca.desktop` points at `blanca-tile`; the installer and the
  AUR package put the SVGs in the hicolor theme.
