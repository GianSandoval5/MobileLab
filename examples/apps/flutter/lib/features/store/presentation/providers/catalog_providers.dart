import 'package:app_prueba_mobilelab/core/providers/repository_providers.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/business.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/product.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

final productsProvider = FutureProvider<List<Product>>((ref) {
  return ref.watch(getProductsUseCaseProvider)();
});

final businessesProvider = FutureProvider<List<Business>>((ref) {
  return ref.watch(getBusinessesUseCaseProvider)();
});

final businessProductsProvider = FutureProvider.family<List<Product>, String>((
  ref,
  businessId,
) {
  return ref.watch(getBusinessProductsUseCaseProvider)(businessId);
});
