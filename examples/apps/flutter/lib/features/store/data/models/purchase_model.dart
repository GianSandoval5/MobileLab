import 'package:app_prueba_mobilelab/core/errors/app_exception.dart';
import 'package:app_prueba_mobilelab/features/store/data/models/product_model.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/cart_item.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/purchase.dart';

abstract final class PurchaseModel {
  static Purchase fromJson(Map<String, dynamic> json) {
    final id = json['id'];
    final createdAt = json['createdAt'];
    final status = json['status'];
    final total = json['total'];
    final rawItems = json['items'];

    if (id is! String ||
        createdAt is! String ||
        status is! String ||
        total is! num ||
        rawItems is! List) {
      throw const AppException('Una compra no tiene el formato esperado.');
    }

    final items = rawItems.map((raw) {
      if (raw is! Map<String, dynamic> ||
          raw['quantity'] is! num ||
          raw['product'] is! Map<String, dynamic>) {
        throw const AppException('Un artículo comprado no es válido.');
      }
      return CartItem(
        product: ProductModel.fromJson(raw['product'] as Map<String, dynamic>),
        quantity: (raw['quantity'] as num).toInt(),
      );
    }).toList();

    return Purchase(
      id: id,
      createdAt: DateTime.parse(createdAt),
      status: status,
      total: total.toDouble(),
      items: items,
    );
  }
}
