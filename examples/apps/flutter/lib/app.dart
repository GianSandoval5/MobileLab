import 'package:app_prueba_mobilelab/core/theme/app_theme.dart';
import 'package:app_prueba_mobilelab/features/auth/presentation/pages/auth_page.dart';
import 'package:app_prueba_mobilelab/features/auth/presentation/providers/auth_provider.dart';
import 'package:app_prueba_mobilelab/features/store/presentation/pages/store_shell.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class MobileLabShopApp extends ConsumerWidget {
  const MobileLabShopApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final auth = ref.watch(authControllerProvider);

    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'MobileLab Shop',
      theme: AppTheme.light,
      home: auth.user == null ? const AuthPage() : const StoreShell(),
    );
  }
}
