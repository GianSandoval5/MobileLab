import {
  Business,
  Product,
} from '../../features/store/domain/entities/storeEntities';

export type RootStackParamList = {
  Auth: undefined;
  Main: undefined;
  Cart: undefined;
  Checkout: undefined;
  BusinessDetail: { business: Business };
  EditProfile: undefined;
  MyProducts: undefined;
  ProductForm: { product?: Product } | undefined;
};

export type MainTabsParamList = {
  Home: undefined;
  Businesses: undefined;
  Purchases: undefined;
  Profile: undefined;
};
