import 'dart:convert';
import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:mobilelab_flutter/mobilelab_flutter.dart';

void main() {
  test('HTTP transport posts JSON to the SDK bridge path', () async {
    final server = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
    addTearDown(() => server.close(force: true));
    final received = <String, Object?>{};
    server.listen((request) async {
      received.addAll(
        (jsonDecode(await utf8.decoder.bind(request).join()) as Map)
            .cast<String, Object?>(),
      );
      expect(request.uri.path, '/__mobilelab/sdk/events');
      expect(
        request.headers.value('X-MobileLab-SDK'),
        'mobilelab_flutter/0.7.0',
      );
      request.response.statusCode = HttpStatus.accepted;
      await request.response.close();
    });

    final client = MobileLabClient(
      endpoint: Uri.parse('http://127.0.0.1:${server.port}/nested/path'),
    );
    await client.marker('transport.ready');

    expect(received, containsPair('name', 'transport.ready'));
  });
}
