import 'dart:async';

import 'package:flutter/widgets.dart';

import 'client.dart';

typedef MobileLabErrorHandler = void Function(
    Object error, StackTrace stackTrace);

final class MobileLabLifecycleReporter with WidgetsBindingObserver {
  MobileLabLifecycleReporter(this.client, {this.onError});

  final MobileLabClient client;
  final MobileLabErrorHandler? onError;
  bool _attached = false;
  String? _lastLifecycle;

  void attach({bool reportReady = true}) {
    if (_attached) return;
    WidgetsBinding.instance.addObserver(this);
    _attached = true;
    if (reportReady) _report(client.lifecycle('ready'));
  }

  void detach() {
    if (!_attached) return;
    WidgetsBinding.instance.removeObserver(this);
    _attached = false;
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    final name = switch (state) {
      AppLifecycleState.resumed => 'foreground',
      AppLifecycleState.detached => 'detached',
      AppLifecycleState.inactive ||
      AppLifecycleState.hidden ||
      AppLifecycleState.paused =>
        'background',
    };
    if (name == _lastLifecycle) return;
    _lastLifecycle = name;
    _report(client.lifecycle(name, attributes: {'flutterState': state.name}));
  }

  void _report(Future<void> operation) {
    unawaited(operation.catchError((Object error, StackTrace stackTrace) {
      onError?.call(error, stackTrace);
    }));
  }
}
