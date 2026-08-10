import React, { useState } from 'react';
import { Alert, ScrollView, StyleSheet, Text } from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { colors } from '../../../../core/theme/colors';
import { AppButton } from '../../../../presentation/components/AppButton';
import { AppField } from '../../../../presentation/components/AppField';
import { useAuthStore } from '../../../auth/presentation/state/authStore';

export function EditProfileScreen() {
  const navigation = useNavigation<any>();
  const user = useAuthStore(state => state.user)!;
  const { loading, error, updateProfile } = useAuthStore();
  const [name, setName] = useState(user.name);
  const [email, setEmail] = useState(user.email);
  const [phone, setPhone] = useState(user.phone);
  const [validation, setValidation] = useState('');

  const save = async () => {
    if (!name.trim() || !email.includes('@')) {
      setValidation('Ingresa un nombre y correo válidos.');
      return;
    }
    if (await updateProfile(name.trim(), email.trim(), phone.trim())) {
      Alert.alert('Perfil actualizado', 'Tus cambios fueron guardados.', [
        { text: 'Continuar', onPress: () => navigation.goBack() },
      ]);
    }
  };

  return (
    <ScrollView
      contentContainerStyle={styles.content}
      keyboardShouldPersistTaps="handled"
    >
      <AppField label="Nombre" value={name} onChangeText={setName} />
      <AppField
        label="Correo"
        value={email}
        onChangeText={setEmail}
        autoCapitalize="none"
        keyboardType="email-address"
      />
      <AppField
        label="Teléfono"
        value={phone}
        onChangeText={setPhone}
        keyboardType="phone-pad"
      />
      {!!(validation || error) && (
        <Text style={styles.error}>{validation || error}</Text>
      )}
      <AppButton title="Guardar cambios" onPress={save} loading={loading} />
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  content: {
    padding: 18,
    gap: 14,
    backgroundColor: colors.background,
    flexGrow: 1,
  },
  error: { color: colors.error, textAlign: 'center' },
});
