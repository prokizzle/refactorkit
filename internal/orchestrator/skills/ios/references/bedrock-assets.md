# AWS Bedrock Asset Generation Reference

## Prerequisites
- nova-canvas MCP server configured with AWS_PROFILE=bedrock
- AWS account with Bedrock model access enabled

## Illustration Categories

Define prompt templates for each category:

### Onboarding Illustrations
- Style: "minimal flat illustration, soft gradients, [theme-color-palette], clean lines, modern app style"
- Subjects: welcome, feature highlight, getting started, permissions explanation

### Empty State Illustrations
- Style: "friendly minimal illustration, muted colors, [theme-color-palette], centered composition"
- Subjects: no items yet, no search results, offline state, first-time use

### Feature Header Illustrations
- Style: "abstract geometric illustration, [theme-color-palette], subtle depth, app header crop ratio 3:1"
- Subjects: dashboard, settings, profile, activity

### Error State Illustrations
- Style: "empathetic minimal illustration, warm tones, [theme-color-palette]"
- Subjects: connection error, server error, not found, expired session

### Custom Icon Set (replaces SF Symbols / Material Icons)
- Style: "minimal vector icon, single color, [theme-accent-color], transparent background, app icon style, no text, no gradients"
- Generate ALL app icons in one batch session for visual consistency
- Subjects: tab bar icons, feature section icons, settings menu icons, action icons
- Dimensions: 1024x1024 (scale down to @2x/@3x or mdpi-xxxhdpi)
- Export: PDF vector (iOS) or SVG/WebP (Android)
- Generate a complete set of 15-30 icons covering all app sections

## Style Guide for Consistency
- Always include theme color palette in prompt
- Use consistent style keywords across all prompts
- Request specific dimensions: 1024x1024 for illustrations, 1024x341 for headers
- Include "no text, no words, no letters" to avoid baked-in text

## Nuke Caching Strategy
- Cache generated assets locally after first generation
- Use Nuke's DataCache for disk persistence
- LazyImage for displaying in SwiftUI views
- Placeholder with DSSkeleton while loading

## Optimization
- Generate @2x and @3x variants
- Export as PNG for illustrations, WebP for photos
- Add to Asset Catalog via Tuist configuration
- Consider generating at build time vs runtime
