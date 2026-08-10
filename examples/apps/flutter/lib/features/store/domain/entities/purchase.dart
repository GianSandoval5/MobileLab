import 'package:app_prueba_mobilelab/features/store/domain/entities/cart_item.dart';

class Purchase {
  const Purchase({
    required this.id,
    required this.createdAt,
    required this.status,
    required this.total,
    required this.items,
  });

  final String id;
  final DateTime createdAt;
  final String status;
  final double total;
  final List<CartItem> items;
}

class PaymentResult {
  const PaymentResult({
    required this.purchaseId,
    required this.status,
    required this.message,
  });

  final String purchaseId;
  final String status;
  final String message;
}
