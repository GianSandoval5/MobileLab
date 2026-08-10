import {
  cartSubtotal,
  Product,
} from '../src/features/store/domain/entities/storeEntities';
import {
  cartCount,
  cartTotal,
} from '../src/features/store/presentation/state/cartStore';

const product: Product = {
  id: 'prod-001',
  name: 'Audífonos Nova',
  description: 'Audio inalámbrico',
  price: 149.9,
  stock: 18,
  category: 'Tecnología',
  businessId: 'biz-001',
  businessName: 'Nova Tech',
};

test('calcula cantidades y totales del carrito', () => {
  const items = [{ product, quantity: 2 }];
  expect(cartSubtotal(items[0])).toBeCloseTo(299.8);
  expect(cartCount(items)).toBe(2);
  expect(cartTotal(items)).toBeCloseTo(299.8);
});
