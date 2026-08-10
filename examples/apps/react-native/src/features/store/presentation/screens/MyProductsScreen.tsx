import React, { useEffect } from 'react';
import { FlatList, Pressable, StyleSheet, Text, View } from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { colors } from '../../../../core/theme/colors';
import { AppButton } from '../../../../presentation/components/AppButton';
import {
  ErrorView,
  LoadingView,
} from '../../../../presentation/components/StateView';
import { Product } from '../../domain/entities/storeEntities';
import { useMyProductsStore } from '../state/myProductsStore';

const Separator = () => <View style={styles.separator} />;

export function MyProductsScreen() {
  const navigation = useNavigation<any>();
  const { products, loading, error, load } = useMyProductsStore();
  useEffect(() => {
    if (!products.length) {
      load();
    }
  }, [load, products.length]);

  if (loading && !products.length) {
    return <LoadingView />;
  }
  if (error && !products.length) {
    return <ErrorView error={error} retry={load} />;
  }

  const renderItem = ({ item }: { item: Product }) => (
    <Pressable
      style={styles.card}
      onPress={() => navigation.navigate('ProductForm', { product: item })}
    >
      <View style={styles.avatar}>
        <Text style={styles.avatarText}>{item.name[0]}</Text>
      </View>
      <View style={styles.info}>
        <Text style={styles.name}>{item.name}</Text>
        <Text style={styles.meta}>
          S/ {item.price.toFixed(2)} · Stock {item.stock}
        </Text>
      </View>
      <Text style={styles.edit}>Editar</Text>
    </Pressable>
  );

  return (
    <View style={styles.container}>
      <FlatList
        data={products}
        keyExtractor={item => item.id}
        renderItem={renderItem}
        contentContainerStyle={styles.list}
        ItemSeparatorComponent={Separator}
        refreshing={loading}
        onRefresh={load}
      />
      <View style={styles.footer}>
        <AppButton
          title="＋ Publicar producto"
          onPress={() => navigation.navigate('ProductForm')}
        />
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  list: { padding: 14, paddingBottom: 90 },
  separator: { height: 10 },
  card: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 14,
    backgroundColor: '#FFF',
    borderRadius: 18,
    borderWidth: 1,
    borderColor: colors.border,
  },
  avatar: {
    width: 50,
    height: 50,
    borderRadius: 14,
    backgroundColor: colors.primarySoft,
    alignItems: 'center',
    justifyContent: 'center',
  },
  avatarText: { fontSize: 20, fontWeight: '800', color: colors.primary },
  info: { flex: 1, marginLeft: 12 },
  name: { fontWeight: '800', color: colors.text },
  meta: { color: colors.muted, marginTop: 4 },
  edit: { color: colors.primary, fontWeight: '700' },
  footer: { position: 'absolute', left: 14, right: 14, bottom: 14 },
});
