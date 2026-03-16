# Tuist Configuration Reference

## Project.swift Template

```swift
import ProjectDescription

let project = Project(
    name: "FeelingMindful",
    settings: .settings(
        base: [
            "DEVELOPMENT_TEAM": "YOUR_TEAM_ID"
        ],
        configurations: [
            .debug(name: "Debug"),
            .release(name: "Release"),
        ]
    ),
    targets: [
        .target(
            name: "FeelingMindful",
            destinations: .iOS,
            product: .app,
            bundleId: "com.yourcompany.feelingmindful",
            infoPlist: .extendingDefault(with: [
                "UIAppFonts": .array([
                    .string("CustomFont-Regular.ttf"),
                    .string("CustomFont-Bold.ttf"),
                ]),
            ]),
            sources: ["Sources/**"],
            resources: [
                "Resources/**",
                "StoreKitConfiguration.storekit",
            ],
            dependencies: [
                .external(name: "SwiftUIDesignSystemPro"),
                .external(name: "Lottie"),
                .external(name: "Pow"),
                .external(name: "Factory"),
                .external(name: "Nuke"),
                .external(name: "NavigationTransitions"),
            ],
            settings: .settings(base: [
                "CODE_SIGN_STYLE": "Automatic",
            ])
        ),
    ]
)
```

## Tuist/Package.swift (External Dependencies)

Place this file at `Tuist/Package.swift`:

```swift
// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "Dependencies",
    dependencies: [
        .package(url: "https://github.com/muhittincamdali/SwiftUI-Design-System-Pro", from: "2.0.0"),
        .package(url: "https://github.com/airbnb/lottie-ios", from: "4.4.0"),
        .package(url: "https://github.com/EmergeTools/Pow", from: "1.0.0"),
        .package(url: "https://github.com/hmlongco/Factory", from: "2.3.0"),
        .package(url: "https://github.com/kean/Nuke", from: "12.0.0"),
        .package(url: "https://github.com/davdroman/swiftui-navigation-transitions", from: "0.13.0"),
    ]
)
```

## Tuist Cache Setup

- Run `tuist cache` to pre-build dependencies as xcframeworks.
- Cached frameworks skip recompilation on `tuist generate`.
- On CI, cache the `.tuist/Cache` directory between runs.

## CLAUDE.md Configuration

Add this to the project CLAUDE.md:

```
## Build Commands
- Install dependencies: `tuist install`
- Generate project: `tuist generate`
- Build: `tuist build`
- Test: `tuist test`
- Cache dependencies: `tuist cache`

Do NOT open .xcodeproj directly. Always use `tuist generate` first.
```

## Common Tuist Issues

| Problem | Fix |
|---|---|
| Team ID lost on generate | Set `DEVELOPMENT_TEAM` in Project.swift `settings`, not in Xcode |
| Signing errors | Use `CODE_SIGN_STYLE: Automatic` in target settings |
| SPM dependency not found | Run `tuist install` before `tuist generate` |
| Stale project state | Delete `Derived/` and re-run `tuist generate` |
| Wrong product name in `.external()` | Check the library's Package.swift for the exact target name |

## Workflow

1. `tuist install` -- resolves and fetches SPM packages
2. `tuist cache` -- (optional) pre-builds dependencies
3. `tuist generate` -- creates .xcodeproj/.xcworkspace
4. `tuist build` -- builds the project
