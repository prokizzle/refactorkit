# Compose Animation Recipes (Premium Micro-Interactions)

Replaces the Pow library from the iOS version. Compose has built-in animation
APIs that cover all the same use cases without third-party dependencies (except
Lottie for pre-authored vector animations).

---

## Lottie Compose Setup

```kotlin
// build.gradle.kts
implementation("com.airbnb.android:lottie-compose:6.4.0")

// Usage
val composition by rememberLottieComposition(LottieCompositionSpec.RawRes(R.raw.animation))
val progress by animateLottieCompositionAsState(composition)
LottieAnimation(composition, { progress })
```

### Lottie Integration Points

- **Onboarding**: full-screen illustrations per step (`iterations = 1`)
- **Loading**: custom Lottie replacing `CircularProgressIndicator` for hero sections (`iterations = LottieConstants.IterateForever`)
- **Success/error**: animated checkmark/X (`iterations = 1`, `restartOnPlay = false`)
- **Empty states**: looping animated illustrations
- **Celebration**: subscription purchase confirmation overlay

---

## Screen Transitions (replaces Pow .movingParts)

```kotlin
AnimatedVisibility(
    visible = isVisible,
    enter = fadeIn(spring(stiffness = Spring.StiffnessLow)) + slideInVertically { it / 3 },
    exit = fadeOut(tween(150)) + slideOutVertically { -it / 4 }
)
```

For full-screen route transitions with Navigation Compose:

```kotlin
composable(
    route = "detail/{id}",
    enterTransition = { fadeIn(tween(300)) + slideInHorizontally { it } },
    exitTransition = { fadeOut(tween(200)) },
    popEnterTransition = { fadeIn(tween(300)) + slideInHorizontally { -it } },
    popExitTransition = { fadeOut(tween(200)) + slideOutHorizontally { it } }
)
```

---

## Shared Element Transitions (replaces swiftui-navigation-transitions)

Requires Navigation Compose 2.8+.

```kotlin
SharedTransitionLayout {
    NavHost(navController, startDestination = "list") {
        composable("list") {
            ListScreen(
                onItemClick = { id -> navController.navigate("detail/$id") },
                animatedVisibilityScope = this
            )
        }
        composable("detail/{id}") {
            DetailScreen(animatedVisibilityScope = this)
        }
    }
}

// In list item composable
Image(
    modifier = Modifier.sharedElement(
        rememberSharedContentState(key = "image-$id"),
        animatedVisibilityScope = animatedVisibilityScope
    )
)
```

---

## Button Feedback (replaces Pow .changeEffect)

```kotlin
fun Modifier.bounceClick() = composed {
    var isPressed by remember { mutableStateOf(false) }
    val scale by animateFloatAsState(
        targetValue = if (isPressed) 0.95f else 1f,
        animationSpec = spring(dampingRatio = 0.4f, stiffness = Spring.StiffnessMedium)
    )
    this
        .graphicsLayer { scaleX = scale; scaleY = scale }
        .pointerInput(Unit) {
            awaitEachGesture {
                awaitFirstDown(requireUnconsumed = false)
                isPressed = true
                waitForUpOrCancellation()
                isPressed = false
            }
        }
}
```

---

## Expanding Card (replaces Pow .glow)

```kotlin
Card(
    modifier = Modifier
        .fillMaxWidth()
        .animateContentSize(animationSpec = spring(dampingRatio = 0.7f, stiffness = Spring.StiffnessLow))
        .clickable { expanded = !expanded }
) {
    Column(Modifier.padding(16.dp)) {
        Text(title, style = MaterialTheme.typography.titleMedium)
        if (expanded) {
            Spacer(Modifier.height(8.dp))
            Text(body, style = MaterialTheme.typography.bodyMedium)
        }
    }
}
```

---

## Shimmer Loading (replaces skeleton)

```kotlin
fun Modifier.shimmer(): Modifier = composed {
    val transition = rememberInfiniteTransition(label = "shimmer")
    val offset by transition.animateFloat(
        initialValue = -300f,
        targetValue = 300f,
        animationSpec = infiniteRepeatable(tween(1200, easing = LinearEasing)),
        label = "shimmerOffset"
    )
    drawWithContent {
        drawContent()
        drawRect(
            brush = Brush.linearGradient(
                colors = listOf(Color.Transparent, Color.White.copy(alpha = 0.4f), Color.Transparent),
                start = Offset(offset, 0f),
                end = Offset(offset + 200f, 0f)
            ),
            blendMode = BlendMode.SrcAtop
        )
    }
}
```

---

## Confetti / Celebration

```kotlin
@Composable
fun ConfettiOverlay(trigger: Boolean) {
    val particles = remember { List(80) { ConfettiParticle.random() } }
    val progress by animateFloatAsState(
        targetValue = if (trigger) 1f else 0f,
        animationSpec = tween(2000, easing = LinearOutSlowInEasing)
    )
    if (progress > 0f) {
        Canvas(Modifier.fillMaxSize()) {
            particles.forEach { p ->
                val x = size.width * p.xStart + p.xDrift * progress * size.width
                val y = -40f + (size.height + 80f) * progress * p.speed
                rotate(degrees = 360f * progress * p.spin, pivot = Offset(x, y)) {
                    drawRect(p.color, Offset(x, y), Size(8.dp.toPx(), 12.dp.toPx()))
                }
            }
        }
    }
}

data class ConfettiParticle(
    val xStart: Float, val xDrift: Float, val speed: Float,
    val spin: Float, val color: Color
) {
    companion object {
        private val colors = listOf(Color(0xFFFF6B6B), Color(0xFF4ECDC4), Color(0xFFFFE66D), Color(0xFF95E1D3))
        fun random() = ConfettiParticle(
            Random.nextFloat(), (Random.nextFloat() - 0.5f) * 0.3f,
            0.6f + Random.nextFloat() * 0.4f, Random.nextFloat() * 3f,
            colors.random()
        )
    }
}
```

---

## Tab Switching

```kotlin
AnimatedContent(
    targetState = selectedTab,
    transitionSpec = {
        fadeIn(tween(250)) + slideInHorizontally { if (targetState > initialState) it / 4 else -it / 4 } togetherWith
        fadeOut(tween(150))
    },
    label = "tabContent"
) { tab ->
    when (tab) {
        0 -> HomeTab()
        1 -> StatsTab()
        2 -> SettingsTab()
    }
}
```

---

## Spring Constants for Premium Feel

```kotlin
object PremiumAnimations {
    val PremiumSpring = spring<Float>(dampingRatio = 0.6f, stiffness = Spring.StiffnessLow)
    val SnappySpring = spring<Float>(dampingRatio = 0.8f, stiffness = Spring.StiffnessMedium)
    val GentleSpring = spring<Float>(dampingRatio = 0.7f, stiffness = Spring.StiffnessVeryLow)

    const val FADE_DURATION = 200
    const val SLIDE_DURATION = 350
}
```

Use `PremiumSpring` as the default for interactive elements. Use `SnappySpring`
for toggles and quick responses. Use `GentleSpring` for background/ambient
motion like parallax or breathing effects.
