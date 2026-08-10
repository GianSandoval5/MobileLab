import 'package:app_prueba_mobilelab/features/store/domain/entities/business.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/cart_item.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/product.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/purchase.dart';

abstract interface class StoreRepository {
  Future<List<Product>> getProducts();

  Future<List<Business>> getBusinesses();

  Future<List<Product>> getBusinessProducts(String businessId);

  Future<void> addCartItem(Product product, int quantity);

  Future<void> updateCartItem(Product product, int quantity);

  Future<void> removeCartItem(String productId);

  Future<PaymentResult> pay(List<CartItem> items, double total);

  Future<List<Purchase>> getPurchases();

  Future<List<Product>> getMyProducts();

  Future<Product> createProduct(Product product);

  Future<Product> updateProduct(Product product);
}
