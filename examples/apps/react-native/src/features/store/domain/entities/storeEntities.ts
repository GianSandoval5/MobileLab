export interface Product {
  id: string;
  name: string;
  description: string;
  price: number;
  stock: number;
  category: string;
  businessId: string;
  businessName: string;
}

export interface Business {
  id: string;
  name: string;
  description: string;
  category: string;
  rating: number;
}

export interface CartItem {
  product: Product;
  quantity: number;
}

export interface Purchase {
  id: string;
  createdAt: string;
  status: string;
  total: number;
  items: CartItem[];
}

export interface PaymentResult {
  purchaseId: string;
  status: string;
  message: string;
}

export const cartSubtotal = (item: CartItem) =>
  item.product.price * item.quantity;
