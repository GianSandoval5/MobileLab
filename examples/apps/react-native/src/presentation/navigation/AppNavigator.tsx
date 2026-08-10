import React from 'react';
import { Text } from 'react-native';
import { NavigationContainer, DefaultTheme } from '@react-navigation/native';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { colors } from '../../core/theme/colors';
import { AuthScreen } from '../../features/auth/presentation/screens/AuthScreen';
import { useAuthStore } from '../../features/auth/presentation/state/authStore';
import { BusinessDetailScreen } from '../../features/store/presentation/screens/BusinessDetailScreen';
import { BusinessesScreen } from '../../features/store/presentation/screens/BusinessesScreen';
import { CartScreen } from '../../features/store/presentation/screens/CartScreen';
import { CheckoutScreen } from '../../features/store/presentation/screens/CheckoutScreen';
import { EditProfileScreen } from '../../features/store/presentation/screens/EditProfileScreen';
import { HomeScreen } from '../../features/store/presentation/screens/HomeScreen';
import { MyProductsScreen } from '../../features/store/presentation/screens/MyProductsScreen';
import { ProductFormScreen } from '../../features/store/presentation/screens/ProductFormScreen';
import { ProfileScreen } from '../../features/store/presentation/screens/ProfileScreen';
import { PurchasesScreen } from '../../features/store/presentation/screens/PurchasesScreen';
import { CartButton } from '../components/CartButton';
import { MainTabsParamList, RootStackParamList } from './types';

const Stack = createNativeStackNavigator<RootStackParamList>();
const Tabs = createBottomTabNavigator<MainTabsParamList>();

const HomeIcon = () => <Text style={styles.tabIcon}>🏠</Text>;
const BusinessesIcon = () => <Text style={styles.tabIcon}>🏪</Text>;
const PurchasesIcon = () => <Text style={styles.tabIcon}>🧾</Text>;
const ProfileIcon = () => <Text style={styles.tabIcon}>👤</Text>;

const theme = {
  ...DefaultTheme,
  colors: {
    ...DefaultTheme.colors,
    primary: colors.primary,
    background: colors.background,
    card: '#FFF',
    text: colors.text,
    border: colors.border,
  },
};

function MainTabs() {
  return (
    <Tabs.Navigator
      screenOptions={{
        headerRight: CartButton,
        headerTitleStyle: { fontWeight: '800' },
        tabBarActiveTintColor: colors.primary,
        tabBarInactiveTintColor: colors.muted,
      }}
    >
      <Tabs.Screen
        name="Home"
        component={HomeScreen}
        options={{
          title: 'Descubre',
          tabBarLabel: 'Inicio',
          tabBarIcon: HomeIcon,
        }}
      />
      <Tabs.Screen
        name="Businesses"
        component={BusinessesScreen}
        options={{
          title: 'Negocios',
          tabBarLabel: 'Negocios',
          tabBarIcon: BusinessesIcon,
        }}
      />
      <Tabs.Screen
        name="Purchases"
        component={PurchasesScreen}
        options={{
          title: 'Mis compras',
          tabBarLabel: 'Compras',
          tabBarIcon: PurchasesIcon,
        }}
      />
      <Tabs.Screen
        name="Profile"
        component={ProfileScreen}
        options={{
          title: 'Mi perfil',
          tabBarLabel: 'Perfil',
          tabBarIcon: ProfileIcon,
        }}
      />
    </Tabs.Navigator>
  );
}

export function AppNavigator() {
  const authenticated = useAuthStore(state => !!state.user);
  return (
    <NavigationContainer theme={theme}>
      <Stack.Navigator
        screenOptions={{
          headerTitleStyle: { fontWeight: '800' },
          contentStyle: { backgroundColor: colors.background },
        }}
      >
        {!authenticated ? (
          <Stack.Screen
            name="Auth"
            component={AuthScreen}
            options={{ headerShown: false }}
          />
        ) : (
          <>
            <Stack.Screen
              name="Main"
              component={MainTabs}
              options={{ headerShown: false }}
            />
            <Stack.Screen
              name="Cart"
              component={CartScreen}
              options={{ title: 'Mi carrito' }}
            />
            <Stack.Screen
              name="Checkout"
              component={CheckoutScreen}
              options={{ title: 'Pagar' }}
            />
            <Stack.Screen
              name="BusinessDetail"
              component={BusinessDetailScreen}
              options={({ route }) => ({
                title: route.params.business.name,
                headerRight: CartButton,
              })}
            />
            <Stack.Screen
              name="EditProfile"
              component={EditProfileScreen}
              options={{ title: 'Editar perfil' }}
            />
            <Stack.Screen
              name="MyProducts"
              component={MyProductsScreen}
              options={{ title: 'Mis productos' }}
            />
            <Stack.Screen
              name="ProductForm"
              component={ProductFormScreen}
              options={({ route }) => ({
                title: route.params?.product
                  ? 'Editar producto'
                  : 'Publicar producto',
              })}
            />
          </>
        )}
      </Stack.Navigator>
    </NavigationContainer>
  );
}

const styles = { tabIcon: { fontSize: 20 } };
