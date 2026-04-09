# CSS Discrepancy Analysis: Current UI vs Original JSX Prototype

## Methodology

Compared the original `glazed-docs-browser(2).jsx` inline styles against the
current modular React components' CSS files, using Playwright-computed styles
and visual side-by-side inspection.

## Critical Discrepancies

### 1. Title bar stripe pattern

**Original:** `backgroundImage: STRIPE` where STRIPE is a 1×1 black pixel in a
1×2 SVG, with `backgroundSize: "1px 2px"`. This creates a **horizontal
hairline** (every other row).

**Current:** `repeating-linear-gradient(to bottom, transparent 1px, #000 1px, #000 2px)`.
Visually similar but uses CSS gradient instead of SVG data URI. Works, but
doesn't match the original's `backgroundSize: "1px 2px"` / `backgroundRepeat: "repeat"`.

**Fix:** Use the same SVG data URI pattern as the original.

### 2. Title bar ruler height

**Original:** The ruler div has `height: 6` inline, and stretches to fill the
22px title bar. The stripes are visible as 6px-high horizontal rules flanking
the title.

**Current:** `[data-part='titlebar-ruler']` sets `height: 6px` but uses
`overflow: hidden` which clips. The stripes work but the ruler is only 6px
instead of stretching to fill the bar.

**Fix:** Remove `height: 6px` and `overflow: hidden` from the ruler. Let the
ruler stretch to fill the title bar (flex:1 + align-items:center already
handle vertical centering). The stripes should be fixed-height children.

### 3. Badge border width

**Original:** `border: "1.5px solid ${color}"` — 1.5px.

**Current:** `border: 1.5px solid var(--badge-color, #000)` — correct!

**Verdict:** Match.

### 4. Filter button inactive border

**Original:** `border: "1px solid #999"` for inactive buttons.

**Current:** `border: 1px solid #999` — correct.

**Verdict:** Match.

### 5. Filter button active border

**Original:** `border: "2px solid #000"` — 2px border when active.

**Current:** `border: 2px solid #000` — correct.

**Verdict:** Match.

### 6. Background dither pattern

**Original:** `background: "#a8a8a8", backgroundImage: DITHER, backgroundSize: "4px 4px"`
where DITHER is a 4×4 SVG with two 2×2 black squares at (0,0) and (2,2) — **solid
black** (no opacity).

**Current:** Uses `opacity='0.08'` on the SVG rects. This produces a much
lighter, barely-visible dither compared to the original's strong checkerboard.

**Fix:** Remove opacity, match the original's solid-black dither pattern.

### 7. Empty state layout

**Original:** Icon and text are stacked vertically:
```jsx
<div style={{ textAlign: "center" }}>
  <div style={{ fontSize: 40, marginBottom: 10, filter: "grayscale(1)" }}>📖</div>
  Select a section from the list.
</div>
```

**Current:** Uses `display: flex` which puts icon and text side-by-side
instead of stacked vertically. The icon has `display: block` but the flex
container is `flex-direction: row` by default.

**Fix:** Add `flex-direction: column` to the empty state container, or switch
to the original's simpler block layout.

### 8. Card even/odd striping

**Original:** `background: idx % 2 === 0 ? "#fff" : "#f5f5f5"` — even cards are
white, odd cards are #f5f5f5.

**Current:** `[data-part='section-list-item']:nth-child(even) { background: #f5f5f5; }`
— even cards are #f5f5f5. **Inverted** from the original!

**Fix:** Change to `:nth-child(odd)` for #f5f5f5, or swap the default/alternate
colors.

### 9. Active card text color for description

**Original:** Active card short description uses `color: on ? "#aaa" : "#777"`.

**Current:** Only has `color: #aaa` for the active state. The default is
`color: #777`. These match.

**Verdict:** Match.

### 10. Active card badge rendering

**Original:** When active, replaces the Badge component with a plain span:
```jsx
{on && <span style={{ fontSize: 10, fontWeight: 700, color: "#aaa" }}>
  {TYPE_BADGE[s.type]?.label}
</span>}
```

**Current:** SectionCard conditionally renders either Badge or a plain span.
Matches the original behavior.

**Verdict:** Match.

## Minor Discrepancies

### 11. Badge `letterSpacing`

**Original:** `letterSpacing: 0.3` (0.3px).

**Current:** `letter-spacing: 0.3px` in badge.css — correct.

**Verdict:** Match.

### 12. Search input placeholder style

**Current CSS has no placeholder style.** Should add:
```css
[data-part='searchbar-input']::placeholder {
  color: #999;
}
```

This already exists in searchbar.css — **match**.

### 13. Global font-face for Chicago

**Original JSX includes:**
```css
@font-face { font-family: 'Chicago_'; src: url('https://cdn.jsdelivr.net/gh/polgfred/mac-fonts@main/Chicago.woff2') format('woff2'); }
```

**Current:** No `@font-face` declaration. The font stack references `Chicago_`
but it never loads because the font isn't declared. Users see fallback Geneva/Charcoal.

**Fix:** Add the `@font-face` declaration to global.css.

## Summary Table

| # | Element | Issue | Severity |
|---|---------|-------|----------|
| 1 | Title bar stripes | SVG pattern vs CSS gradient | Low |
| 2 | Title bar ruler height | 6px clipped vs full-height | Medium |
| 6 | Background dither | opacity 0.08 vs solid black | High |
| 7 | Empty state | flex-row vs vertical stack | Medium |
| 8 | Card striping | even/odd inverted | Medium |
| 13 | Chicago font | Missing @font-face | High |
