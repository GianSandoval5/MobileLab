import { AuthSession, User } from '../entities/User';

export interface AuthRepository {
  login(email: string, password: string): Promise<AuthSession>;
  register(name: string, email: string, password: string): Promise<AuthSession>;
  updateProfile(name: string, email: string, phone: string): Promise<User>;
}
