# @mobilelab/react-native

Optional advanced observability hooks for MobileLab. The API Sandbox does not require this package.

```ts
import { AppState } from 'react-native';
import { MobileLabClient, MobileLabLifecycleReporter } from '@mobilelab/react-native';

const mobileLab = new MobileLabClient({
  endpoint: 'http://10.0.2.2:4566',
  sessionId: 'checkout-test',
});

const lifecycle = new MobileLabLifecycleReporter(mobileLab, AppState);
lifecycle.attach();

await mobileLab.marker('checkout.loaded');
await mobileLab.assertThat('cart.total', total === expected);
```

Call `lifecycle.detach()` when reporting is no longer needed. Explicit calls reject on transport errors; automatic lifecycle reporting routes failures to the optional `onError` callback and never throws into the application lifecycle.

Use `127.0.0.1` for iOS Simulator and `10.0.2.2` for Android Emulator. Development builds may need platform-local network or cleartext HTTP configuration. The package uses the React Native global `fetch` implementation and has no runtime dependency on the MobileLab CLI package.
