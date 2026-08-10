import 'package:app_prueba_mobilelab/app.dart';
import 'package:app_prueba_mobilelab/features/store/data/models/product_model.dart';
import 'package:app_prueba_mobilelab/features/store/domain/entities/cart_item.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('muestra la pantalla de inicio de sesión', (tester) async {
    await tester.pumpWidget(const ProviderScope(child: MobileLabShopApp()));

    expect(find.text('Bienvenido'), findsOneWidget);
    expect(find.text('Ingresar'), findsOneWidget);
    expect(find.text('Crear una cuenta'), findsOneWidget);
  });

  test('convierte un producto y calcula el subtotal del carrito', () {
    final product = ProductModel.fromJson({
      'id': 'prod-001',
      'name': 'Audífonos Nova',
      'description': 'Audio inalámbrico',
      'price': 149.90,
      'stock': 18,
      'category': 'Tecnología',
      'businessId': 'biz-001',
      'businessName': 'Nova Tech',
    });

    final item = CartItem(product: product, quantity: 2);

    expect(product.name, 'Audífonos Nova');
    expect(item.subtotal, closeTo(299.80, 0.001));
  });
}
