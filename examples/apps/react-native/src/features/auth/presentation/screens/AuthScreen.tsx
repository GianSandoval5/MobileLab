import React, { useState } from 'react';
import {
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from 'react-native';
import { API_BASE_URL } from '../../../../core/config/apiConfig';
import { colors } from '../../../../core/theme/colors';
import { AppButton } from '../../../../presentation/components/AppButton';
import { AppField } from '../../../../presentation/components/AppField';
import { useAuthStore } from '../state/authStore';

export function AuthScreen() {
  const [registerMode, setRegisterMode] = useState(false);
  const [name, setName] = useState('');
  const [email, setEmail] = useState('demo@mobilelab.dev');
  const [password, setPassword] = useState('123456');
  const [validation, setValidation] = useState('');
  const { loading, error, login, register, clearError } = useAuthStore();

  const submit = async () => {
    if (registerMode && !name.trim()) {
      setValidation('Ingresa tu nombre.');
      return;
    }
    if (!email.includes('@')) {
      setValidation('Ingresa un correo válido.');
      return;
    }
    if (password.length < 6) {
      setValidation('La contraseña debe tener al menos 6 caracteres.');
      return;
    }
    setValidation('');
    registerMode
      ? await register(name.trim(), email.trim(), password)
      : await login(email.trim(), password);
  };

  return (
    <KeyboardAvoidingView
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
      style={styles.flex}
    >
      <ScrollView
        contentContainerStyle={styles.container}
        keyboardShouldPersistTaps="handled"
      >
        <View style={styles.card}>
          <View style={styles.logo}>
            <Text style={styles.logoText}>🛍️</Text>
          </View>
          <Text style={styles.title}>
            {registerMode ? 'Crea tu cuenta' : 'Bienvenido'}
          </Text>
          <Text style={styles.subtitle}>
            {registerMode
              ? 'Regístrate para comprar y vender.'
              : 'Ingresa a MobileLab Shop.'}
          </Text>
          <View style={styles.form}>
            {registerMode && (
              <AppField label="Nombre" value={name} onChangeText={setName} />
            )}
            <AppField
              label="Correo"
              value={email}
              onChangeText={setEmail}
              autoCapitalize="none"
              keyboardType="email-address"
            />
            <AppField
              label="Contraseña"
              value={password}
              onChangeText={setPassword}
              secureTextEntry
            />
            {!!(validation || error) && (
              <Text style={styles.error}>{validation || error}</Text>
            )}
            <AppButton
              title={registerMode ? 'Registrarme' : 'Ingresar'}
              onPress={submit}
              loading={loading}
            />
            <AppButton
              title={registerMode ? 'Ya tengo cuenta' : 'Crear una cuenta'}
              variant="text"
              disabled={loading}
              onPress={() => {
                clearError();
                setValidation('');
                setRegisterMode(value => !value);
              }}
            />
          </View>
          <Text style={styles.api}>API: {API_BASE_URL}</Text>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  flex: { flex: 1, backgroundColor: colors.background },
  container: { flexGrow: 1, justifyContent: 'center', padding: 22 },
  card: {
    backgroundColor: '#FFF',
    borderRadius: 24,
    padding: 26,
    borderWidth: 1,
    borderColor: colors.border,
  },
  logo: {
    alignSelf: 'center',
    width: 66,
    height: 66,
    borderRadius: 20,
    backgroundColor: colors.primarySoft,
    alignItems: 'center',
    justifyContent: 'center',
  },
  logoText: { fontSize: 34 },
  title: {
    fontSize: 30,
    fontWeight: '800',
    color: colors.text,
    textAlign: 'center',
    marginTop: 18,
  },
  subtitle: { color: colors.muted, textAlign: 'center', marginTop: 6 },
  form: { gap: 14, marginTop: 26 },
  error: { color: colors.error, textAlign: 'center' },
  api: {
    color: colors.muted,
    fontSize: 11,
    textAlign: 'center',
    marginTop: 16,
  },
});
