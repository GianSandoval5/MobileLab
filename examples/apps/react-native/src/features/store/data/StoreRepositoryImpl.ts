import { ApiClient } from '../../../core/network/ApiClient';
import {
  Business,
  CartItem,
  PaymentResult,
  Product,
  Purchase,
} from '../domain/entities/storeEntities';
import { StoreRepository } from '../domain/repositories/StoreRepository';

export class StoreRepositoryImpl implements StoreRepository {
  constructor(private readonly client: ApiClient) {}

  getProducts() {
    return this.client.get<Product[]>('/api/products');
  }

  getBusinesses() {
    return this.client.get<Business[]>('/api/businesses');
  }

  getBusinessProducts(id: string) {
    return this.client.get<Product[]>(`/api/businesses/${id}/products`);
  }

  async addCartItem(product: Product, quantity: number) {
    await this.client.post('/api/cart/items', {
      productId: product.id,
      quantity,
    });
  }

  async updateCartItem(product: Product, quantity: number) {
    await this.client.patch('/api/cart/items', {
      productId: product.id,
      quantity,
    });
  }

  async removeCartItem(productId: string) {
    await this.client.delete('/api/cart/items', { productId });
  }

  pay(items: CartItem[], total: number) {
    return this.client.post<PaymentResult>('/api/payments', {
      paymentMethod: 'card',
      total,
      items: items.map(item => ({
        productId: item.product.id,
        quantity: item.quantity,
      })),
    });
  }

  getPurchases() {
    return this.client.get<Purchase[]>('/api/purchases');
  }

  getMyProducts() {
    return this.client.get<Product[]>('/api/my-products');
  }

  async createProduct(product: Product) {
    const created = await this.client.post<Product>(
      '/api/my-products',
      product,
    );
    return { ...product, id: created.id };
  }

  async updateProduct(product: Product) {
    await this.client.put(`/api/my-products/${product.id}`, product);
    return product;
  }
}
