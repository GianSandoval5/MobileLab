import 'package:app_prueba_mobilelab/core/network/api_client.dart';
import 'package:app_prueba_mobilelab/features/auth/data/datasources/auth_remote_data_source.dart';
import 'package:app_prueba_mobilelab/features/auth/data/repositories/auth_repository_impl.dart';
import 'package:app_prueba_mobilelab/features/auth/domain/repositories/auth_repository.dart';
import 'package:app_prueba_mobilelab/features/auth/domain/usecases/login.dart';
import 'package:app_prueba_mobilelab/features/auth/domain/usecases/register.dart';
import 'package:app_prueba_mobilelab/features/store/data/datasources/store_remote_data_source.dart';
import 'package:app_prueba_mobilelab/features/store/data/repositories/store_repository_impl.dart';
import 'package:app_prueba_mobilelab/features/store/domain/repositories/store_repository.dart';
import 'package:app_prueba_mobilelab/features/store/domain/usecases/get_business_products.dart';
import 'package:app_prueba_mobilelab/features/store/domain/usecases/get_businesses.dart';
import 'package:app_prueba_mobilelab/features/store/domain/usecases/get_products.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

final apiClientProvider = Provider<ApiClient>((ref) => ApiClient());

final authRepositoryProvider = Provider<AuthRepository>((ref) {
  final remote = AuthRemoteDataSource(ref.watch(apiClientProvider));
  return AuthRepositoryImpl(remote);
});

final storeRepositoryProvider = Provider<StoreRepository>((ref) {
  final remote = StoreRemoteDataSource(ref.watch(apiClientProvider));
  return StoreRepositoryImpl(remote);
});

final loginUseCaseProvider = Provider<Login>(
  (ref) => Login(ref.watch(authRepositoryProvider)),
);

final registerUseCaseProvider = Provider<Register>(
  (ref) => Register(ref.watch(authRepositoryProvider)),
);

final getProductsUseCaseProvider = Provider<GetProducts>(
  (ref) => GetProducts(ref.watch(storeRepositoryProvider)),
);

final getBusinessesUseCaseProvider = Provider<GetBusinesses>(
  (ref) => GetBusinesses(ref.watch(storeRepositoryProvider)),
);

final getBusinessProductsUseCaseProvider = Provider<GetBusinessProducts>(
  (ref) => GetBusinessProducts(ref.watch(storeRepositoryProvider)),
);
