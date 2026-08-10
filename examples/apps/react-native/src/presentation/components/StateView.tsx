import React from 'react';
import { ActivityIndicator, StyleSheet, Text, View } from 'react-native';
import { colors } from '../../core/theme/colors';
import { AppButton } from './AppButton';

export function LoadingView() {
  return (
    <View style={styles.center}>
      <ActivityIndicator size="large" color={colors.primary} />
      <Text style={styles.muted}>Cargando…</Text>
    </View>
  );
}

export function ErrorView({
  error,
  retry,
}: {
  error: string;
  retry: () => void;
}) {
  return (
    <View style={styles.center}>
      <Text style={styles.icon}>☁️</Text>
      <Text style={styles.title}>No pudimos cargar la información</Text>
      <Text style={styles.muted}>{error}</Text>
      <AppButton title="Reintentar" variant="outline" onPress={retry} />
    </View>
  );
}

const styles = StyleSheet.create({
  center: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    padding: 28,
    gap: 12,
  },
  icon: { fontSize: 46 },
  title: {
    fontSize: 18,
    fontWeight: '700',
    color: colors.text,
    textAlign: 'center',
  },
  muted: { color: colors.muted, textAlign: 'center' },
});
