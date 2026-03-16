# Premium Material 3 Theme Reference

## Foundation

Custom Material 3 theme that replaces every default token. The "premium" factor comes from intentional color, type, and shape customization -- not from Material's default purple palette.

## Color Scheme Presets

### Serene (Wellness)

```kotlin
private val SereneLightColors = lightColorScheme(
    primary = Color(0xFF5B8A72),        // sage green
    onPrimary = Color(0xFFFFFFFF),
    secondary = Color(0xFF7BA7BC),       // soft blue
    onSecondary = Color(0xFFFFFFFF),
    tertiary = Color(0xFFC4A882),        // warm neutral
    background = Color(0xFFF8F6F2),      // warm white
    surface = Color(0xFFFFFFFF),
    error = Color(0xFFBA4B4B),
)

private val SereneDarkColors = darkColorScheme(
    primary = Color(0xFF8FBFA6),
    onPrimary = Color(0xFF1B3528),
    secondary = Color(0xFFA3C9D9),
    onSecondary = Color(0xFF0E2A35),
    tertiary = Color(0xFFD9C4A5),
    background = Color(0xFF1A1C1B),
    surface = Color(0xFF222524),
    error = Color(0xFFE5736E),
)
```

### Bold (Finance)

```kotlin
private val BoldLightColors = lightColorScheme(
    primary = Color(0xFF1B2D4F),        // deep navy
    onPrimary = Color(0xFFFFFFFF),
    secondary = Color(0xFFC9983A),       // gold accent
    onSecondary = Color(0xFF1A1A1A),
    tertiary = Color(0xFF3D6B8E),        // steel blue
    background = Color(0xFFFAFAFA),      // crisp white
    surface = Color(0xFFFFFFFF),
    error = Color(0xFFC62828),
)
```

### Vivid (Lifestyle)

```kotlin
private val VividLightColors = lightColorScheme(
    primary = Color(0xFFE8725A),        // coral
    onPrimary = Color(0xFFFFFFFF),
    secondary = Color(0xFF2A9D8F),       // teal
    onSecondary = Color(0xFFFFFFFF),
    tertiary = Color(0xFFF4A261),        // warm amber
    background = Color(0xFFFFF8F5),      // warm tint
    surface = Color(0xFFFFFFFF),
    error = Color(0xFFD32F2F),
)
```

## Typography with Google Fonts

```kotlin
val provider = GoogleFont.Provider(
    providerAuthority = "com.google.android.gms.fonts",
    providerPackage = "com.google.android.gms",
    certificates = R.array.com_google_android_gms_fonts_certs
)

val interFamily = FontFamily(
    Font(GoogleFont("Inter"), provider, FontWeight.Normal),
    Font(GoogleFont("Inter"), provider, FontWeight.Medium),
    Font(GoogleFont("Inter"), provider, FontWeight.SemiBold),
    Font(GoogleFont("Inter"), provider, FontWeight.Bold),
)

val PremiumTypography = Typography(
    displayLarge = TextStyle(fontFamily = interFamily, fontWeight = FontWeight.Bold, fontSize = 34.sp, lineHeight = 40.sp),
    headlineMedium = TextStyle(fontFamily = interFamily, fontWeight = FontWeight.SemiBold, fontSize = 24.sp, lineHeight = 32.sp),
    titleLarge = TextStyle(fontFamily = interFamily, fontWeight = FontWeight.SemiBold, fontSize = 20.sp, lineHeight = 28.sp),
    bodyLarge = TextStyle(fontFamily = interFamily, fontWeight = FontWeight.Normal, fontSize = 16.sp, lineHeight = 24.sp),
    bodyMedium = TextStyle(fontFamily = interFamily, fontWeight = FontWeight.Normal, fontSize = 14.sp, lineHeight = 20.sp),
    labelLarge = TextStyle(fontFamily = interFamily, fontWeight = FontWeight.Medium, fontSize = 14.sp, lineHeight = 20.sp, letterSpacing = 0.1.sp),
)
```

## Custom Shapes

```kotlin
val PremiumShapes = Shapes(
    extraSmall = RoundedCornerShape(6.dp),
    small = RoundedCornerShape(10.dp),
    medium = RoundedCornerShape(16.dp),
    large = RoundedCornerShape(24.dp),
    extraLarge = RoundedCornerShape(32.dp),
)
```

## Theme Composable

```kotlin
@Composable
fun PremiumTheme(
    preset: ThemePreset = ThemePreset.Serene,
    darkTheme: Boolean = isSystemInDarkTheme(),
    dynamicColor: Boolean = false,
    content: @Composable () -> Unit
) {
    val colorScheme = when {
        dynamicColor && Build.VERSION.SDK_INT >= Build.VERSION_CODES.S -> {
            val context = LocalContext.current
            if (darkTheme) dynamicDarkColorScheme(context) else dynamicLightColorScheme(context)
        }
        darkTheme -> preset.darkColors
        else -> preset.lightColors
    }

    MaterialTheme(
        colorScheme = colorScheme,
        typography = PremiumTypography,
        shapes = PremiumShapes,
        content = content
    )
}
```

## Iconography: No Material Icons Defaults, No Emojis

AI-coded apps are instantly recognizable by their heavy reliance on `Icons.Default.*` and emoji. Premium apps use original artwork.

**Rules:**
- NEVER use emojis in UI (no headings, buttons, labels, or list items)
- MINIMIZE `Icons.Default.*` / `Icons.Outlined.*` usage — only for system-standard actions (back arrow, share, settings) where users expect the platform icon
- PREFER custom icons generated with Nova Canvas or sourced from a custom icon set
- PREFER Lottie animated icons over static ones for key actions
- For bottom nav, toolbars, and feature icons: generate custom icons via Bedrock

**Where Material Icons are acceptable:**
- Navigation back/up arrows
- System share icon
- Standard toolbar actions (search, more options)

**Where Material Icons must be replaced:**
- Bottom navigation bar icons (generate custom set via Bedrock)
- Feature/section icons throughout the app
- Onboarding step icons
- Settings menu icons
- Empty state illustrations (use Lottie or Bedrock instead)

**Generation approach:**
- Use Nova Canvas to generate a consistent icon set matching the theme
- Prompt: "minimal vector icon, single color, [theme-accent-color], transparent background, [subject], app icon style, no text"
- Generate all icons in one session for visual consistency
- Export as SVG or WebP for Android

## Premium vs Vibe-Coded Checklist

- Custom color scheme (not Material default purple/pink)
- Custom font family via GoogleFont provider (not Roboto)
- Intentional shape tokens with varied radii (not 4dp everywhere)
- Consistent elevation using `tonalElevation` and `shadowElevation`
- Dynamic color support for Android 12+ with fallback to custom seed
- Dark theme that is hand-tuned, not auto-inverted
- Color roles mapped correctly: onPrimary, onSecondary, surfaceVariant
- No emojis anywhere in the UI
- Custom icons instead of Material Icons defaults for app-defining visuals
- Bottom nav uses custom icon set, not Icons.Default

## Lottie Integration Points

Use `com.airbnb.lottie:lottie-compose`. Key placements:

- **Onboarding illustrations** -- full-screen Lottie replacing static images
- **Loading states** -- `LottieAnimation(composition, iterations = LottieConstants.IterateForever)` replacing `CircularProgressIndicator`
- **Success/error feedback** -- short one-shot animations after form submission
- **Empty states** -- subtle loop animation on empty list screens
- **Celebration moments** -- confetti or checkmark on goal completion

```kotlin
val composition by rememberLottieComposition(LottieCompositionSpec.RawRes(R.raw.success))
LottieAnimation(composition, iterations = 1, modifier = Modifier.size(120.dp))
```

## Compose Animation Recipes

### AnimatedVisibility with custom transitions

```kotlin
AnimatedVisibility(
    visible = isVisible,
    enter = fadeIn(tween(300)) + slideInVertically(initialOffsetY = { it / 4 }),
    exit = fadeOut(tween(200)) + slideOutVertically(targetOffsetY = { -it / 4 })
) { Content() }
```

### Shared element transitions

```kotlin
SharedTransitionLayout {
    AnimatedContent(targetState = showDetail) { target ->
        if (!target) {
            Card(modifier = Modifier.sharedElement(
                rememberSharedContentState("image-$id"), this@AnimatedContent
            )) { Thumbnail() }
        } else {
            DetailScreen(modifier = Modifier.sharedElement(
                rememberSharedContentState("image-$id"), this@AnimatedContent
            ))
        }
    }
}
```

### Expanding card with animateContentSize

```kotlin
Card(modifier = Modifier.animateContentSize(spring(dampingRatio = 0.7f, stiffness = 300f))) {
    Text(title)
    if (expanded) { Text(details) }
}
```

### Button tap feedback (scale + alpha)

```kotlin
val interactionSource = remember { MutableInteractionSource() }
val isPressed by interactionSource.collectIsPressedAsState()
val scale by animateFloatAsState(if (isPressed) 0.95f else 1f, spring(stiffness = 600f))
val alpha by animateFloatAsState(if (isPressed) 0.85f else 1f)

Button(
    onClick = onClick,
    interactionSource = interactionSource,
    modifier = Modifier.graphicsLayer { scaleX = scale; scaleY = scale; this.alpha = alpha }
) { Text(label) }
```

### Custom spring for nav transitions

```kotlin
val springSpec = spring<Float>(dampingRatio = 0.8f, stiffness = 200f)
```

Use `dampingRatio` 0.6-0.8 for playful bounce, 0.9-1.0 for professional settle. `stiffness` 100-300 for smooth, 400+ for snappy.
