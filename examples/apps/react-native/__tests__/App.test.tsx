import { Product } from '../src/features/store/domain/entities/storeEntities';

test('representa un producto del catálogo', () => {
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

  expect(product.name).toBe('Audífonos Nova');
  expect(product.stock).toBeGreaterThan(0);
});
