import 'package:app_prueba_mobilelab/core/errors/app_exception.dart';
import 'package:app_prueba_mobilelab/core/network/api_client.dart';
import 'package:app_prueba_mobilelab/features/auth/data/models/user_model.dart';
import 'package:app_prueba_mobilelab/features/auth/domain/entities/app_user.dart';

class AuthRemoteDataSource {
  const AuthRemoteDataSource(this._client);

  final ApiClient _client;

  Future<AuthSession> login({required String email, required String password}) {
    return _authenticate(
      '/api/auth/login',
      body: {'email': email, 'password': password},
    );
  }

  Future<AuthSession> register({
    required String name,
    required String email,
    required String password,
  }) {
    return _authenticate(
      '/api/auth/register',
      body: {'name': name, 'email': email, 'password': password},
    );
  }

  Future<AuthSession> _authenticate(
    String path, {
    required Map<String, dynamic> body,
  }) async {
    final json = await _client.post(path, body: body);
    if (json is! Map<String, dynamic> ||
        json['token'] is! String ||
        json['user'] is! Map<String, dynamic>) {
      throw const AppException(
        'La sesión recibida no tiene el formato esperado.',
      );
    }

    final session = AuthSession(
      token: json['token'] as String,
      user: UserModel.fromJson(json['user'] as Map<String, dynamic>),
    );
    _client.token = session.token;
    return session;
  }

  Future<AppUser> getProfile() async {
    final json = await _client.get('/api/profile');
    if (json is! Map<String, dynamic>) {
      throw const AppException(
        'El perfil recibido no tiene el formato esperado.',
      );
    }
    return UserModel.fromJson(json);
  }

  Future<AppUser> updateProfile({
    required String name,
    required String email,
    required String phone,
  }) async {
    final json = await _client.put(
      '/api/profile',
      body: {'name': name, 'email': email, 'phone': phone},
    );
    if (json is! Map<String, dynamic>) {
      throw const AppException(
        'El perfil recibido no tiene el formato esperado.',
      );
    }
    return UserModel.fromJson(json);
  }
}
