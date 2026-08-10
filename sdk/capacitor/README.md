# @mobilelab/capacitor

Optional Capacitor hooks using MobileLab protocol v1. The API Sandbox works without this package; add it only for lifecycle events, markers, and scenario assertions.

Install the v1.1 GitHub Release asset locally with `npm install ./mobilelab-capacitor-1.1.0.tgz`.

The package takes the Capacitor App plugin through a structural interface, so it does not bundle or impose a second `@capacitor/app` version:

```ts
import { App } from '@capacitor/app';
import { MobileLabClient, MobileLabLifecycleReporter } from '@mobilelab/capacitor';

const client = new MobileLabClient({ endpoint: 'http://10.0.2.2:4566' });
const lifecycle = new MobileLabLifecycleReporter(client, App);
await lifecycle.attach();

await client.marker('checkout.loaded', { screen: 'checkout' });
await client.assertThat('cart.total', cart.total === expectedTotal);
```

Use `10.0.2.2` from the standard Android Emulator and `127.0.0.1` from iOS Simulator. Physical devices need a reachable host address; exposing MobileLab beyond loopback is explicit and carries a security warning. Native clear-text transport policies may require development-only configuration.

Node 18 or newer is needed only to build/test this package; Node is not required by the MobileLab binary.
