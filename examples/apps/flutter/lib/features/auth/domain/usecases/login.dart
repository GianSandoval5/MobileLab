import 'package:app_prueba_mobilelab/features/auth/domain/entities/app_user.dart';
import 'package:app_prueba_mobilelab/features/auth/domain/repositories/auth_repository.dart';

class Login {
  const Login(this._repository);

  final AuthRepository _repository;

  Future<AuthSession> call({required String email, required String password}) {
    return _repository.login(email: email, password: password);
  }
}
