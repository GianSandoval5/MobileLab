import 'package:app_prueba_mobilelab/core/providers/repository_providers.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/product.dart';
import 'package:app_prueba_mobilelab/features/store/domain/repositories/store_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class MyProductsState {
  const MyProductsState({
    this.products = const [],
    this.isLoading = false,
    this.error,
  });

  final List<Product> products;
  final bool isLoading;
  final String? error;
}

class MyProductsController extends StateNotifier<MyProductsState> {
  MyProductsController(this._repository) : super(const MyProductsState()) {
    load();
  }

  final StoreRepository _repository;

  Future<void> load() async {
    state = MyProductsState(products: state.products, isLoading: true);
    try {
      state = MyProductsState(products: await _repository.getMyProducts());
    } catch (error) {
      state = MyProductsState(
        products: state.products,
        error: error.toString(),
      );
    }
  }

  Future<bool> save(Product product, {required bool isNew}) async {
    state = MyProductsState(products: state.products, isLoading: true);
    try {
      final saved = isNew
          ? await _repository.createProduct(product)
          : await _repository.updateProduct(product);
      final products = isNew
          ? [saved, ...state.products]
          : state.products
                .map((item) => item.id == saved.id ? saved : item)
                .toList();
      state = MyProductsState(products: products);
      return true;
    } catch (error) {
      state = MyProductsState(
        products: state.products,
        error: error.toString(),
      );
      return false;
    }
  }
}

final myProductsControllerProvider =
    StateNotifierProvider<MyProductsController, MyProductsState>((ref) {
      return MyProductsController(ref.watch(storeRepositoryProvider));
    });
