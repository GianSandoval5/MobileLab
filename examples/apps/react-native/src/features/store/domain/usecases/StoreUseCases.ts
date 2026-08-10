import { StoreRepository } from '../repositories/StoreRepository';

export class GetProducts {
  constructor(private readonly repository: StoreRepository) {}
  execute() {
    return this.repository.getProducts();
  }
}

export class GetBusinesses {
  constructor(private readonly repository: StoreRepository) {}
  execute() {
    return this.repository.getBusinesses();
  }
}

export class GetBusinessProducts {
  constructor(private readonly repository: StoreRepository) {}
  execute(id: string) {
    return this.repository.getBusinessProducts(id);
  }
}
