import 'package:app_prueba_mobilelab/features/store/presentation/providers/cart_provider.dart';
import 'package:app_prueba_mobilelab/features/store/presentation/providers/purchases_provider.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class CheckoutPage extends ConsumerStatefulWidget {
  const CheckoutPage({super.key});

  @override
  ConsumerState<CheckoutPage> createState() => _CheckoutPageState();
}

class _CheckoutPageState extends ConsumerState<CheckoutPage> {
  final _formKey = GlobalKey<FormState>();
  final _nameController = TextEditingController(text: 'MobileLab User');
  final _cardController = TextEditingController(text: '4242424242424242');
  final _expiryController = TextEditingController(text: '12/30');
  final _cvvController = TextEditingController(text: '123');

  @override
  void dispose() {
    _nameController.dispose();
    _cardController.dispose();
    _expiryController.dispose();
    _cvvController.dispose();
    super.dispose();
  }

  Future<void> _pay() async {
    if (!_formKey.currentState!.validate()) return;
    FocusScope.of(context).unfocus();
    final purchase = await ref.read(cartControllerProvider.notifier).checkout();
    if (!mounted) return;

    if (purchase == null) {
      final error = ref.read(cartControllerProvider).error;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(error ?? 'No se pudo procesar el pago.')),
      );
      return;
    }

    ref.read(purchasesControllerProvider.notifier).add(purchase);
    await showDialog<void>(
      context: context,
      barrierDismissible: false,
      builder: (dialogContext) => AlertDialog(
        icon: const Icon(Icons.check_circle, size: 52, color: Colors.green),
        title: const Text('¡Pago completado!'),
        content: Text(
          'La compra ${purchase.id} fue registrada correctamente.',
          textAlign: TextAlign.center,
        ),
        actionsAlignment: MainAxisAlignment.center,
        actions: [
          FilledButton(
            onPressed: () => Navigator.pop(dialogContext),
            child: const Text('Continuar'),
          ),
        ],
      ),
    );
    if (mounted) {
      ref.read(cartControllerProvider.notifier).completeCheckout();
      Navigator.pop(context, true);
    }
  }

  @override
  Widget build(BuildContext context) {
    final cart = ref.watch(cartControllerProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Pagar')),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(20),
          children: [
            Card(
              color: Theme.of(context).colorScheme.primaryContainer,
              child: Padding(
                padding: const EdgeInsets.all(18),
                child: Row(
                  children: [
                    const Icon(Icons.shopping_bag_outlined),
                    const SizedBox(width: 12),
                    Expanded(child: Text('${cart.itemCount} artículos')),
                    Text(
                      'S/ ${cart.total.toStringAsFixed(2)}',
                      style: const TextStyle(fontWeight: FontWeight.w800),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 24),
            Text(
              'Datos de la tarjeta',
              style: Theme.of(
                context,
              ).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w700),
            ),
            const SizedBox(height: 14),
            TextFormField(
              controller: _nameController,
              decoration: const InputDecoration(
                labelText: 'Nombre del titular',
                prefixIcon: Icon(Icons.person_outline),
              ),
              validator: _required,
            ),
            const SizedBox(height: 14),
            TextFormField(
              controller: _cardController,
              keyboardType: TextInputType.number,
              inputFormatters: [FilteringTextInputFormatter.digitsOnly],
              decoration: const InputDecoration(
                labelText: 'Número de tarjeta',
                prefixIcon: Icon(Icons.credit_card),
              ),
              validator: (value) =>
                  (value?.length ?? 0) < 16 ? 'Ingresa 16 dígitos.' : null,
            ),
            const SizedBox(height: 14),
            Row(
              children: [
                Expanded(
                  child: TextFormField(
                    controller: _expiryController,
                    keyboardType: TextInputType.datetime,
                    decoration: const InputDecoration(labelText: 'MM/AA'),
                    validator: _required,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: TextFormField(
                    controller: _cvvController,
                    keyboardType: TextInputType.number,
                    obscureText: true,
                    inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                    decoration: const InputDecoration(labelText: 'CVV'),
                    validator: (value) =>
                        (value?.length ?? 0) < 3 ? 'CVV inválido.' : null,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 10),
            const Text(
              'Pago simulado: MobileLab no procesa datos bancarios reales.',
              style: TextStyle(fontSize: 12),
            ),
            const SizedBox(height: 28),
            FilledButton.icon(
              onPressed: cart.isBusy ? null : _pay,
              icon: cart.isBusy
                  ? const SizedBox.square(
                      dimension: 18,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.lock_outline),
              label: Padding(
                padding: const EdgeInsets.symmetric(vertical: 13),
                child: Text(
                  cart.isBusy
                      ? 'Procesando…'
                      : 'Pagar S/ ${cart.total.toStringAsFixed(2)}',
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  String? _required(String? value) {
    return value == null || value.trim().isEmpty ? 'Campo requerido.' : null;
  }
}
