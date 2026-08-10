import 'package:app_prueba_mobilelab/core/errors/app_exception.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/product.dart';

abstract final class ProductModel {
  static Product fromJson(Map<String, dynamic> json) {
    final id = json['id'];
    final name = json['name'];
    final description = json['description'];
    final price = json['price'];
    final stock = json['stock'];
    final category = json['category'];
    final businessId = json['businessId'];
    final businessName = json['businessName'];

    if (id is! String ||
        name is! String ||
        description is! String ||
        price is! num ||
        stock is! num ||
        category is! String ||
        businessId is! String ||
        businessName is! String) {
      throw const AppException('Un producto no tiene el formato esperado.');
    }

    return Product(
      id: id,
      name: name,
      description: description,
      price: price.toDouble(),
      stock: stock.toInt(),
      category: category,
      businessId: businessId,
      businessName: businessName,
    );
  }

  static Map<String, dynamic> toJson(Product product) => {
    'id': product.id,
    'name': product.name,
    'description': product.description,
    'price': product.price,
    'stock': product.stock,
    'category': product.category,
    'businessId': product.businessId,
    'businessName': product.businessName,
  };
}
