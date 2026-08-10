import 'package:app_prueba_mobilelab/features/store/data/datasources/store_remote_data_source.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/business.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/cart_item.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/product.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/purchase.dart';
import 'package:app_prueba_mobilelab/features/store/domain/repositories/store_repository.dart';

class StoreRepositoryImpl implements StoreRepository {
  const StoreRepositoryImpl(this._remote);

  final StoreRemoteDataSource _remote;

  @override
  Future<List<Product>> getProducts() => _remote.getProducts();

  @override
  Future<List<Business>> getBusinesses() => _remote.getBusinesses();

  @override
  Future<List<Product>> getBusinessProducts(String businessId) {
    return _remote.getBusinessProducts(businessId);
  }

  @override
  Future<void> addCartItem(Product product, int quantity) {
    return _remote.addCartItem(product, quantity);
  }

  @override
  Future<void> updateCartItem(Product product, int quantity) {
    return _remote.updateCartItem(product, quantity);
  }

  @override
  Future<void> removeCartItem(String productId) {
    return _remote.removeCartItem(productId);
  }

  @override
  Future<PaymentResult> pay(List<CartItem> items, double total) {
    return _remote.pay(items, total);
  }

  @override
  Future<List<Purchase>> getPurchases() => _remote.getPurchases();

  @override
  Future<List<Product>> getMyProducts() => _remote.getMyProducts();

  @override
  Future<Product> createProduct(Product product) {
    return _remote.createProduct(product);
  }

  @override
  Future<Product> updateProduct(Product product) {
    return _remote.updateProduct(product);
  }
}
