import 'package:app_prueba_mobilelab/features/store/domain/entities/purchase.dart';
import 'package:app_prueba_mobilelab/features/store/presentation/providers/purchases_provider.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class PurchasesPage extends ConsumerWidget {
  const PurchasesPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(purchasesControllerProvider);

    if (state.isLoading && state.purchases.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }
    if (state.error != null && state.purchases.isEmpty) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(state.error!, textAlign: TextAlign.center),
              const SizedBox(height: 12),
              OutlinedButton(
                onPressed: () =>
                    ref.read(purchasesControllerProvider.notifier).load(),
                child: const Text('Reintentar'),
              ),
            ],
          ),
        ),
      );
    }
    if (state.purchases.isEmpty) {
      return const Center(child: Text('Todavía no tienes compras.'));
    }

    return RefreshIndicator(
      onRefresh: () => ref.read(purchasesControllerProvider.notifier).load(),
      child: ListView.separated(
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
        itemCount: state.purchases.length,
        separatorBuilder: (_, _) => const SizedBox(height: 12),
        itemBuilder: (context, index) =>
            _PurchaseCard(purchase: state.purchases[index]),
      ),
    );
  }
}

class _PurchaseCard extends StatelessWidget {
  const _PurchaseCard({required this.purchase});

  final Purchase purchase;

  @override
  Widget build(BuildContext context) {
    final date = purchase.createdAt.toLocal();
    final formattedDate =
        '${date.day.toString().padLeft(2, '0')}/'
        '${date.month.toString().padLeft(2, '0')}/${date.year}';
    return Card(
      child: ExpansionTile(
        shape: const Border(),
        leading: CircleAvatar(
          child: Icon(
            purchase.status == 'Entregado'
                ? Icons.inventory_2_outlined
                : Icons.schedule,
          ),
        ),
        title: Text(
          'Compra ${purchase.id}',
          style: const TextStyle(fontWeight: FontWeight.w700),
        ),
        subtitle: Text('$formattedDate · ${purchase.status}'),
        trailing: Text(
          'S/ ${purchase.total.toStringAsFixed(2)}',
          style: const TextStyle(fontWeight: FontWeight.w800),
        ),
        children: purchase.items
            .map(
              (item) => ListTile(
                title: Text(item.product.name),
                subtitle: Text('${item.quantity} unidad(es)'),
                trailing: Text('S/ ${item.subtotal.toStringAsFixed(2)}'),
              ),
            )
            .toList(),
      ),
    );
  }
}
