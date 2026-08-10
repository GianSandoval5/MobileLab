import { ApiClient } from '../network/ApiClient';
import { AuthRepositoryImpl } from '../../features/auth/data/AuthRepositoryImpl';
import {
  Login,
  Register,
} from '../../features/auth/domain/usecases/AuthUseCases';
import { StoreRepositoryImpl } from '../../features/store/data/StoreRepositoryImpl';
import {
  GetBusinesses,
  GetBusinessProducts,
  GetProducts,
} from '../../features/store/domain/usecases/StoreUseCases';

export const apiClient = new ApiClient();
export const authRepository = new AuthRepositoryImpl(apiClient);
export const storeRepository = new StoreRepositoryImpl(apiClient);

export const loginUseCase = new Login(authRepository);
export const registerUseCase = new Register(authRepository);
export const getProductsUseCase = new GetProducts(storeRepository);
export const getBusinessesUseCase = new GetBusinesses(storeRepository);
export const getBusinessProductsUseCase = new GetBusinessProducts(
  storeRepository,
);
