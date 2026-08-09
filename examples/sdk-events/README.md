# SDK event example environment

This one MobileLab environment is shared by the Android, iOS, and Capacitor SDK examples. It demonstrates that framework packages are optional clients of the same Core protocol rather than separate backends.

Start MobileLab from this directory, run the matching scenario with the fake device adapter, and trigger `example.clicked` in the sample application after the run begins:

```sh
mobilelab start
mobilelab scenario run android-marker --platform fake
mobilelab scenario run ios-marker --platform fake
mobilelab scenario run capacitor-marker --platform fake
```

Emulator and Simulator endpoint rules are documented in each SDK README. The scenarios intentionally differ only by the `framework` filter.
