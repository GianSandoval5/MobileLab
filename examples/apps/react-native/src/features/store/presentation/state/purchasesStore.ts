import { create } from 'zustand';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { createJSONStorage, persist } from 'zustand/middleware';
import { storeRepository } from '../../../../core/di/container';
import { errorMessage } from '../../../../core/errors/AppError';
import { Purchase } from '../../domain/entities/storeEntities';

interface PurchasesState {
  purchases: Purchase[];
  loading: boolean;
  error?: string;
  load(): Promise<void>;
  add(purchase: Purchase): void;
  reset(): void;
}

export const usePurchasesStore = create<PurchasesState>()(
  persist(
    (set, get) => ({
      purchases: [],
      loading: false,
      async load() {
        set({ loading: true, error: undefined });
        try {
          const remote = await storeRepository.getPurchases();
          const local = get().purchases;
          const localIds = new Set(local.map(item => item.id));
          set({
            purchases: [
              ...local,
              ...remote.filter(item => !localIds.has(item.id)),
            ],
            loading: false,
          });
        } catch (error) {
          set({ loading: false, error: errorMessage(error) });
        }
      },
      add: purchase =>
        set(state => {
          const duplicate = state.purchases.some(
            item => item.id === purchase.id,
          );
          const saved = duplicate
            ? { ...purchase, id: `${purchase.id}-${Date.now()}` }
            : purchase;
          return { purchases: [saved, ...state.purchases] };
        }),
      reset: () => set({ purchases: [], loading: false, error: undefined }),
    }),
    {
      name: 'mobilelab-purchases',
      storage: createJSONStorage(() => AsyncStorage),
      partialize: state => ({ ...state, loading: false, error: undefined }),
    },
  ),
);
