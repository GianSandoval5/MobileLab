import 'package:app_prueba_mobilelab/features/store/domain/entities/business.dart';
import 'package:app_prueba_mobilelab/features/store/presentation/providers/catalog_providers.dart';
import 'package:app_prueba_mobilelab/features/store/presentation/widgets/async_error_view.dart';
import 'package:app_prueba_mobilelab/features/store/presentation/widgets/cart_action.dart';
import 'package:app_prueba_mobilelab/features/store/presentation/widgets/product_grid.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class BusinessDetailPage extends ConsumerWidget {
  const BusinessDetailPage({super.key, required this.business});

  final Business business;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final products = ref.watch(businessProductsProvider(business.id));
    return Scaffold(
      appBar: AppBar(title: Text(business.name), actions: const [CartAction()]),
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 4, 16, 16),
            child: Card(
              color: Theme.of(context).colorScheme.secondaryContainer,
              child: Padding(
                padding: const EdgeInsets.all(18),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      business.description,
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                    const SizedBox(height: 8),
                    Text('⭐ ${business.rating} · ${business.category}'),
                  ],
                ),
              ),
            ),
          ),
          Expanded(
            child: products.when(
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (error, _) => AsyncErrorView(
                error: error,
                onRetry: () =>
                    ref.invalidate(businessProductsProvider(business.id)),
              ),
              data: (items) => ProductGrid(products: items),
            ),
          ),
        ],
      ),
    );
  }
}
