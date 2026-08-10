import React, { useEffect } from 'react';
import { FlatList, StyleSheet, Text, View } from 'react-native';
import { colors } from '../../../../core/theme/colors';
import { useAuthStore } from '../../../auth/presentation/state/authStore';
import { ProductCard } from '../../../../presentation/components/ProductCard';
import {
  ErrorView,
  LoadingView,
} from '../../../../presentation/components/StateView';
import { useCatalogStore } from '../state/catalogStore';

export function HomeScreen() {
  const user = useAuthStore(state => state.user);
  const { products, loadingProducts, productsError, loadProducts } =
    useCatalogStore();

  useEffect(() => {
    if (!products.length) {
      loadProducts();
    }
  }, [loadProducts, products.length]);

  if (loadingProducts && !products.length) {
    return <LoadingView />;
  }
  if (productsError && !products.length) {
    return <ErrorView error={productsError} retry={loadProducts} />;
  }

  return (
    <FlatList
      data={products}
      numColumns={2}
      keyExtractor={item => item.id}
      contentContainerStyle={styles.content}
      columnWrapperStyle={styles.row}
      refreshing={loadingProducts}
      onRefresh={loadProducts}
      ListHeaderComponent={
        <View>
          <View style={styles.hero}>
            <Text style={styles.heroTitle}>
              Hola, {user?.name.split(' ')[0]} 👋
            </Text>
            <Text style={styles.heroText}>
              Encuentra productos de negocios locales.
            </Text>
          </View>
          <Text style={styles.sectionTitle}>Productos destacados</Text>
        </View>
      }
      renderItem={({ item }) => (
        <View style={styles.item}>
          <ProductCard product={item} />
        </View>
      )}
    />
  );
}

const styles = StyleSheet.create({
  content: { padding: 14, paddingBottom: 28 },
  row: { gap: 10 },
  item: { flex: 1, marginBottom: 10 },
  hero: {
    backgroundColor: colors.primary,
    borderRadius: 22,
    padding: 20,
    marginBottom: 22,
  },
  heroTitle: { fontSize: 24, fontWeight: '800', color: '#FFF' },
  heroText: { color: '#DDDDFD', marginTop: 6 },
  sectionTitle: {
    fontSize: 20,
    fontWeight: '800',
    color: colors.text,
    marginBottom: 14,
  },
});
