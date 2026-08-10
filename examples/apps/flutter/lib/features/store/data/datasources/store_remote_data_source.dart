import 'package:app_prueba_mobilelab/core/errors/app_exception.dart';
import 'package:app_prueba_mobilelab/core/network/api_client.dart';
import 'package:app_prueba_mobilelab/features/store/data/models/business_model.dart';
import 'package:app_prueba_mobilelab/features/store/data/models/product_model.dart';
import 'package:app_prueba_mobilelab/features/store/data/models/purchase_model.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/business.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/cart_item.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/product.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/purchase.dart';

class StoreRemoteDataSource {
  const StoreRemoteDataSource(this._client);

  final ApiClient _client;

  Future<List<Product>> getProducts() async {
    return _productList(await _client.get('/api/products'));
  }

  Future<List<Business>> getBusinesses() async {
    final json = await _client.get('/api/businesses');
    if (json is! List) {
      throw const AppException('La lista de negocios no es válida.');
    }
    return json.map((item) {
      if (item is! Map<String, dynamic>) {
        throw const AppException('Un negocio no es válido.');
      }
      return BusinessModel.fromJson(item);
    }).toList();
  }

  Future<List<Product>> getBusinessProducts(String businessId) async {
    return _productList(
      await _client.get('/api/businesses/$businessId/products'),
    );
  }

  Future<void> addCartItem(Product product, int quantity) async {
    await _client.post(
      '/api/cart/items',
      body: {'productId': product.id, 'quantity': quantity},
    );
  }

  Future<void> updateCartItem(Product product, int quantity) async {
    await _client.patch(
      '/api/cart/items',
      body: {'productId': product.id, 'quantity': quantity},
    );
  }

  Future<void> removeCartItem(String productId) async {
    await _client.delete('/api/cart/items', body: {'productId': productId});
  }

  Future<PaymentResult> pay(List<CartItem> items, double total) async {
    final json = await _client.post(
      '/api/payments',
      body: {
        'paymentMethod': 'card',
        'total': total,
        'items': items
            .map(
              (item) => {
                'productId': item.product.id,
                'quantity': item.quantity,
              },
            )
            .toList(),
      },
    );
    if (json is! Map<String, dynamic> ||
        json['purchaseId'] is! String ||
        json['status'] is! String ||
        json['message'] is! String) {
      throw const AppException('La confirmación de pago no es válida.');
    }
    return PaymentResult(
      purchaseId: json['purchaseId'] as String,
      status: json['status'] as String,
      message: json['message'] as String,
    );
  }

  Future<List<Purchase>> getPurchases() async {
    final json = await _client.get('/api/purchases');
    if (json is! List) {
      throw const AppException('La lista de compras no es válida.');
    }
    return json.map((item) {
      if (item is! Map<String, dynamic>) {
        throw const AppException('Una compra no es válida.');
      }
      return PurchaseModel.fromJson(item);
    }).toList();
  }

  Future<List<Product>> getMyProducts() async {
    return _productList(await _client.get('/api/my-products'));
  }

  Future<Product> createProduct(Product product) async {
    final json = await _client.post(
      '/api/my-products',
      body: ProductModel.toJson(product),
    );
    if (json is! Map<String, dynamic>) {
      throw const AppException('El producto creado no es válido.');
    }
    final created = ProductModel.fromJson(json);
    return product.copyWith(id: created.id);
  }

  Future<Product> updateProduct(Product product) async {
    final json = await _client.put(
      '/api/my-products/${product.id}',
      body: ProductModel.toJson(product),
    );
    if (json is! Map<String, dynamic>) {
      throw const AppException('El producto actualizado no es válido.');
    }
    ProductModel.fromJson(json);
    return product;
  }

  List<Product> _productList(dynamic json) {
    if (json is! List) {
      throw const AppException('La lista de productos no es válida.');
    }
    return json.map((item) {
      if (item is! Map<String, dynamic>) {
        throw const AppException('Un producto no es válido.');
      }
      return ProductModel.fromJson(item);
    }).toList();
  }
}
