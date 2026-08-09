import 'package:flutter/widgets.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mobilelab_flutter/mobilelab_flutter.dart';

final class RecordingTransport implements MobileLabTransport {
  final events = <Map<String, Object?>>[];

  @override
  Future<void> send(Map<String, Object?> event) async {
    events.add(event);
  }
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  test('client emits the versioned Flutter contract', () async {
    final transport = RecordingTransport();
    final client = MobileLabClient(
      endpoint: Uri.parse('http://10.0.2.2:4566'),
      sessionId: 'run-1',
      transport: transport,
    );

    await client.marker('checkout.loaded', attributes: {'screen': 'checkout'});
    await client.assertThat('cart.total', true);

    expect(transport.events, hasLength(2));
    expect(transport.events.first, containsPair('protocolVersion', 1));
    expect(transport.events.first, containsPair('framework', 'flutter'));
    expect(transport.events.first, containsPair('sessionId', 'run-1'));
    expect(transport.events.last, containsPair('passed', true));
  });

  test('lifecycle reporter de-duplicates equivalent background states',
      () async {
    final transport = RecordingTransport();
    final reporter = MobileLabLifecycleReporter(MobileLabClient(
      endpoint: Uri.parse('http://127.0.0.1:4566'),
      transport: transport,
    ));

    reporter.didChangeAppLifecycleState(AppLifecycleState.inactive);
    reporter.didChangeAppLifecycleState(AppLifecycleState.paused);
    reporter.didChangeAppLifecycleState(AppLifecycleState.resumed);
    await Future<void>.delayed(Duration.zero);

    expect(transport.events.map((event) => event['name']),
        ['background', 'foreground']);
  });
}
