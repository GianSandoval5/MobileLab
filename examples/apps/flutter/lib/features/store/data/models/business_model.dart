import 'package:app_prueba_mobilelab/core/errors/app_exception.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/business.dart';

abstract final class BusinessModel {
  static Business fromJson(Map<String, dynamic> json) {
    final id = json['id'];
    final name = json['name'];
    final description = json['description'];
    final category = json['category'];
    final rating = json['rating'];

    if (id is! String ||
        name is! String ||
        description is! String ||
        category is! String ||
        rating is! num) {
      throw const AppException('Un negocio no tiene el formato esperado.');
    }

    return Business(
      id: id,
      name: name,
      description: description,
      category: category,
      rating: rating.toDouble(),
    );
  }
}
