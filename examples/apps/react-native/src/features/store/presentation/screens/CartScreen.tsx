import React from 'react';
import { FlatList, Pressable, StyleSheet, Text, View } from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { colors } from '../../../../core/theme/colors';
import { AppButton } from '../../../../presentation/components/AppButton';
import { CartItem, cartSubtotal } from '../../domain/entities/storeEntities';
import { cartTotal, useCartStore } from '../state/cartStore';

const Separator = () => <View style={styles.separator} />;

export function CartScreen() {
  const navigation = useNavigation<any>();
  const { items, busy, changeQuantity, remove } = useCartStore();
  const total = cartTotal(items);

  if (!items.length) {
    return (
      <View style={styles.empty}>
        <Text style={styles.emptyIcon}>🛍️</Text>
        <Text style={styles.emptyTitle}>Tu carrito está vacío</Text>
        <Text style={styles.muted}>Agrega productos para comenzar.</Text>
      </View>
    );
  }

  const renderItem = ({ item }: { item: CartItem }) => (
    <View style={styles.card}>
      <View style={styles.productIcon}>
        <Text style={styles.productEmoji}>📦</Text>
      </View>
      <View style={styles.productInfo}>
        <Text style={styles.name}>{item.product.name}</Text>
        <Text style={styles.muted}>S/ {item.product.price.toFixed(2)} c/u</Text>
        <View style={styles.quantityRow}>
          <Pressable
            disabled={busy}
            onPress={() => changeQuantity(item, item.quantity - 1)}
            style={styles.quantityButton}
          >
            <Text>−</Text>
          </Pressable>
          <Text style={styles.quantity}>{item.quantity}</Text>
          <Pressable
            disabled={busy || item.quantity >= item.product.stock}
            onPress={() => changeQuantity(item, item.quantity + 1)}
            style={styles.quantityButton}
          >
            <Text>＋</Text>
          </Pressable>
        </View>
      </View>
      <View style={styles.right}>
        <Pressable disabled={busy} onPress={() => remove(item.product.id)}>
          <Text style={styles.delete}>Eliminar</Text>
        </Pressable>
        <Text style={styles.subtotal}>S/ {cartSubtotal(item).toFixed(2)}</Text>
      </View>
    </View>
  );

  return (
    <View style={styles.container}>
      {busy && <View style={styles.progress} />}
      <FlatList
        data={items}
        keyExtractor={item => item.product.id}
        renderItem={renderItem}
        contentContainerStyle={styles.list}
        ItemSeparatorComponent={Separator}
      />
      <View style={styles.summary}>
        <View style={styles.totalRow}>
          <Text style={styles.totalLabel}>Total</Text>
          <Text style={styles.total}>S/ {total.toFixed(2)}</Text>
        </View>
        <AppButton
          title="Continuar al pago"
          disabled={busy}
          onPress={() => navigation.navigate('Checkout')}
        />
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  progress: { height: 3, backgroundColor: colors.primary },
  list: { padding: 14 },
  separator: { height: 10 },
  card: {
    flexDirection: 'row',
    padding: 14,
    backgroundColor: '#FFF',
    borderRadius: 18,
    borderWidth: 1,
    borderColor: colors.border,
  },
  productIcon: {
    width: 56,
    height: 56,
    borderRadius: 14,
    backgroundColor: colors.primarySoft,
    alignItems: 'center',
    justifyContent: 'center',
  },
  productEmoji: { fontSize: 26 },
  productInfo: { flex: 1, marginLeft: 12 },
  name: { fontWeight: '800', color: colors.text },
  muted: { color: colors.muted, marginTop: 4 },
  quantityRow: { flexDirection: 'row', alignItems: 'center', marginTop: 10 },
  quantityButton: {
    width: 30,
    height: 30,
    borderRadius: 10,
    borderWidth: 1,
    borderColor: colors.border,
    alignItems: 'center',
    justifyContent: 'center',
  },
  quantity: { marginHorizontal: 12, fontWeight: '800' },
  right: { alignItems: 'flex-end', justifyContent: 'space-between' },
  delete: { fontSize: 12, color: colors.error },
  subtotal: { fontWeight: '800', color: colors.text },
  summary: {
    padding: 18,
    backgroundColor: '#FFF',
    borderTopWidth: 1,
    borderColor: colors.border,
    gap: 14,
  },
  totalRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  totalLabel: { fontSize: 20, color: colors.text },
  total: { fontSize: 25, fontWeight: '900', color: colors.text },
  empty: { flex: 1, alignItems: 'center', justifyContent: 'center', gap: 10 },
  emptyIcon: { fontSize: 62 },
  emptyTitle: { fontSize: 21, fontWeight: '800', color: colors.text },
});
