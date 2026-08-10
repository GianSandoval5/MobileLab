import 'package:app_prueba_mobilelab/features/store/presentation/pages/cart_page.dart';
import 'package:app_prueba_mobilelab/features/store/presentation/providers/cart_provider.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class CartAction extends ConsumerWidget {
  const CartAction({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final count = ref.watch(cartControllerProvider).itemCount;

    return Padding(
      padding: const EdgeInsets.only(right: 8),
      child: Badge(
        isLabelVisible: count > 0,
        label: Text('$count'),
        child: IconButton(
          tooltip: 'Carrito',
          onPressed: () => Navigator.of(
            context,
          ).push(MaterialPageRoute(builder: (_) => const CartPage())),
          icon: const Icon(Icons.shopping_bag_outlined),
        ),
      ),
    );
  }
}
