import 'package:app_prueba_mobilelab/features/auth/domain/entities/app_user.dart';

abstract interface class AuthRepository {
  Future<AuthSession> login({required String email, required String password});

  Future<AuthSession> register({
    required String name,
    required String email,
    required String password,
  });

  Future<AppUser> getProfile();

  Future<AppUser> updateProfile({
    required String name,
    required String email,
    required String phone,
  });
}
