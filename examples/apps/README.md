# Runnable mobile application examples

These applications demonstrate that MobileLab's API sandbox is framework-neutral. Both use the same local e-commerce API contract and equivalent fixtures while keeping application state in their framework's normal client-side architecture.

| Example | Stack | Application directory |
| --- | --- | --- |
| [Flutter shop](flutter/README.md) | Flutter, Dart, Riverpod | `examples/apps/flutter` |
| [React Native shop](react-native/README.md) | React Native CLI, TypeScript, Zustand | `examples/apps/react-native` |

Each example includes its own `mobilelab.yaml`, JSON fixtures, and a profile scenario. Run only one example sandbox at a time because both intentionally use port `4566`.

From either application directory:

```sh
../../../bin/mobilelab doctor
../../../bin/mobilelab start
```

Then open:

- API information: `http://127.0.0.1:4566/`
- Example endpoint: `http://127.0.0.1:4566/api/products`
- Dashboard: `http://127.0.0.1:4566/dashboard`

Android Emulator applications use `http://10.0.2.2:4566` to reach the host. iOS Simulator, web, and desktop builds normally use `http://127.0.0.1:4566`. Read the application-specific guide for dependency installation, launch commands, physical-device configuration, and tests.

Generated dependencies, platform builds, IDE metadata, local MobileLab state, and databases are intentionally excluded. Restore them with the standard Flutter, npm, CocoaPods, and Gradle commands instead of committing them.
