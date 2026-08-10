import React, { useEffect } from 'react';
import { FlatList, StyleSheet, Text, View } from 'react-native';
import { colors } from '../../../../core/theme/colors';
import {
  ErrorView,
  LoadingView,
} from '../../../../presentation/components/StateView';
import { Purchase, cartSubtotal } from '../../domain/entities/storeEntities';
import { usePurchasesStore } from '../state/purchasesStore';

const Separator = () => <View style={styles.separator} />;

export function PurchasesScreen() {
  const { purchases, loading, error, load } = usePurchasesStore();
  useEffect(() => {
    if (!purchases.length) {
      load();
    }
  }, [load, purchases.length]);

  if (loading && !purchases.length) {
    return <LoadingView />;
  }
  if (error && !purchases.length) {
    return <ErrorView error={error} retry={load} />;
  }
  if (!purchases.length) {
    return (
      <View style={styles.empty}>
        <Text>Todavía no tienes compras.</Text>
      </View>
    );
  }

  const renderPurchase = ({ item }: { item: Purchase }) => (
    <View style={styles.card}>
      <View style={styles.header}>
        <View>
          <Text style={styles.id}>Compra {item.id}</Text>
          <Text style={styles.meta}>
            {new Date(item.createdAt).toLocaleDateString()} · {item.status}
          </Text>
        </View>
        <Text style={styles.total}>S/ {item.total.toFixed(2)}</Text>
      </View>
      <View style={styles.divider} />
      {item.items.map(cartItem => (
        <View key={cartItem.product.id} style={styles.item}>
          <View style={styles.itemInfo}>
            <Text style={styles.itemName}>{cartItem.product.name}</Text>
            <Text style={styles.meta}>{cartItem.quantity} unidad(es)</Text>
          </View>
          <Text>S/ {cartSubtotal(cartItem).toFixed(2)}</Text>
        </View>
      ))}
    </View>
  );

  return (
    <FlatList
      data={purchases}
      keyExtractor={item => item.id}
      renderItem={renderPurchase}
      contentContainerStyle={styles.content}
      ItemSeparatorComponent={Separator}
      refreshing={loading}
      onRefresh={load}
    />
  );
}

const styles = StyleSheet.create({
  content: { padding: 14, paddingBottom: 28 },
  separator: { height: 12 },
  card: {
    padding: 16,
    borderRadius: 18,
    backgroundColor: '#FFF',
    borderWidth: 1,
    borderColor: colors.border,
  },
  header: { flexDirection: 'row', justifyContent: 'space-between' },
  id: { fontWeight: '800', color: colors.text },
  meta: { fontSize: 12, color: colors.muted, marginTop: 3 },
  total: { fontWeight: '900', color: colors.primary },
  divider: { height: 1, backgroundColor: colors.border, marginVertical: 12 },
  item: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginVertical: 5,
  },
  itemInfo: { flex: 1 },
  itemName: { fontWeight: '600', color: colors.text },
  empty: { flex: 1, alignItems: 'center', justifyContent: 'center' },
});
