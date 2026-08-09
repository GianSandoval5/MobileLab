import 'dart:async';
import 'dart:convert';
import 'dart:io';

const int mobileLabProtocolVersion = 1;

abstract interface class MobileLabTransport {
  Future<void> send(Map<String, Object?> event);
}

final class MobileLabHttpTransport implements MobileLabTransport {
  MobileLabHttpTransport(this.baseUri,
      {this.timeout = const Duration(seconds: 3)});

  final Uri baseUri;
  final Duration timeout;

  Uri get eventsUri => baseUri.resolve('/__mobilelab/sdk/events');

  @override
  Future<void> send(Map<String, Object?> event) async {
    final client = HttpClient()..connectionTimeout = timeout;
    try {
      final request = await client.postUrl(eventsUri).timeout(timeout);
      request.headers.contentType = ContentType.json;
      request.headers.set('X-MobileLab-SDK', 'mobilelab_flutter/0.3.0');
      request.write(jsonEncode(event));
      final response = await request.close().timeout(timeout);
      await response.drain<void>();
      if (response.statusCode < 200 || response.statusCode > 299) {
        throw MobileLabException(
            'SDK bridge returned HTTP ${response.statusCode}');
      }
    } on TimeoutException catch (error) {
      throw MobileLabException('SDK bridge timed out', error);
    } on SocketException catch (error) {
      throw MobileLabException(
          'SDK bridge is unreachable at $eventsUri', error);
    } finally {
      client.close(force: true);
    }
  }
}

final class MobileLabClient {
  MobileLabClient({
    required Uri endpoint,
    this.sessionId,
    MobileLabTransport? transport,
  }) : _transport = transport ?? MobileLabHttpTransport(endpoint);

  final String? sessionId;
  final MobileLabTransport _transport;

  Future<void> lifecycle(
    String name, {
    Map<String, Object?> attributes = const {},
  }) =>
      _send('lifecycle', name, attributes: attributes);

  Future<void> marker(
    String name, {
    Map<String, Object?> attributes = const {},
  }) =>
      _send('marker', name, attributes: attributes);

  Future<void> assertThat(
    String name,
    bool passed, {
    Map<String, Object?> attributes = const {},
  }) =>
      _send('assertion', name, passed: passed, attributes: attributes);

  Future<void> _send(
    String kind,
    String name, {
    bool? passed,
    Map<String, Object?> attributes = const {},
  }) {
    final event = <String, Object?>{
      'protocolVersion': mobileLabProtocolVersion,
      'framework': 'flutter',
      'kind': kind,
      'name': name,
      if (passed != null) 'passed': passed,
      if (sessionId != null && sessionId!.isNotEmpty) 'sessionId': sessionId,
      if (attributes.isNotEmpty) 'attributes': Map.of(attributes),
    };
    return _transport.send(event);
  }
}

final class MobileLabException implements Exception {
  const MobileLabException(this.message, [this.cause]);

  final String message;
  final Object? cause;

  @override
  String toString() => cause == null ? message : '$message: $cause';
}
