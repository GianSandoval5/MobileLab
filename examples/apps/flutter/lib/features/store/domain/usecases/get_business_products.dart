import 'package:app_prueba_mobilelab/features/store/domain/entities/product.dart';
import 'package:app_prueba_mobilelab/features/store/domain/repositories/store_repository.dart';

class GetBusinessProducts {
  const GetBusinessProducts(this._repository);

  final StoreRepository _repository;

  Future<List<Product>> call(String businessId) {
    return _repository.getBusinessProducts(businessId);
  }
}
