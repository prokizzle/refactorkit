# Google Play Billing Library 7+ Reference

## Setup

Add dependency via version catalog:

```toml
# libs.versions.toml
billing = "7.1.1"

[libraries]
billing-ktx = { group = "com.android.billingclient", name = "billing-ktx", version.ref = "billing" }
```

AndroidManifest.xml:

```xml
<uses-permission android:name="com.android.vending.BILLING" />
```

## BillingManager with Hilt

```kotlin
@Singleton
class BillingManager @Inject constructor(
    @ApplicationContext private val context: Context
) {
    private var billingClient: BillingClient? = null
    private val _purchases = MutableStateFlow<List<Purchase>>(emptyList())
    val purchases: StateFlow<List<Purchase>> = _purchases

    private val purchasesUpdatedListener = PurchasesUpdatedListener { billingResult, purchases ->
        if (billingResult.responseCode == BillingClient.BillingResponseCode.OK && purchases != null) {
            _purchases.value = purchases
            purchases.forEach { if (!it.isAcknowledged) acknowledgePurchase(it) }
        }
    }

    fun startConnection(onConnected: () -> Unit = {}) {
        billingClient = BillingClient.newBuilder(context)
            .setListener(purchasesUpdatedListener)
            .enablePendingPurchases(PendingPurchasesParams.newBuilder().enableOneTimeProducts().build())
            .build()

        billingClient?.startConnection(object : BillingClientStateListener {
            override fun onBillingSetupFinished(result: BillingResult) {
                if (result.responseCode == BillingClient.BillingResponseCode.OK) onConnected()
            }
            override fun onBillingServiceDisconnected() {
                // Retry with exponential backoff in production
                startConnection(onConnected)
            }
        })
    }

    suspend fun queryProductDetails(): List<ProductDetails> {
        val params = QueryProductDetailsParams.newBuilder()
            .setProductList(subscriptionProducts)
            .build()
        val result = billingClient?.queryProductDetails(params)
        return result?.productDetailsList.orEmpty()
    }

    fun launchBillingFlow(activity: Activity, productDetails: ProductDetails, offerToken: String) {
        val flowParams = BillingFlowParams.newBuilder()
            .setProductDetailsParamsList(listOf(
                BillingFlowParams.ProductDetailsParams.newBuilder()
                    .setProductDetails(productDetails)
                    .setOfferToken(offerToken)
                    .build()
            )).build()
        billingClient?.launchBillingFlow(activity, flowParams)
    }

    suspend fun queryPurchases(): List<Purchase> {
        val params = QueryPurchasesParams.newBuilder()
            .setProductType(BillingClient.ProductType.SUBS)
            .build()
        val result = billingClient?.queryPurchasesAsync(params)
        _purchases.value = result?.purchasesList.orEmpty()
        return _purchases.value
    }

    private fun acknowledgePurchase(purchase: Purchase) {
        val params = AcknowledgePurchaseParams.newBuilder()
            .setPurchaseToken(purchase.purchaseToken)
            .build()
        billingClient?.acknowledgePurchase(params) { /* handle result */ }
    }

    fun endConnection() { billingClient?.endConnection() }
}
```

## Product Configuration

```kotlin
val subscriptionProducts = listOf(
    QueryProductDetailsParams.Product.newBuilder()
        .setProductId("premium_monthly")
        .setProductType(BillingClient.ProductType.SUBS)
        .build(),
    QueryProductDetailsParams.Product.newBuilder()
        .setProductId("premium_annual")
        .setProductType(BillingClient.ProductType.SUBS)
        .build(),
    QueryProductDetailsParams.Product.newBuilder()
        .setProductId("premium_lifetime")
        .setProductType(BillingClient.ProductType.INAPP)
        .build(),
)
```

## Hilt Module

```kotlin
@Module
@InstallIn(SingletonComponent::class)
object BillingModule {
    @Provides @Singleton
    fun provideBillingManager(@ApplicationContext context: Context): BillingManager {
        return BillingManager(context)
    }
}
```

## Paywall Composable

```kotlin
@Composable
fun PaywallScreen(viewModel: PaywallViewModel = hiltViewModel()) {
    val plans by viewModel.plans.collectAsStateWithLifecycle()
    val purchaseState by viewModel.purchaseState.collectAsStateWithLifecycle()
    val context = LocalContext.current

    Column(modifier = Modifier.fillMaxSize().padding(24.dp)) {
        // Lottie premium badge
        val composition by rememberLottieComposition(LottieCompositionSpec.RawRes(R.raw.premium_badge))
        LottieAnimation(composition, iterations = LottieConstants.IterateForever, modifier = Modifier.size(120.dp).align(CenterHorizontally))

        Text("Unlock Premium", style = MaterialTheme.typography.headlineMedium)
        Spacer(Modifier.height(16.dp))

        plans.forEach { plan ->
            Card(
                modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp),
                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant),
                onClick = { viewModel.purchase(context as Activity, plan) }
            ) {
                Row(Modifier.padding(16.dp), verticalAlignment = Alignment.CenterVertically) {
                    Column(Modifier.weight(1f)) {
                        Text(plan.title, style = MaterialTheme.typography.titleSmall)
                        Text(plan.formattedPrice, style = MaterialTheme.typography.bodyMedium)
                    }
                    if (plan.isBestValue) Badge { Text("Best Value") }
                }
            }
        }

        when (purchaseState) {
            is PurchaseState.Loading -> CircularProgressIndicator(Modifier.align(CenterHorizontally))
            is PurchaseState.Success -> { /* Show confetti via accompanist or custom Canvas animation */ }
            is PurchaseState.Error -> Text("Purchase failed. Try again.", color = MaterialTheme.colorScheme.error)
            else -> {}
        }
    }
}
```

## Testing

- **Test accounts**: Add Gmail addresses under Play Console > Setup > License testing. These accounts can make test purchases without charges.
- **Test tracks**: Use internal testing track for rapid iteration. Builds are available within minutes.
- **BillingClient test mode**: Test purchases return `PURCHASED` state immediately. Use `isAcknowledged` checks to verify acknowledgment flow.
- **Static responses**: Use reserved product IDs (`android.test.purchased`, `android.test.canceled`) for unit-level billing response testing.
- **Verify server-side**: Always validate purchase tokens server-side via Google Play Developer API for production.
