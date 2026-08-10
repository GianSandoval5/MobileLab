import React, { useEffect } from 'react';
import { FlatList, Pressable, StyleSheet, Text, View } from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { colors } from '../../../../core/theme/colors';
import {
  ErrorView,
  LoadingView,
} from '../../../../presentation/components/StateView';
import { Business } from '../../domain/entities/storeEntities';
import { useCatalogStore } from '../state/catalogStore';

const Separator = () => <View style={styles.separator} />;

export function BusinessesScreen() {
  const navigation = useNavigation<any>();
  const { businesses, loadingBusinesses, businessesError, loadBusinesses } =
    useCatalogStore();

  useEffect(() => {
    if (!businesses.length) {
      loadBusinesses();
    }
  }, [businesses.length, loadBusinesses]);

  if (loadingBusinesses && !businesses.length) {
    return <LoadingView />;
  }
  if (businessesError && !businesses.length) {
    return <ErrorView error={businessesError} retry={loadBusinesses} />;
  }

  const renderBusiness = ({ item }: { item: Business }) => (
    <Pressable
      style={styles.card}
      onPress={() => navigation.navigate('BusinessDetail', { business: item })}
    >
      <View style={styles.logo}>
        <Text style={styles.logoText}>🏪</Text>
      </View>
      <View style={styles.info}>
        <Text style={styles.name}>{item.name}</Text>
        <Text numberOfLines={2} style={styles.description}>
          {item.description}
        </Text>
        <Text style={styles.meta}>
          ⭐ {item.rating} · {item.category}
        </Text>
      </View>
      <Text style={styles.chevron}>›</Text>
    </Pressable>
  );

  return (
    <FlatList
      data={businesses}
      keyExtractor={item => item.id}
      renderItem={renderBusiness}
      contentContainerStyle={styles.content}
      ItemSeparatorComponent={Separator}
      refreshing={loadingBusinesses}
      onRefresh={loadBusinesses}
    />
  );
}

const styles = StyleSheet.create({
  content: { padding: 14, paddingBottom: 28 },
  separator: { height: 12 },
  card: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#FFF',
    borderRadius: 20,
    padding: 16,
    borderWidth: 1,
    borderColor: colors.border,
  },
  logo: {
    width: 60,
    height: 60,
    borderRadius: 17,
    backgroundColor: colors.primarySoft,
    alignItems: 'center',
    justifyContent: 'center',
  },
  logoText: { fontSize: 30 },
  info: { flex: 1, marginLeft: 14 },
  name: { fontSize: 17, fontWeight: '800', color: colors.text },
  description: { fontSize: 13, color: colors.muted, marginTop: 4 },
  meta: { fontSize: 13, color: colors.text, marginTop: 8 },
  chevron: { fontSize: 30, color: colors.muted },
});
