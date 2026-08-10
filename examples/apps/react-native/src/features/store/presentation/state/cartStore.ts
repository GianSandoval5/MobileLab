import { create } from 'zustand';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { createJSONStorage, persist } from 'zustand/middleware';
import { storeRepository } from '../../../../core/di/container';
import { errorMessage } from '../../../../core/errors/AppError';
import {
  CartItem,
  Product,
  Purchase,
} from '../../domain/entities/storeEntities';

interface CartState {
  items: CartItem[];
  busy: boolean;
  error?: string;
  add(product: Product): Promise<boolean>;
  changeQuantity(item: CartItem, quantity: number): Promise<void>;
  remove(productId: string): Promise<void>;
  checkout(): Promise<Purchase | null>;
  completeCheckout(): void;
  reset(): void;
}

export const cartCount = (items: CartItem[]) =>
  items.reduce((sum, item) => sum + item.quantity, 0);
export const cartTotal = (items: CartItem[]) =>
  items.reduce((sum, item) => sum + item.product.price * item.quantity, 0);

export const useCartStore = create<CartState>()(
  persist(
    (set, get) => ({
      items: [],
      busy: false,
      async add(product) {
        const current = get().items;
        const index = current.findIndex(item => item.product.id === product.id);
        const quantity = index < 0 ? 1 : current[index].quantity + 1;
        set({ busy: true, error: undefined });
        try {
          if (index < 0) {
            await storeRepository.addCartItem(product, quantity);
            set({ items: [...current, { product, quantity }], busy: false });
          } else {
            await storeRepository.updateCartItem(product, quantity);
            set({
              items: current.map((item, itemIndex) =>
                itemIndex === index ? { ...item, quantity } : item,
              ),
              busy: false,
            });
          }
          return true;
        } catch (error) {
          set({ busy: false, error: errorMessage(error) });
          return false;
        }
      },
      async changeQuantity(item, quantity) {
        if (quantity <= 0) {
          await get().remove(item.product.id);
          return;
        }
        set({ busy: true, error: undefined });
        try {
          await storeRepository.updateCartItem(item.product, quantity);
          set({
            items: get().items.map(current =>
              current.product.id === item.product.id
                ? { ...current, quantity }
                : current,
            ),
            busy: false,
          });
        } catch (error) {
          set({ busy: false, error: errorMessage(error) });
        }
      },
      async remove(productId) {
        set({ busy: true, error: undefined });
        try {
          await storeRepository.removeCartItem(productId);
          set({
            items: get().items.filter(item => item.product.id !== productId),
            busy: false,
          });
        } catch (error) {
          set({ busy: false, error: errorMessage(error) });
        }
      },
      async checkout() {
        const items = [...get().items];
        if (!items.length) {
          return null;
        }
        const total = cartTotal(items);
        set({ busy: true, error: undefined });
        try {
          const payment = await storeRepository.pay(items, total);
          set({ busy: false });
          return {
            id: payment.purchaseId,
            createdAt: new Date().toISOString(),
            status: 'Pagado',
            total,
            items,
          };
        } catch (error) {
          set({ busy: false, error: errorMessage(error) });
          return null;
        }
      },
      completeCheckout: () => set({ items: [], error: undefined }),
      reset: () => set({ items: [], busy: false, error: undefined }),
    }),
    {
      name: 'mobilelab-cart',
      storage: createJSONStorage(() => AsyncStorage),
      partialize: state => ({ ...state, busy: false, error: undefined }),
    },
  ),
);
