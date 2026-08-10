import { AuthRepository } from '../repositories/AuthRepository';

export class Login {
  constructor(private readonly repository: AuthRepository) {}
  execute(email: string, password: string) {
    return this.repository.login(email, password);
  }
}

export class Register {
  constructor(private readonly repository: AuthRepository) {}
  execute(name: string, email: string, password: string) {
    return this.repository.register(name, email, password);
  }
}
