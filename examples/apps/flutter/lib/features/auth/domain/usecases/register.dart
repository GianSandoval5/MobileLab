import 'package:app_prueba_mobilelab/features/auth/domain/entities/app_user.dart';
import 'package:app_prueba_mobilelab/features/auth/domain/repositories/auth_repository.dart';

class Register {
  const Register(this._repository);

  final AuthRepository _repository;

  Future<AuthSession> call({
    required String name,
    required String email,
    required String password,
  }) {
    return _repository.register(name: name, email: email, password: password);
  }
}
