# ADR 0002: Terminal graphics — braille for line art, half-blocks for sprites

**Status**: accepted (2026-07)

## Context
We want expressive visuals (the dino, the mini-game) that survive every
terminal/font, including phone terminals that render braille and other
East-Asian-ambiguous glyphs double-wide.

## Decision
- **Braille (2x4 mono dots)** for the compact header animation: densest
  portable mode for small line art. Sprite rows contain ONLY braille chars
  (blanks are U+2800) so width quirks shift whole rows uniformly.
- **Half-block ▀ PixelCanvas (1x2 full-RGB pixels per cell)** for sprite
  scenes (the game): color carries more than dot density; ▀ is universal
  and width-safe. This is the same portable fallback chafa/notcurses use.
- **No octants** (font support too new) and **no sixel/kitty for now**
  (terminal-dependent; image placements fight cell-diff renderers) — see
  issue #94 for the capability-detected future tier.
- Never place ambiguous-width glyphs (✓ · ❯ arrows) inside bordered boxes;
  huh forms use a left-accent-bar style with no right edge for this reason.
