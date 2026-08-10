import {
  Business,
  CartItem,
  PaymentResult,
  Product,
  Purchase,
} from '../entities/storeEntities';

export interface StoreRepository {
  getProducts(): Promise<Product[]>;
  getBusinesses(): Promise<Business[]>;
  getBusinessProducts(id: string): Promise<Product[]>;
  addCartItem(product: Product, quantity: number): Promise<void>;
  updateCartItem(product: Product, quantity: number): Promise<void>;
  removeCartItem(productId: string): Promise<void>;
  pay(items: CartItem[], total: number): Promise<PaymentResult>;
  getPurchases(): Promise<Purchase[]>;
  getMyProducts(): Promise<Product[]>;
  createProduct(product: Product): Promise<Product>;
  updateProduct(product: Product): Promise<Product>;
}
