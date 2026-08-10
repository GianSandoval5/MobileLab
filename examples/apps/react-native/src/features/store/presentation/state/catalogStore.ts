import { create } from 'zustand';
import { errorMessage } from '../../../../core/errors/AppError';
import {
  getBusinessesUseCase,
  getProductsUseCase,
} from '../../../../core/di/container';
import { Business, Product } from '../../domain/entities/storeEntities';

interface CatalogState {
  products: Product[];
  businesses: Business[];
  loadingProducts: boolean;
  loadingBusinesses: boolean;
  productsError?: string;
  businessesError?: string;
  loadProducts(): Promise<void>;
  loadBusinesses(): Promise<void>;
}

export const useCatalogStore = create<CatalogState>(set => ({
  products: [],
  businesses: [],
  loadingProducts: false,
  loadingBusinesses: false,
  async loadProducts() {
    set({ loadingProducts: true, productsError: undefined });
    try {
      set({
        products: await getProductsUseCase.execute(),
        loadingProducts: false,
      });
    } catch (error) {
      set({ loadingProducts: false, productsError: errorMessage(error) });
    }
  },
  async loadBusinesses() {
    set({ loadingBusinesses: true, businessesError: undefined });
    try {
      set({
        businesses: await getBusinessesUseCase.execute(),
        loadingBusinesses: false,
      });
    } catch (error) {
      set({ loadingBusinesses: false, businessesError: errorMessage(error) });
    }
  },
}));
