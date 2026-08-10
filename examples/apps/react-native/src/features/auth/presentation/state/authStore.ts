import { create } from 'zustand';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { createJSONStorage, persist } from 'zustand/middleware';
import { errorMessage } from '../../../../core/errors/AppError';
import {
  apiClient,
  authRepository,
  loginUseCase,
  registerUseCase,
} from '../../../../core/di/container';
import { User } from '../../domain/entities/User';

interface AuthState {
  user: User | null;
  loading: boolean;
  error?: string;
  login(email: string, password: string): Promise<boolean>;
  register(name: string, email: string, password: string): Promise<boolean>;
  updateProfile(name: string, email: string, phone: string): Promise<boolean>;
  clearError(): void;
  logout(): void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    set => ({
      user: null,
      loading: false,
      async login(email, password) {
        set({ loading: true, error: undefined });
        try {
          const session = await loginUseCase.execute(email, password);
          set({ user: session.user, loading: false });
          return true;
        } catch (error) {
          set({ loading: false, error: errorMessage(error) });
          return false;
        }
      },
      async register(name, email, password) {
        set({ loading: true, error: undefined });
        try {
          const session = await registerUseCase.execute(name, email, password);
          set({ user: session.user, loading: false });
          return true;
        } catch (error) {
          set({ loading: false, error: errorMessage(error) });
          return false;
        }
      },
      async updateProfile(name, email, phone) {
        set({ loading: true, error: undefined });
        try {
          const response = await authRepository.updateProfile(
            name,
            email,
            phone,
          );
          set({
            user: { ...response, name, email, phone },
            loading: false,
          });
          return true;
        } catch (error) {
          set({ loading: false, error: errorMessage(error) });
          return false;
        }
      },
      clearError: () => set({ error: undefined }),
      logout: () => {
        apiClient.setToken(undefined);
        set({ user: null, loading: false, error: undefined });
      },
    }),
    {
      name: 'mobilelab-auth',
      storage: createJSONStorage(() => AsyncStorage),
      partialize: state => ({ ...state, loading: false, error: undefined }),
    },
  ),
);
