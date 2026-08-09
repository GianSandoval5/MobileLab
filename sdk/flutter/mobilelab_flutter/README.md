# mobilelab_flutter

Optional advanced observability hooks for MobileLab. The API Sandbox does not require this package.

```dart
final mobileLab = MobileLabClient(
  endpoint: Uri.parse('http://10.0.2.2:4566'),
  sessionId: 'checkout-test',
);

final lifecycle = MobileLabLifecycleReporter(mobileLab)..attach();
await mobileLab.marker('checkout.loaded');
await mobileLab.assertThat('cart.total', total == expected);
```

Call `lifecycle.detach()` when the application no longer needs reporting. SDK failures are returned by explicit calls; automatic lifecycle reporting never crashes the application and can be observed with `onError`.

Use `127.0.0.1` for iOS Simulator and `10.0.2.2` for Android Emulator. Android development builds may need a debug-only cleartext HTTP exception. This package targets Flutter mobile and uses `dart:io`; it is not a Flutter Web transport.
