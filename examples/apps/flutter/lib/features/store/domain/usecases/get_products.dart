import 'package:app_prueba_mobilelab/features/store/domain/entities/product.dart';
import 'package:app_prueba_mobilelab/features/store/domain/repositories/store_repository.dart';

class GetProducts {
  const GetProducts(this._repository);

  final StoreRepository _repository;

  Future<List<Product>> call() => _repository.getProducts();
}
