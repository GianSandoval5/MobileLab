import React from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { useNavigation } from '@react-navigation/native';
import {
  cartCount,
  useCartStore,
} from '../../features/store/presentation/state/cartStore';
import { colors } from '../../core/theme/colors';

export function CartButton() {
  const navigation = useNavigation<any>();
  const count = useCartStore(state => cartCount(state.items));
  return (
    <Pressable
      onPress={() => navigation.navigate('Cart')}
      style={styles.button}
    >
      <Text style={styles.icon}>🛍️</Text>
      {count > 0 && (
        <View style={styles.badge}>
          <Text style={styles.badgeText}>{count}</Text>
        </View>
      )}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  button: { padding: 8, marginRight: 8 },
  icon: { fontSize: 22 },
  badge: {
    position: 'absolute',
    top: 0,
    right: 0,
    minWidth: 18,
    height: 18,
    borderRadius: 9,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.error,
  },
  badgeText: { color: '#FFF', fontSize: 11, fontWeight: '800' },
});
