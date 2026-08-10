import { ApiClient } from '../../../core/network/ApiClient';
import { AuthSession, User } from '../domain/entities/User';
import { AuthRepository } from '../domain/repositories/AuthRepository';

export class AuthRepositoryImpl implements AuthRepository {
  constructor(private readonly client: ApiClient) {}

  async login(email: string, password: string) {
    const session = await this.client.post<AuthSession>('/api/auth/login', {
      email,
      password,
    });
    this.client.setToken(session.token);
    return session;
  }

  async register(name: string, email: string, password: string) {
    const session = await this.client.post<AuthSession>('/api/auth/register', {
      name,
      email,
      password,
    });
    this.client.setToken(session.token);
    return session;
  }

  updateProfile(name: string, email: string, phone: string) {
    return this.client.put<User>('/api/profile', { name, email, phone });
  }
}
