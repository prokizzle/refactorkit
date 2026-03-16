# StoreKit 2 Configuration Reference

## .storekit Configuration File

- Create in Xcode: File > New > StoreKit Configuration File.
- The skill populates it with subscription products.
- Example subscription group structure:

| Group | Product ID | Price |
|---|---|---|
| Premium | com.app.premium.monthly | $4.99/mo |
| Premium | com.app.premium.annual | $39.99/yr |
| Premium | com.app.premium.lifetime | $99.99 |

## StoreManager with Factory DI

```swift
import StoreKit
import Factory

@Observable
final class StoreManager {
    @ObservationIgnored
    @Injected(\.analyticsService) private var analytics

    private(set) var products: [Product] = []
    private(set) var purchasedProductIDs: Set<String> = []
    private(set) var isLoading: Bool = false

    private var transactionListener: Task<Void, Error>?

    private let productIDs = [
        "com.app.premium.monthly",
        "com.app.premium.annual",
        "com.app.premium.lifetime"
    ]

    init() {
        transactionListener = listenForTransactions()
    }

    deinit { transactionListener?.cancel() }

    var isPremium: Bool { !purchasedProductIDs.isEmpty }

    @MainActor
    func loadProducts() async {
        isLoading = true
        defer { isLoading = false }
        do {
            products = try await Product.products(for: productIDs)
                .sorted { $0.price < $1.price }
        } catch {
            products = []
        }
    }

    @MainActor
    func purchase(_ product: Product) async throws -> Bool {
        let result = try await product.purchase()
        switch result {
        case .success(let verification):
            let transaction = try checkVerified(verification)
            purchasedProductIDs.insert(transaction.productID)
            await transaction.finish()
            analytics.track(.purchaseCompleted(product.id))
            return true
        case .userCancelled, .pending:
            return false
        @unknown default:
            return false
        }
    }

    @MainActor
    func restorePurchases() async {
        try? await AppStore.sync()
        await refreshEntitlements()
    }

    @MainActor
    func refreshEntitlements() async {
        var active: Set<String> = []
        for await result in Transaction.currentEntitlements {
            if let transaction = try? checkVerified(result) {
                active.insert(transaction.productID)
            }
        }
        purchasedProductIDs = active
    }

    private func listenForTransactions() -> Task<Void, Error> {
        Task.detached { [weak self] in
            for await result in Transaction.updates {
                if let transaction = try? self?.checkVerified(result) {
                    await self?.refreshEntitlements()
                    await transaction.finish()
                }
            }
        }
    }

    private func checkVerified<T>(_ result: VerificationResult<T>) throws -> T {
        switch result {
        case .unverified: throw StoreError.verificationFailed
        case .verified(let value): return value
        }
    }
}

enum StoreError: LocalizedError {
    case verificationFailed
}
```

## Factory Registration

```swift
extension Container {
    var storeManager: Factory<StoreManager> {
        Factory(self) { StoreManager() }.singleton
    }
}
```

## Paywall View

```swift
struct PaywallView: View {
    let model: PaywallModel
    let intent: PaywallIntent

    var body: some View {
        VStack(spacing: Spacing.lg) {
            LottieView(animation: .named("premium-badge"))
                .playing(loopMode: .loop)
                .frame(width: 120, height: 120)

            Text("Unlock Premium")
                .font(Typography.title)
                .foregroundStyle(Colors.textPrimary)

            ForEach(model.products) { product in
                DSCard {
                    HStack {
                        VStack(alignment: .leading, spacing: Spacing.xs) {
                            Text(product.displayName).font(Typography.body)
                            Text(product.displayPrice).font(Typography.caption)
                        }
                        Spacer()
                        DSButton(
                            model.purchasingProductID == product.id ? "..." : "Subscribe",
                            style: .primary
                        ) {
                            intent.send(.purchase(product))
                        }
                        .disabled(model.purchasingProductID != nil)
                    }
                }
            }

            Button("Restore Purchases") { intent.send(.restore) }
                .font(Typography.caption)
        }
        .padding(Spacing.lg)
        .changeEffect(.spray { Text("!").font(.title) }, value: model.showCelebration)
        .task { intent.send(.loadProducts) }
    }
}
```

## Tuist Integration

- Add `.storekit` file to project resources in `Project.swift`:
  ```swift
  resources: ["Resources/**", "Products.storekit"]
  ```
- Add StoreKit capability in target entitlements.
- Set the `.storekit` file in the scheme's Run > Options > StoreKit Configuration for testing.

## Testing

- Enable StoreKit Testing in Xcode via the scheme's StoreKit Configuration option.
- `Transaction.updates` listener must be started early (app launch) to catch background renewals.
- Sandbox accounts: create in App Store Connect under Users and Access > Sandbox.
- Production: `AppStore.sync()` replaces the legacy `SKPaymentQueue.restoreCompletedTransactions()`.
- Clear StoreKit test data between runs: Editor > Reset StoreKit Transactions.
