---
name: Mythic Journey Redux
colors:
  surface: '#131313'
  surface-dim: '#131313'
  surface-bright: '#393939'
  surface-container-lowest: '#0e0e0e'
  surface-container-low: '#1c1b1b'
  surface-container: '#20201f'
  surface-container-high: '#2a2a2a'
  surface-container-highest: '#353535'
  on-surface: '#e5e2e1'
  on-surface-variant: '#d0c6ab'
  inverse-surface: '#e5e2e1'
  inverse-on-surface: '#313030'
  outline: '#999077'
  outline-variant: '#4d4732'
  surface-tint: '#e9c400'
  primary: '#fff6df'
  on-primary: '#3a3000'
  primary-container: '#ffd700'
  on-primary-container: '#705e00'
  inverse-primary: '#705d00'
  secondary: '#ffb4a9'
  on-secondary: '#690002'
  secondary-container: '#970205'
  on-secondary-container: '#ff9f92'
  tertiary: '#dbffe6'
  on-tertiary: '#003921'
  tertiary-container: '#6ff2ae'
  on-tertiary-container: '#006d44'
  error: '#ffb4ab'
  on-error: '#690005'
  error-container: '#93000a'
  on-error-container: '#ffdad6'
  primary-fixed: '#ffe16d'
  primary-fixed-dim: '#e9c400'
  on-primary-fixed: '#221b00'
  on-primary-fixed-variant: '#544600'
  secondary-fixed: '#ffdad5'
  secondary-fixed-dim: '#ffb4a9'
  on-secondary-fixed: '#410001'
  on-secondary-fixed-variant: '#930004'
  tertiary-fixed: '#78fbb6'
  tertiary-fixed-dim: '#59de9b'
  on-tertiary-fixed: '#002111'
  on-tertiary-fixed-variant: '#005232'
  background: '#131313'
  on-background: '#e5e2e1'
  surface-variant: '#353535'
  quality-ur: '#FF4500'
  quality-ssr: '#FFD700'
  quality-sr: '#DA70D6'
  quality-r: '#1E90FF'
  quality-n: '#A9A9A9'
  ink-wash: '#2D2D2D'
  parchment: '#F5E6BE'
typography:
  display-hero:
    fontFamily: Epilogue
    fontSize: 48px
    fontWeight: '800'
    lineHeight: '1.1'
    letterSpacing: -0.02em
  headline-lg:
    fontFamily: Epilogue
    fontSize: 32px
    fontWeight: '700'
    lineHeight: '1.2'
  headline-lg-mobile:
    fontFamily: Epilogue
    fontSize: 24px
    fontWeight: '700'
    lineHeight: '1.2'
  title-md:
    fontFamily: Epilogue
    fontSize: 20px
    fontWeight: '600'
    lineHeight: '1.4'
  body-md:
    fontFamily: Hanken Grotesk
    fontSize: 16px
    fontWeight: '400'
    lineHeight: '1.6'
  stats-num:
    fontFamily: JetBrains Mono
    fontSize: 18px
    fontWeight: '600'
    lineHeight: '1'
  label-sm:
    fontFamily: JetBrains Mono
    fontSize: 12px
    fontWeight: '500'
    lineHeight: '1.2'
    letterSpacing: 0.05em
rounded:
  sm: 0.125rem
  DEFAULT: 0.25rem
  md: 0.375rem
  lg: 0.5rem
  xl: 0.75rem
  full: 9999px
spacing:
  unit: 4px
  gutter: 16px
  margin-mobile: 20px
  margin-desktop: 40px
  stack-xs: 4px
  stack-md: 12px
  stack-lg: 24px
---

## Brand & Style
The design system embodies a **Mythical Oriental** aesthetic, blending the grandeur of ancient Chinese legends with a sleek, modern mobile RPG interface. It aims to evoke a sense of divine power, ancient mystery, and high-stakes adventure.

The style is a hybrid of **Corporate/Modern** systematic layouts and **Tactile/Skeuomorphic** elements. While the information architecture remains clean and functional (Modern), the UI surfaces use physical metaphors like carved stone, polished jade, and brushed ink (Tactile). High-contrast accents and ornate decorative borders differentiate the experience from standard flat SaaS-like games, creating an immersive "Westward Journey" atmosphere.

## Colors
The palette is rooted in traditional imperial symbolism. **Imperial Gold** is used for primary actions, currency, and high-tier highlights. **Vermilion Red** serves as the secondary color for combat actions, alerts, and vital UI markers. **Jade Green** is reserved for growth, health, and positive progression.

The background is dominated by **Deep Ink Black**, providing a high-contrast canvas that allows the metallic and jewel-toned colors to glow. For rarity tiers (UR to N), the design system utilizes a specific high-chroma spectrum to ensure instant recognition during gacha and inventory management.

## Typography
The typography strategy balances cultural flair with functional clarity. 
- **Headlines:** Use **Epilogue**. Its geometric but distinctive character emulates the weight of calligraphic strokes while remaining perfectly legible on high-resolution screens.
- **Body Text:** Use **Hanken Grotesk** for long-form descriptions and task lists. It provides a clean, contemporary contrast to the ornate decorative elements.
- **Data & Numbers:** Use **JetBrains Mono**. This monospaced font ensures that hero stats, gold counts, and timers remain aligned and easy to scan during fast-paced gameplay.

## Layout & Spacing
This design system uses a **Fluid Grid** model with a base unit of **4px**. 

- **Mobile:** A 4-column grid with 20px side margins. Content cards typically span the full width or 2 columns.
- **Desktop/Tablet:** A 12-column grid with 40px margins. 
- **Rhythm:** Vertical spacing follows a "Stack" philosophy. Use `stack-xs` for related labels/icons, `stack-md` for elements within a card, and `stack-lg` to separate distinct functional sections (e.g., Hero Info vs. Skill List).

The layout should prioritize a "Bottom-Heavy" thumb zone for mobile navigation, with hero avatars and immersive backgrounds occupying the upper two-thirds of the screen.

## Elevation & Depth
Depth is conveyed through **Tonal Layering** and **Tactile Textures** rather than traditional soft shadows.

1.  **Base Layer:** Deep Ink Black background, often with a subtle "Silk" or "Rice Paper" texture overlay at 5% opacity.
2.  **Surface Layer:** "Ink Wash" dark grays (#2D2D2D) used for container backgrounds.
3.  **Raised Layer:** Interactive elements like buttons use stone-texture bitmaps or high-contrast Vermilion borders.
4.  **Divine Glow:** Critical UI elements (UR Heroes, Active Quests) utilize an outer glow (bloom effect) in Imperial Gold rather than a drop shadow, simulating a spiritual aura.

## Shapes
The shape language is **Soft (0.25rem)**, moving away from ultra-round modern trends to maintain a sense of "carved" structural integrity. 

- **Containers:** Use 4px (Soft) corners to mimic cut stone.
- **Ornate Details:** Use vector-based "Cloud" or "Dragon" flourishes on the corners of major panels to reinforce the theme.
- **Progress Bars:** Should be strictly rectangular or slightly tapered at the ends to resemble jade tablets.

## Components

- **Stone Buttons:** Primary buttons feature a dark granite texture with inner-beveled edges. On-press, they shift slightly downward with a Gold inner-glow to indicate "activation."
- **Jade Progress Bars:** Used for XP and Stamina. The "fill" uses a vertical gradient of Jade Green (#00A86B) with a high-gloss "glass" sheen, while the "track" is a carved hollow in the UI surface.
- **Ornate Borders:** All major modal windows must be framed by a 2px Imperial Gold border, featuring interlocking "Meander" patterns at the four corners.
- **Status Badges (Rarity):** High-contrast, pill-shaped tags. UR/SSR badges should include a "shimmer" CSS animation that sweeps across the text.
- **Input Fields:** Dark "Ink Wash" backgrounds with a simple bottom-border in Gold that glows when the field is focused.
- **Cards (Hero/Item):** Hero cards use the full artwork as the background, with a Vermilion-to-Transparent gradient at the bottom to house the Hanken Grotesk name and Level labels.