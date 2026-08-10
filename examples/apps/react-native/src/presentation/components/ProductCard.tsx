import React from 'react';
import { Alert, Pressable, StyleSheet, Text, View } from 'react-native';
import { colors } from '../../core/theme/colors';
import { Product } from '../../features/store/domain/entities/storeEntities';
import { useCartStore } from '../../features/store/presentation/state/cartStore';

export function ProductCard({ product }: { product: Product }) {
  const add = useCartStore(state => state.add);
  const busy = useCartStore(state => state.busy);
  const onAdd = async () => {
    const success = await add(product);
    if (!success) {
      Alert.alert('No se pudo agregar', useCartStore.getState().error);
    }
  };
  const emoji =
    product.category === 'Tecnología'
      ? '🎧'
      : product.category === 'Hogar'
      ? '🏠'
      : product.category === 'Alimentos'
      ? '☕'
      : '📦';

  return (
    <View style={styles.card}>
      <View style={styles.art}>
        <Text style={styles.emoji}>{emoji}</Text>
      </View>
      <Text numberOfLines={1} style={styles.name}>
        {product.name}
      </Text>
      <Text numberOfLines={1} style={styles.business}>
        {product.businessName}
      </Text>
      <Text numberOfLines={2} style={styles.description}>
        {product.description}
      </Text>
      <View style={styles.footer}>
        <Text style={styles.price}>S/ {product.price.toFixed(2)}</Text>
        <Pressable disabled={busy} onPress={onAdd} style={styles.add}>
          <Text style={styles.addText}>＋</Text>
        </Pressable>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  card: {
    flex: 1,
    minHeight: 260,
    padding: 14,
    borderRadius: 18,
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
  },
  art: {
    height: 78,
    borderRadius: 14,
    backgroundColor: colors.primarySoft,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 12,
  },
  emoji: { fontSize: 38 },
  name: { fontSize: 16, fontWeight: '800', color: colors.text },
  business: { fontSize: 12, color: colors.muted, marginTop: 3 },
  description: {
    fontSize: 12,
    color: colors.muted,
    marginTop: 8,
    lineHeight: 17,
  },
  footer: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'flex-end',
    marginTop: 10,
  },
  price: { flex: 1, color: colors.primary, fontWeight: '800', fontSize: 15 },
  add: {
    width: 38,
    height: 38,
    borderRadius: 12,
    backgroundColor: colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
  },
  addText: { color: '#FFF', fontSize: 22, fontWeight: '700' },
});
