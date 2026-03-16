# Premium Design System Reference

## Foundation: SwiftUI-Design-System-Pro

- Import: `import DesignSystemPro` (or modular: `Core`, `Components`, `Theme`)
- Wrap root view in `DSApp { }` for theme injection

## Design Tokens Quick Reference

| Category           | Examples                                                        |
|--------------------|-----------------------------------------------------------------|
| ColorTokens        | `Accent.primary`, `Background.primary`, `Status.success`, `Foreground.primary` |
| TypographyTokens   | `Display.large` (48pt), `Heading.h1` (24pt), `Body.medium` (16pt base), `UI.button` (14pt) |
| SpacingScale       | `xxs` (4pt), `sm` (8pt base), `lg` (16pt most common), `xxxl` (32pt) |
| ShadowTokens       | `card`, `level1`-`level6`                                       |
| AnimationTokens    | `micro` (button feedback), `page` (screen transitions)          |

## Theme Presets

Map plugin themes to DesignSystemPro presets:

| Plugin Theme       | Base              | Characteristics                          |
|--------------------|-------------------|------------------------------------------|
| serene (wellness)  | Ocean/Forest      | Soft gradients, organic shapes           |
| bold (finance)     | Dark theme        | Sharp corners, gold accents              |
| vivid (lifestyle)  | Sunset/Lavender   | Vibrant colors, rounded shapes           |
| custom             | `WhiteLabelConfiguration` with `BrandConfiguration` |

## Lottie Integration Points

Use Lottie animations at these touchpoints:

- **Onboarding** -- full-screen animated illustrations per step
- **Loading states** -- custom Lottie replacing `DSSkeleton` shimmer for hero sections
- **Success/error** -- animated checkmark/X replacing static icons
- **Empty states** -- animated illustrations instead of static images
- **Pull-to-refresh** -- custom Lottie animation
- **Celebration** -- subscription purchase, achievement unlocked

## Pow Integration Points

Use Pow effects at these touchpoints:

- **Screen transitions** -- `.movingParts.blur`, `.movingParts.swoosh`
- **Button feedback** -- `.changeEffect(.spray)`, `.changeEffect(.rise)`
- **Card interactions** -- `.changeEffect(.glow)` on tap
- **Tab switching** -- `.transition(.movingParts.iris)`
- **Notification entry** -- `.transition(.movingParts.clock)`
- **Delete confirmation** -- `.changeEffect(.shake)`

## Typography with Custom Fonts

1. Download via `google-font-downloader` skill
2. Convert woff2 to ttf with `fonttools`
3. Register in `Info.plist` `UIAppFonts` (via Tuist config)
4. Map to DesignSystemPro typography tokens via `WhiteLabelConfiguration` `fontFamily`

## Iconography: No SF Symbols, No Emojis

AI-coded apps are instantly recognizable by their heavy reliance on SF Symbols and emoji. Premium apps use original artwork.

**Rules:**
- NEVER use emojis in UI (no headings, buttons, labels, or list items)
- MINIMIZE SF Symbols usage — only acceptable for system-standard actions (back arrow, share, settings gear) where users expect the system icon
- PREFER custom icons generated with Nova Canvas or sourced from a custom icon set
- PREFER Lottie animated icons over static ones for key actions
- For tab bars, toolbars, and feature icons: generate custom vector-style icons via Bedrock

**Where SF Symbols are acceptable:**
- Navigation bar back/forward arrows
- System share sheet icon
- Standard toolbar actions (compose, delete, search)
- Accessibility-critical system icons

**Where SF Symbols must be replaced:**
- Tab bar icons (generate custom set via Bedrock)
- Feature/section icons throughout the app
- Onboarding step icons
- Settings menu icons
- Empty state illustrations (use Lottie or Bedrock instead)
- Any icon that defines the app's visual identity

**Generation approach:**
- Use Nova Canvas to generate a consistent icon set matching the theme
- Prompt: "minimal vector icon, single color, [theme-accent-color], transparent background, [subject], app icon style, no text"
- Generate all icons in one session for visual consistency
- Export as PDF vector for resolution independence

## What Makes It "Premium" (Not Vibe-Coded)

- [ ] Intentional spacing using 8pt grid tokens, not arbitrary padding
- [ ] Semantic colors via tokens, not hardcoded hex values
- [ ] Consistent elevation via shadow tokens across all cards
- [ ] Custom typography (not SF Pro default)
- [ ] Lottie animations for key moments (not just stock transitions)
- [ ] Pow micro-interactions on every interactive element
- [ ] Category-specific theme applied via tokens, not manual styling
- [ ] Custom illustrations (Bedrock) not stock photos
- [ ] No emojis anywhere in the UI
- [ ] Custom icons instead of SF Symbols for app-defining visuals
- [ ] Tab bar uses custom icon set, not SF Symbols
