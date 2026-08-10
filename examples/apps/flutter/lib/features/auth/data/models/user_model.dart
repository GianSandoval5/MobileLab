import 'package:app_prueba_mobilelab/core/errors/app_exception.dart';
import 'package:app_prueba_mobilelab/features/auth/domain/entities/app_user.dart';

abstract final class UserModel {
  static AppUser fromJson(Map<String, dynamic> json) {
    final id = json['id'];
    final name = json['name'];
    final email = json['email'];
    final phone = json['phone'] ?? '';

    if (id is! String ||
        name is! String ||
        email is! String ||
        phone is! String) {
      throw const AppException(
        'El usuario recibido no tiene el formato esperado.',
      );
    }

    return AppUser(id: id, name: name, email: email, phone: phone);
  }
}
