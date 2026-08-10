import React from 'react';
import { Alert, Pressable, StyleSheet, Text, View } from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { colors } from '../../../../core/theme/colors';
import { useAuthStore } from '../../../auth/presentation/state/authStore';
import { useCartStore } from '../state/cartStore';
import { useMyProductsStore } from '../state/myProductsStore';
import { usePurchasesStore } from '../state/purchasesStore';

export function ProfileScreen() {
  const navigation = useNavigation<any>();
  const user = useAuthStore(state => state.user);
  const logout = useAuthStore(state => state.logout);
  if (!user) {
    return null;
  }

  const confirmLogout = () => {
    Alert.alert('Cerrar sesión', '¿Quieres salir de tu cuenta?', [
      { text: 'Cancelar', style: 'cancel' },
      {
        text: 'Salir',
        style: 'destructive',
        onPress: () => {
          useCartStore.getState().reset();
          usePurchasesStore.getState().reset();
          useMyProductsStore.getState().reset();
          logout();
        },
      },
    ]);
  };

  return (
    <View style={styles.container}>
      <View style={styles.profileCard}>
        <View style={styles.avatar}>
          <Text style={styles.avatarText}>{user.name[0].toUpperCase()}</Text>
        </View>
        <Text style={styles.name}>{user.name}</Text>
        <Text style={styles.meta}>{user.email}</Text>
        {!!user.phone && <Text style={styles.meta}>{user.phone}</Text>}
      </View>
      <View style={styles.menu}>
        <MenuItem
          emoji="✏️"
          title="Editar perfil"
          subtitle="Nombre, correo y teléfono"
          onPress={() => navigation.navigate('EditProfile')}
        />
        <View style={styles.divider} />
        <MenuItem
          emoji="📦"
          title="Mis productos"
          subtitle="Publica y administra lo que vendes"
          onPress={() => navigation.navigate('MyProducts')}
        />
      </View>
      <Pressable style={styles.logout} onPress={confirmLogout}>
        <Text style={styles.logoutText}>Cerrar sesión</Text>
      </Pressable>
    </View>
  );
}

function MenuItem({
  emoji,
  title,
  subtitle,
  onPress,
}: {
  emoji: string;
  title: string;
  subtitle: string;
  onPress: () => void;
}) {
  return (
    <Pressable onPress={onPress} style={styles.menuItem}>
      <Text style={styles.menuEmoji}>{emoji}</Text>
      <View style={styles.menuInfo}>
        <Text style={styles.menuTitle}>{title}</Text>
        <Text style={styles.meta}>{subtitle}</Text>
      </View>
      <Text style={styles.chevron}>›</Text>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 14,
    backgroundColor: colors.background,
    gap: 14,
  },
  profileCard: {
    alignItems: 'center',
    padding: 24,
    borderRadius: 20,
    backgroundColor: colors.primarySoft,
  },
  avatar: {
    width: 76,
    height: 76,
    borderRadius: 38,
    backgroundColor: '#FFF',
    alignItems: 'center',
    justifyContent: 'center',
  },
  avatarText: { fontSize: 30, fontWeight: '800', color: colors.primary },
  name: { fontSize: 22, fontWeight: '900', color: colors.text, marginTop: 12 },
  meta: { color: colors.muted, marginTop: 3 },
  menu: {
    backgroundColor: '#FFF',
    borderRadius: 20,
    borderWidth: 1,
    borderColor: colors.border,
  },
  menuItem: { flexDirection: 'row', alignItems: 'center', padding: 16 },
  menuEmoji: { fontSize: 24 },
  menuInfo: { flex: 1, marginLeft: 12 },
  menuTitle: { fontWeight: '800', color: colors.text },
  chevron: { fontSize: 28, color: colors.muted },
  divider: { height: 1, backgroundColor: colors.border, marginHorizontal: 16 },
  logout: {
    minHeight: 48,
    borderRadius: 14,
    borderWidth: 1,
    borderColor: colors.error,
    alignItems: 'center',
    justifyContent: 'center',
  },
  logoutText: { color: colors.error, fontWeight: '800' },
});
