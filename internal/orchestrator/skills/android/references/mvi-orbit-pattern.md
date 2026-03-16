# MVI Architecture with Orbit + Hilt

## Structure

Each feature follows this layout:

```
feature/
  FeatureState.kt        # Data class with UI state
  FeatureSideEffect.kt   # Sealed class for one-shot events (navigation, toasts)
  FeatureViewModel.kt    # Orbit ContainerHost with @Inject constructor
  FeatureScreen.kt       # Composable, collectAsState, collectSideEffect
```

## State

Immutable data class. Single source of truth for the screen.

```kotlin
data class ProfileState(
    val user: User? = null,
    val isLoading: Boolean = false,
    val error: String? = null
)
```

All fields must have defaults so `ProfileState()` produces a valid initial state.

## Side Effects

Sealed interface for one-shot events that must not survive recomposition.

```kotlin
sealed interface ProfileSideEffect {
    data class ShowToast(val message: String) : ProfileSideEffect
    data object NavigateBack : ProfileSideEffect
}
```

Use side effects for: navigation, snackbars/toasts, analytics events, one-time dialogs.
Do NOT use side effects for anything the UI should re-render on recomposition.

## ViewModel (Orbit ContainerHost)

```kotlin
@HiltViewModel
class ProfileViewModel @Inject constructor(
    private val userRepository: UserRepository
) : ViewModel(), ContainerHost<ProfileState, ProfileSideEffect> {

    override val container = container<ProfileState, ProfileSideEffect>(ProfileState())

    init {
        loadUser()
    }

    fun loadUser() = intent {
        reduce { state.copy(isLoading = true, error = null) }
        runCatching { userRepository.getUser() }
            .onSuccess { user ->
                reduce { state.copy(user = user, isLoading = false) }
            }
            .onFailure { e ->
                reduce { state.copy(error = e.message, isLoading = false) }
                postSideEffect(ProfileSideEffect.ShowToast("Failed to load profile"))
            }
    }

    fun onBackPressed() = intent {
        postSideEffect(ProfileSideEffect.NavigateBack)
    }
}
```

Key rules:
- All mutations happen inside `intent { }` blocks.
- Use `reduce { state.copy(...) }` to update state. Never mutate state directly.
- Use `postSideEffect()` for one-shot events only.
- Keep `intent` blocks small; delegate business logic to repositories/use cases.

## Screen (Composable)

```kotlin
@Composable
fun ProfileScreen(
    viewModel: ProfileViewModel = hiltViewModel(),
    onNavigateBack: () -> Unit
) {
    val state by viewModel.collectAsState()

    viewModel.collectSideEffect { sideEffect ->
        when (sideEffect) {
            is ProfileSideEffect.ShowToast -> {
                // show toast or snackbar
            }
            ProfileSideEffect.NavigateBack -> onNavigateBack()
        }
    }

    ProfileContent(
        state = state,
        onRetry = viewModel::loadUser,
        onBack = viewModel::onBackPressed
    )
}

@Composable
private fun ProfileContent(
    state: ProfileState,
    onRetry: () -> Unit,
    onBack: () -> Unit
) {
    // Pure UI. No ViewModel reference. Testable in previews.
}
```

Split into a wiring composable (handles VM + side effects) and a pure content composable (stateless, previewable).

## Hilt DI Setup

Application class:

```kotlin
@HiltAndroidApp
class App : Application()
```

Activity:

```kotlin
@AndroidEntryPoint
class MainActivity : ComponentActivity()
```

Module for providing dependencies:

```kotlin
@Module
@InstallIn(SingletonComponent::class)
object DataModule {

    @Provides
    @Singleton
    fun provideUserRepository(api: ApiService): UserRepository {
        return UserRepositoryImpl(api)
    }
}
```

Use `@InstallIn(SingletonComponent::class)` for app-scoped singletons.
Use `@InstallIn(ViewModelComponent::class)` for ViewModel-scoped dependencies.

## Gradle Dependencies

```kotlin
// Orbit MVI
implementation("org.orbit-mvi:orbit-core:<version>")
implementation("org.orbit-mvi:orbit-viewmodel:<version>")
implementation("org.orbit-mvi:orbit-compose:<version>")

// Hilt
implementation("com.google.dagger:hilt-android:<version>")
kapt("com.google.dagger:hilt-android-compiler:<version>")
implementation("androidx.hilt:hilt-navigation-compose:<version>")
```

## Testing

```kotlin
@Test
fun `loadUser sets user on success`() = runTest {
    val repo = FakeUserRepository(user = testUser)
    val vm = ProfileViewModel(repo)

    vm.test(this) {
        expectInitialState()
        expectState { copy(isLoading = true) }
        expectState { copy(user = testUser, isLoading = false) }
    }
}
```

Use `orbit-test` artifact. Create the ViewModel directly with fakes -- no Hilt in unit tests.

## Adaptive DI for Existing Projects

**Koin in place:** Keep Koin for existing features. Use Hilt for new features. Both can coexist. Migrate incrementally if desired.

**Manual DI in place:** Add Hilt to the app module first. Migrate one feature at a time by adding `@HiltViewModel` and `@Inject constructor`.

**No DI:** Add Hilt from scratch following the setup above.

In all cases, repositories and data sources should be injected -- never instantiated inside ViewModels.
