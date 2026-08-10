import 'package:app_prueba_mobilelab/features/store/presentation/pages/businesses_page.dart';
import 'package:app_prueba_mobilelab/features/store/presentation/pages/home_page.dart';
import 'package:app_prueba_mobilelab/features/store/presentation/pages/profile_page.dart';
import 'package:app_prueba_mobilelab/features/store/presentation/pages/purchases_page.dart';
import 'package:app_prueba_mobilelab/features/store/presentation/widgets/cart_action.dart';
import 'package:flutter/material.dart';

class StoreShell extends StatefulWidget {
  const StoreShell({super.key});

  @override
  State<StoreShell> createState() => _StoreShellState();
}

class _StoreShellState extends State<StoreShell> {
  var _index = 0;

  static const _titles = ['Descubre', 'Negocios', 'Mis compras', 'Mi perfil'];
  static const _pages = [
    HomePage(),
    BusinessesPage(),
    PurchasesPage(),
    ProfilePage(),
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(
          _titles[_index],
          style: const TextStyle(fontWeight: FontWeight.w700),
        ),
        actions: const [CartAction()],
      ),
      body: IndexedStack(index: _index, children: _pages),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _index,
        onDestinationSelected: (value) => setState(() => _index = value),
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.home_outlined),
            selectedIcon: Icon(Icons.home_rounded),
            label: 'Inicio',
          ),
          NavigationDestination(
            icon: Icon(Icons.store_outlined),
            selectedIcon: Icon(Icons.store_rounded),
            label: 'Negocios',
          ),
          NavigationDestination(
            icon: Icon(Icons.receipt_long_outlined),
            selectedIcon: Icon(Icons.receipt_long_rounded),
            label: 'Compras',
          ),
          NavigationDestination(
            icon: Icon(Icons.person_outline),
            selectedIcon: Icon(Icons.person_rounded),
            label: 'Perfil',
          ),
        ],
      ),
    );
  }
}
