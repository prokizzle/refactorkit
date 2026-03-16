# Gradle Configuration Reference

## settings.gradle.kts

```kotlin
pluginManagement {
    repositories {
        google()
        mavenCentral()
        gradlePluginPortal()
    }
}

plugins {
    id("dev.tuist") version "0.2.2"
}

dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.FAIL_ON_PROJECT_REPOS)
    repositories {
        google()
        mavenCentral()
    }
}

rootProject.name = "FeelingMindful"
include(":app")
```

## gradle/libs.versions.toml

```toml
[versions]
kotlin = "2.1.0"
agp = "8.7.3"
compose-bom = "2025.01.00"
hilt = "2.51"
orbit = "9.0.0"
lottie = "6.4.0"
coil = "2.6.0"
billing = "7.0.0"
ksp = "2.1.0-1.0.29"
core-ktx = "1.15.0"
lifecycle = "2.8.7"
activity-compose = "1.9.3"
navigation = "2.8.5"

[libraries]
compose-bom = { group = "androidx.compose", name = "compose-bom", version.ref = "compose-bom" }
compose-material3 = { group = "androidx.compose.material3", name = "material3" }
compose-ui-tooling-preview = { group = "androidx.compose.ui", name = "ui-tooling-preview" }
compose-ui-tooling = { group = "androidx.compose.ui", name = "ui-tooling" }
compose-navigation = { group = "androidx.navigation", name = "navigation-compose", version.ref = "navigation" }
core-ktx = { group = "androidx.core", name = "core-ktx", version.ref = "core-ktx" }
lifecycle-runtime-compose = { group = "androidx.lifecycle", name = "lifecycle-runtime-compose", version.ref = "lifecycle" }
activity-compose = { group = "androidx.activity", name = "activity-compose", version.ref = "activity-compose" }
hilt-android = { group = "com.google.dagger", name = "hilt-android", version.ref = "hilt" }
hilt-compiler = { group = "com.google.dagger", name = "hilt-android-compiler", version.ref = "hilt" }
hilt-navigation-compose = { group = "androidx.hilt", name = "hilt-navigation-compose" }
orbit-core = { group = "org.orbit-mvi", name = "orbit-core", version.ref = "orbit" }
orbit-viewmodel = { group = "org.orbit-mvi", name = "orbit-viewmodel", version.ref = "orbit" }
orbit-compose = { group = "org.orbit-mvi", name = "orbit-compose", version.ref = "orbit" }
lottie-compose = { group = "com.airbnb.android", name = "lottie-compose", version.ref = "lottie" }
coil-compose = { group = "io.coil-kt", name = "coil-compose", version.ref = "coil" }
billing = { group = "com.android.billingclient", name = "billing-ktx", version.ref = "billing" }

[plugins]
android-application = { id = "com.android.application", version.ref = "agp" }
kotlin-android = { id = "org.jetbrains.kotlin.android", version.ref = "kotlin" }
kotlin-compose = { id = "org.jetbrains.kotlin.plugin.compose", version.ref = "kotlin" }
hilt = { id = "com.google.dagger.hilt.android", version.ref = "hilt" }
ksp = { id = "com.google.devtools.ksp", version.ref = "ksp" }
```

## app/build.gradle.kts

```kotlin
plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.kotlin.compose)
    alias(libs.plugins.hilt)
    alias(libs.plugins.ksp)
}

android {
    namespace = "com.feelingmindful"
    compileSdk = 35

    defaultConfig {
        applicationId = "com.feelingmindful"
        minSdk = 26
        targetSdk = 35
        versionCode = 1
        versionName = "1.0.0"
    }

    signingConfigs {
        getByName("debug") { /* uses default debug keystore */ }
        create("release") {
            // Populated from environment or local.properties
        }
    }

    buildTypes {
        release {
            isMinifyEnabled = true
            isShrinkResources = true
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
            signingConfig = signingConfigs.getByName("release")
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    kotlinOptions { jvmTarget = "17" }
    buildFeatures { compose = true }
}

dependencies {
    implementation(platform(libs.compose.bom))
    implementation(libs.compose.material3)
    implementation(libs.compose.ui.tooling.preview)
    implementation(libs.compose.navigation)
    implementation(libs.core.ktx)
    implementation(libs.lifecycle.runtime.compose)
    implementation(libs.activity.compose)
    implementation(libs.hilt.android)
    ksp(libs.hilt.compiler)
    implementation(libs.hilt.navigation.compose)
    implementation(libs.orbit.core)
    implementation(libs.orbit.viewmodel)
    implementation(libs.orbit.compose)
    implementation(libs.lottie.compose)
    implementation(libs.coil.compose)
    implementation(libs.billing)
    debugImplementation(libs.compose.ui.tooling)
}
```

## gradle.properties

```properties
org.gradle.caching=true
org.gradle.configuration-cache=true
org.gradle.parallel=true
android.useAndroidX=true
```

## Tuist Integration

- Run `tuist init` in project root to enable.
- Build insights collected automatically on each `./gradlew` invocation.
- Remote cache shared across team -- no config beyond `tuist auth` required.
- Bundle analysis: `tuist inspect bundle app.aab` (requires release build first).

## CLAUDE.md Configuration

Add to project root `CLAUDE.md`:

```
## Build Commands
- Build: ./gradlew build
- Test: ./gradlew test
- Lint: ./gradlew lint
- Bundle analysis: tuist inspect bundle app.aab
```
