import 'package:app_prueba_mobilelab/features/store/domain/entities/product.dart';
import 'package:app_prueba_mobilelab/features/store/presentation/pages/product_form_page.dart';
import 'package:app_prueba_mobilelab/features/store/presentation/providers/my_products_provider.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class MyProductsPage extends ConsumerWidget {
  const MyProductsPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(myProductsControllerProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Mis productos')),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _openForm(context),
        icon: const Icon(Icons.add),
        label: const Text('Publicar'),
      ),
      body: _buildBody(context, ref, state),
    );
  }

  Widget _buildBody(
    BuildContext context,
    WidgetRef ref,
    MyProductsState state,
  ) {
    if (state.isLoading && state.products.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }
    if (state.error != null && state.products.isEmpty) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(state.error!, textAlign: TextAlign.center),
            OutlinedButton(
              onPressed: () =>
                  ref.read(myProductsControllerProvider.notifier).load(),
              child: const Text('Reintentar'),
            ),
          ],
        ),
      );
    }
    return RefreshIndicator(
      onRefresh: () => ref.read(myProductsControllerProvider.notifier).load(),
      child: ListView.separated(
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 100),
        itemCount: state.products.length,
        separatorBuilder: (_, _) => const SizedBox(height: 10),
        itemBuilder: (context, index) => _ProductTile(
          product: state.products[index],
          onEdit: () => _openForm(context, state.products[index]),
        ),
      ),
    );
  }

  void _openForm(BuildContext context, [Product? product]) {
    Navigator.of(context).push(
      MaterialPageRoute(builder: (_) => ProductFormPage(product: product)),
    );
  }
}

class _ProductTile extends StatelessWidget {
  const _ProductTile({required this.product, required this.onEdit});

  final Product product;
  final VoidCallback onEdit;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: ListTile(
        contentPadding: const EdgeInsets.all(14),
        leading: CircleAvatar(
          child: Text(product.name.characters.first.toUpperCase()),
        ),
        title: Text(
          product.name,
          style: const TextStyle(fontWeight: FontWeight.w700),
        ),
        subtitle: Text(
          'S/ ${product.price.toStringAsFixed(2)} · Stock ${product.stock}',
        ),
        trailing: IconButton(
          tooltip: 'Editar',
          onPressed: onEdit,
          icon: const Icon(Icons.edit_outlined),
        ),
      ),
    );
  }
}
