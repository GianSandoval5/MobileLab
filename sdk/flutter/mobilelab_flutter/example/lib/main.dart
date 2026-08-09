import 'package:flutter/material.dart';
import 'package:mobilelab_flutter/mobilelab_flutter.dart';

void main() {
  final client = MobileLabClient(endpoint: Uri.parse('http://10.0.2.2:4566'));
  MobileLabLifecycleReporter(client).attach();
  runApp(ExampleApp(client: client));
}

class ExampleApp extends StatelessWidget {
  const ExampleApp({required this.client, super.key});

  final MobileLabClient client;

  @override
  Widget build(BuildContext context) => MaterialApp(
        home: Scaffold(
          body: Center(
            child: FilledButton(
              onPressed: () => client.marker('example.button_pressed'),
              child: const Text('Report MobileLab marker'),
            ),
          ),
        ),
      );
}
