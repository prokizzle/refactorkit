# MVI Architecture Pattern with Factory DI

## File Layout

```
Feature/
  FeatureModel.swift      # @Observable state
  FeatureIntent.swift     # Actions enum + handler with @Injected services
  FeatureView.swift       # Pure SwiftUI view
  FeatureContainer.swift  # Wires model + intent, provides to view
```

## Model

- Use `@Observable` (iOS 17+) or `ObservableObject` for older targets.
- Pure data holder. No business logic, no service references.
- Single source of truth for the feature's state.

```swift
@Observable
final class UserProfileModel {
    var username: String = ""
    var avatarURL: URL?
    var isLoading: Bool = false
    var errorMessage: String?
}
```

## Intent

- Define an `Action` enum covering every user interaction.
- Handler class owns `@Injected` service dependencies via Factory.
- All async work and model mutations happen here.

```swift
enum UserProfileAction {
    case loadProfile
    case updateUsername(String)
    case logout
}

final class UserProfileIntent {
    @Injected(\.userService) private var userService
    @Injected(\.authService) private var authService

    private let model: UserProfileModel

    init(model: UserProfileModel) {
        self.model = model
    }

    func send(_ action: UserProfileAction) {
        switch action {
        case .loadProfile:
            Task { await loadProfile() }
        case .updateUsername(let name):
            Task { await updateUsername(name) }
        case .logout:
            Task { await authService.signOut() }
        }
    }

    @MainActor
    private func loadProfile() async {
        model.isLoading = true
        defer { model.isLoading = false }
        do {
            let user = try await userService.fetchCurrent()
            model.username = user.name
            model.avatarURL = user.avatarURL
        } catch {
            model.errorMessage = error.localizedDescription
        }
    }

    @MainActor
    private func updateUsername(_ name: String) async {
        do {
            try await userService.updateName(name)
            model.username = name
        } catch {
            model.errorMessage = error.localizedDescription
        }
    }
}
```

## View

- Reads model directly. Sends actions through intent handler.
- No business logic, no service calls. Use design system tokens for styling.

```swift
struct UserProfileView: View {
    let model: UserProfileModel
    let intent: UserProfileIntent

    var body: some View {
        Group {
            if model.isLoading {
                ProgressView()
            } else {
                VStack(spacing: Spacing.md) {
                    AsyncImage(url: model.avatarURL)
                    Text(model.username).font(.headline)
                    Button("Logout") { intent.send(.logout) }
                }
            }
        }
        .task { intent.send(.loadProfile) }
    }
}
```

## Container

- Creates model and intent handler. Entry point for navigation.

```swift
struct UserProfileContainer: View {
    @State private var model = UserProfileModel()

    var body: some View {
        UserProfileView(
            model: model,
            intent: UserProfileIntent(model: model)
        )
    }
}
```

## Factory DI Setup

Register services in a `Container` extension. Use `.singleton` for shared state.

```swift
extension Container {
    var networkService: Factory<NetworkServiceProtocol> {
        Factory(self) { URLSessionNetworkService() }
    }
    var authService: Factory<AuthServiceProtocol> {
        Factory(self) { FirebaseAuthService() }.singleton
    }
    var userService: Factory<UserServiceProtocol> {
        Factory(self) { UserService() }
    }
}
```

## Adaptive DI for Existing Projects

| Current approach | Recommendation |
|---|---|
| Singletons / globals | Migrate to Factory incrementally |
| Manual init injection | Keep if clean; use Factory for new features |
| Another container (Swinject, Resolver) | Keep it; do not force migration |

## Rules

- One `Model`, `Intent`, `View`, `Container` per feature.
- Model never imports services. Intent never imports SwiftUI.
- Views call `intent.send(_:)` -- never call services directly.
- Prefer `@Observable` over `ObservableObject` on iOS 17+.
- Mark model-mutating intent methods `@MainActor`.
