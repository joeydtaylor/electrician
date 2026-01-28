# Steeze Brand + UI System

This repository follows the Steeze UI/UX system for any UI assets or front-end
code that may be added here. Use Steeze-provided assets, tokens, and components
by default. Only create new UI primitives if there is no existing Steeze
counterpart.

> Note: this backend repo does not currently ship UI assets. If UI code or
> assets are added, they must follow the layout and documentation rules below.

---

## What goes where

### Brand (static assets)
`/brand/assets/` is the single source of truth for static design assets.

- `brand/assets/logos/` Steeze icon, lockups (dark/light), mono/inverse.
- `brand/assets/icons/` Topbar/menu icons (home/user/shield/moon/logout).
- `brand/assets/backgrounds/` Non-tiling hero + corner backgrounds (dark/light).
- `brand/assets/motion/` Loaders/spinners/skeleton CSS + SVGs.

### Tokens (design tokens)
`/brand/tokens/` contains canonical tokens:
- `tokens.json` (canonical)
- `tokens.css` (CSS variables for non-MUI surfaces)

### UI system (code)
`/ui/steeze/` contains reusable React/MUI code:
- `theme/` MUI theme + component overrides (dark/light)
- `components/` reusable primitives (TopbarControls, ProfileMenu, StateRail)
- `docs/` usage docs + patterns

---

## Fortune-50 UX rules (baseline)
1) Consistency > novelty.
2) One animated element per region (no shimmer + spinner + pulse stacks).
3) Cyan is a signal, not the default paint bucket.
4) No tiled backgrounds behind content.
5) Admin/support/impersonation must be globally obvious.
6) Prefer skeletons and indeterminate bars over spinners for data surfaces.
7) Respect `prefers-reduced-motion` globally.

---

## Documentation requirement
Every custom Steeze asset/component must include:
- When to use / when not to use
- Usage snippet
- Tokens it uses
- A11y notes (role/aria, reduced motion)

Docs locations:
- `brand/docs/` for static assets
- `ui/steeze/docs/` for components/patterns
