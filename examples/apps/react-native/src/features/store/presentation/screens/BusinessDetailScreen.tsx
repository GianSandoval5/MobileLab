import React, { useEffect, useState } from 'react';
import { FlatList, StyleSheet, Text, View } from 'react-native';
import { NativeStackScreenProps } from '@react-navigation/native-stack';
import { getBusinessProductsUseCase } from '../../../../core/di/container';
import { errorMessage } from '../../../../core/errors/AppError';
import { colors } from '../../../../core/theme/colors';
import { ProductCard } from '../../../../presentation/components/ProductCard';
import {
  ErrorView,
  LoadingView,
} from '../../../../presentation/components/StateView';
import { RootStackParamList } from '../../../../presentation/navigation/types';
import { Product } from '../../domain/entities/storeEntities';

type Props = NativeStackScreenProps<RootStackParamList, 'BusinessDetail'>;

export function BusinessDetailScreen({ route }: Props) {
  const { business } = route.params;
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const load = async () => {
    setLoading(true);
    setError('');
    try {
      setProducts(await getBusinessProductsUseCase.execute(business.id));
    } catch (loadError) {
      setError(errorMessage(loadError));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
    // El negocio cambia al crear una nueva pantalla de detalle.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [business.id]);

  if (loading && !products.length) {
    return <LoadingView />;
  }
  if (error && !products.length) {
    return <ErrorView error={error} retry={load} />;
  }

  return (
    <FlatList
      data={products}
      numColumns={2}
      keyExtractor={item => item.id}
      contentContainerStyle={styles.content}
      columnWrapperStyle={styles.row}
      ListHeaderComponent={
        <View style={styles.header}>
          <Text style={styles.description}>{business.description}</Text>
          <Text style={styles.meta}>
            ⭐ {business.rating} · {business.category}
          </Text>
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
  header: {
    backgroundColor: colors.primarySoft,
    borderRadius: 18,
    padding: 18,
    marginBottom: 16,
  },
  description: { fontSize: 16, color: colors.text, fontWeight: '600' },
  meta: { marginTop: 8, color: colors.muted },
});
