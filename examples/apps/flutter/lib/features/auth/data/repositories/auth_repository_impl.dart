import 'package:app_prueba_mobilelab/features/auth/data/datasources/auth_remote_data_source.dart';
import 'package:app_prueba_mobilelab/features/auth/domain/entities/app_user.dart';
import 'package:app_prueba_mobilelab/features/auth/domain/repositories/auth_repository.dart';

class AuthRepositoryImpl implements AuthRepository {
  const AuthRepositoryImpl(this._remote);

  final AuthRemoteDataSource _remote;

  @override
  Future<AuthSession> login({required String email, required String password}) {
    return _remote.login(email: email, password: password);
  }

  @override
  Future<AuthSession> register({
    required String name,
    required String email,
    required String password,
  }) {
    return _remote.register(name: name, email: email, password: password);
  }

  @override
  Future<AppUser> getProfile() => _remote.getProfile();

  @override
  Future<AppUser> updateProfile({
    required String name,
    required String email,
    required String phone,
  }) {
    return _remote.updateProfile(name: name, email: email, phone: phone);
  }
}
