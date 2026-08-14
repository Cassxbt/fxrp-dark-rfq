# Design system

Visual direction extracted from [ciphermuseum.com](https://ciphermuseum.com) and
retargeted for an OTC desk. This documents what was borrowed, what was changed,
and — more importantly — where the borrowed language deliberately stops.

## What that site does well

| Device | How it's built there |
|---|---|
| Lit darkness | `#060608` base with two low-opacity gold radial ellipses (5% top-left, 3% bottom-right). The black reads as a lit gallery rather than an empty div. |
| Single metallic accent | One gold (`#D4B765`) carries every interactive and emphatic role. Semantics (`#CE6D6D` red, `#5AC8A0` green) are desaturated so they never fight it. |
| Tri-font system | Serif display (Cinzel) + serif body (Cormorant Garamond, often italic) + mono (JetBrains Mono) for labels and data. The serif body is what makes it feel like a placard instead of a dashboard. |
| Flanked eyebrow | Mono, uppercase, `.3em` tracking, with 40px hairlines via `::before`/`::after`. Instantly legible as a section marker. |
| Gradient hairline | `linear-gradient(90deg,transparent,gold 30%,gold 70%,transparent)` at 40% opacity — a rule that fades at both ends rather than butting into nothing. |
| Warm parchment text | `#F3EEDF`, never pure white. Paired with warm dark, the whole surface feels aged rather than cold. |
| Card edge accent | A colored top border per card, gold icon, serif title, gold `action →`. Hover lifts 3px and washes a diagonal gold gradient in via `::after`. |
| Restrained radii | 3 / 4 / 10 / 16px. Nothing pill-shaped, nothing sharp. |

## What we take, and what we change

Borrowed: the lit-dark surface, the single-metallic discipline, the tri-font
split, the flanked eyebrow, the fading hairline, warm text on warm dark, the
card edge accent and hover wash, restrained radii.

Changed deliberately:

- **Different serif.** Cinzel is inscriptional and reads *museum*. We use
  Cormorant Garamond as the **display** face — the same family that site uses
  for body copy, inverted into headings. Same lineage, different voice.
- **Brass, not gold.** `#C9A84C` over `#D4B765`. Slightly cooler and less
  ornamental; this is a desk, not an exhibit.
- **Serif is zoned, not global.** That site is serif everywhere because every
  page is editorial. Ours splits by surface, which is the important idea:

  | Surface | Type | Why |
  |---|---|---|
  | Landing page | Serif display + serif italic ledes | It is an argument, and it should read like one. |
  | `/taker`, `/maker` | Sans UI + mono data | These are working surfaces. A trader reading a fill confirmation wants legibility, not atmosphere. |

  Carrying the serif into the order ticket would be costume. The restraint is
  the point.

## Tokens

Defined in `app/globals.css`.

```
base    #070709   the lit-dark ground
panel   #0d0d10   raised surface
raised  #141418   double-raised (active nav, table zebra)
line    #1f1f26   hairline
strong  #2c2c36   emphasized hairline
ink     #F2EDE0   warm parchment, never #fff
muted   #B9B2A2
faint   #7A7466
accent  #C9A84C   brass — the only interactive colour
positive #5AC8A0  desaturated mint (settled, buy)
negative #CE6D6D  dusty red (failed, sell)
```

Type scale is seven roles — see the comment block in `globals.css`. Motion is
180ms for colour, 350ms for transform, matching that site's `--tr` / `--trs`.

## Ornaments

- `.eyebrow-ruled` — mono uppercase with flanking hairlines.
- `.hairline` — the fading gradient rule.
- `.lit` — the ambient radial glow, applied to page grounds.
- `.edge-*` — card top-edge accent colours.
- `.lift` — hover translate + diagonal wash.

## Rules

1. Brass is the only interactive colour. Green and red mean exactly two things
   each: buy/sell, and settled/failed. Nothing decorative is ever coloured.
2. Text is never `#fff` and the ground is never `#000`.
3. Serif never appears inside an order ticket.
4. Every number is mono and `tabular-nums`.
