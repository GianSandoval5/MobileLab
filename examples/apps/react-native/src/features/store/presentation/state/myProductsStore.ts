import { create } from 'zustand';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { createJSONStorage, persist } from 'zustand/middleware';
import { storeRepository } from '../../../../core/di/container';
import { errorMessage } from '../../../../core/errors/AppError';
import { Product } from '../../domain/entities/storeEntities';

interface MyProductsState {
  products: Product[];
  loading: boolean;
  error?: string;
  load(): Promise<void>;
  save(product: Product, isNew: boolean): Promise<boolean>;
  reset(): void;
}

export const useMyProductsStore = create<MyProductsState>()(
  persist(
    (set, get) => ({
      products: [],
      loading: false,
      async load() {
        set({ loading: true, error: undefined });
        try {
          const remote = await storeRepository.getMyProducts();
          const local = get().products;
          const localIds = new Set(local.map(item => item.id));
          set({
            products: [
              ...local,
              ...remote.filter(item => !localIds.has(item.id)),
            ],
            loading: false,
          });
        } catch (error) {
          set({ loading: false, error: errorMessage(error) });
        }
      },
      async save(product, isNew) {
        set({ loading: true, error: undefined });
        try {
          const saved = isNew
            ? await storeRepository.createProduct(product)
            : await storeRepository.updateProduct(product);
          set({
            products: isNew
              ? [saved, ...get().products]
              : get().products.map(item =>
                  item.id === saved.id ? saved : item,
                ),
            loading: false,
          });
          return true;
        } catch (error) {
          set({ loading: false, error: errorMessage(error) });
          return false;
        }
      },
      reset: () => set({ products: [], loading: false, error: undefined }),
    }),
    {
      name: 'mobilelab-my-products',
      storage: createJSONStorage(() => AsyncStorage),
      partialize: state => ({ ...state, loading: false, error: undefined }),
    },
  ),
);
