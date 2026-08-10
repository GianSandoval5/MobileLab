import 'package:app_prueba_mobilelab/core/providers/repository_providers.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/purchase.dart';
import 'package:app_prueba_mobilelab/features/store/domain/repositories/store_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class PurchasesState {
  const PurchasesState({
    this.purchases = const [],
    this.isLoading = false,
    this.error,
  });

  final List<Purchase> purchases;
  final bool isLoading;
  final String? error;
}

class PurchasesController extends StateNotifier<PurchasesState> {
  PurchasesController(this._repository) : super(const PurchasesState()) {
    load();
  }

  final StoreRepository _repository;

  Future<void> load() async {
    state = PurchasesState(purchases: state.purchases, isLoading: true);
    try {
      final purchases = await _repository.getPurchases();
      state = PurchasesState(purchases: purchases);
    } catch (error) {
      state = PurchasesState(
        purchases: state.purchases,
        error: error.toString(),
      );
    }
  }

  void add(Purchase purchase) {
    state = PurchasesState(purchases: [purchase, ...state.purchases]);
  }
}

final purchasesControllerProvider =
    StateNotifierProvider<PurchasesController, PurchasesState>((ref) {
      return PurchasesController(ref.watch(storeRepositoryProvider));
    });
