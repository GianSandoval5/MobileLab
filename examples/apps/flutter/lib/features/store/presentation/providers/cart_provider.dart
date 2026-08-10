import 'package:app_prueba_mobilelab/core/providers/repository_providers.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/cart_item.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/product.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/purchase.dart';
import 'package:app_prueba_mobilelab/features/store/domain/repositories/store_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class CartState {
  const CartState({this.items = const [], this.isBusy = false, this.error});

  final List<CartItem> items;
  final bool isBusy;
  final String? error;

  int get itemCount => items.fold(0, (sum, item) => sum + item.quantity);
  double get total => items.fold(0, (sum, item) => sum + item.subtotal);
}

class CartController extends StateNotifier<CartState> {
  CartController(this._repository) : super(const CartState());

  final StoreRepository _repository;

  Future<bool> add(Product product) async {
    final index = state.items.indexWhere(
      (item) => item.product.id == product.id,
    );
    final currentQuantity = index == -1 ? 0 : state.items[index].quantity;
    final newQuantity = currentQuantity + 1;
    state = CartState(items: state.items, isBusy: true);

    try {
      if (index == -1) {
        await _repository.addCartItem(product, newQuantity);
        state = CartState(
          items: [
            ...state.items,
            CartItem(product: product, quantity: 1),
          ],
        );
      } else {
        await _repository.updateCartItem(product, newQuantity);
        final updated = [...state.items];
        updated[index] = updated[index].copyWith(quantity: newQuantity);
        state = CartState(items: updated);
      }
      return true;
    } catch (error) {
      state = CartState(items: state.items, error: error.toString());
      return false;
    }
  }

  Future<void> changeQuantity(CartItem item, int quantity) async {
    if (quantity <= 0) {
      await remove(item.product.id);
      return;
    }
    state = CartState(items: state.items, isBusy: true);
    try {
      await _repository.updateCartItem(item.product, quantity);
      state = CartState(
        items: state.items
            .map(
              (current) => current.product.id == item.product.id
                  ? current.copyWith(quantity: quantity)
                  : current,
            )
            .toList(),
      );
    } catch (error) {
      state = CartState(items: state.items, error: error.toString());
    }
  }

  Future<void> remove(String productId) async {
    state = CartState(items: state.items, isBusy: true);
    try {
      await _repository.removeCartItem(productId);
      state = CartState(
        items: state.items
            .where((item) => item.product.id != productId)
            .toList(),
      );
    } catch (error) {
      state = CartState(items: state.items, error: error.toString());
    }
  }

  Future<Purchase?> checkout() async {
    if (state.items.isEmpty) return null;
    final items = [...state.items];
    final total = state.total;
    state = CartState(items: items, isBusy: true);
    try {
      final result = await _repository.pay(items, total);
      final purchase = Purchase(
        id: result.purchaseId,
        createdAt: DateTime.now(),
        status: 'Pagado',
        total: total,
        items: items,
      );
      // Conserva el resumen mientras la UI muestra la confirmación de pago.
      state = CartState(items: items);
      return purchase;
    } catch (error) {
      state = CartState(items: items, error: error.toString());
      return null;
    }
  }

  void completeCheckout() => state = const CartState();
}

final cartControllerProvider = StateNotifierProvider<CartController, CartState>(
  (ref) {
    return CartController(ref.watch(storeRepositoryProvider));
  },
);
