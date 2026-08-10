import 'package:app_prueba_mobilelab/features/store/domain/entities/business.dart';
import 'package:app_prueba_mobilelab/features/store/domain/repositories/store_repository.dart';

class GetBusinesses {
  const GetBusinesses(this._repository);

  final StoreRepository _repository;

  Future<List<Business>> call() => _repository.getBusinesses();
}
