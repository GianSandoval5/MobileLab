import React, { useState } from 'react';
import {
  Alert,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { colors } from '../../../../core/theme/colors';
import { AppButton } from '../../../../presentation/components/AppButton';
import { AppField } from '../../../../presentation/components/AppField';
import { cartCount, cartTotal, useCartStore } from '../state/cartStore';
import { usePurchasesStore } from '../state/purchasesStore';

export function CheckoutScreen() {
  const navigation = useNavigation<any>();
  const { items, busy, error, checkout, completeCheckout } = useCartStore();
  const addPurchase = usePurchasesStore(state => state.add);
  const [name, setName] = useState('MobileLab User');
  const [card, setCard] = useState('4242424242424242');
  const [expiry, setExpiry] = useState('12/30');
  const [cvv, setCvv] = useState('123');
  const [validation, setValidation] = useState('');
  const total = cartTotal(items);

  const pay = async () => {
    if (!name.trim() || card.length < 16 || !expiry.trim() || cvv.length < 3) {
      setValidation('Completa correctamente los datos de la tarjeta.');
      return;
    }
    setValidation('');
    const purchase = await checkout();
    if (!purchase) {
      Alert.alert('No se pudo procesar el pago', useCartStore.getState().error);
      return;
    }
    addPurchase(purchase);
    Alert.alert(
      '¡Pago completado!',
      `La compra ${purchase.id} fue registrada correctamente.`,
      [
        {
          text: 'Continuar',
          onPress: () => {
            completeCheckout();
            navigation.popToTop();
          },
        },
      ],
      { cancelable: false },
    );
  };

  return (
    <KeyboardAvoidingView
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
      style={styles.flex}
    >
      <ScrollView
        contentContainerStyle={styles.content}
        keyboardShouldPersistTaps="handled"
      >
        <View style={styles.summary}>
          <Text>🛍️ {cartCount(items)} artículos</Text>
          <Text style={styles.summaryTotal}>S/ {total.toFixed(2)}</Text>
        </View>
        <Text style={styles.title}>Datos de la tarjeta</Text>
        <View style={styles.form}>
          <AppField
            label="Nombre del titular"
            value={name}
            onChangeText={setName}
          />
          <AppField
            label="Número de tarjeta"
            value={card}
            onChangeText={value =>
              setCard(value.replace(/\D/g, '').slice(0, 16))
            }
            keyboardType="number-pad"
          />
          <View style={styles.row}>
            <View style={styles.half}>
              <AppField
                label="MM/AA"
                value={expiry}
                onChangeText={setExpiry}
                keyboardType="numbers-and-punctuation"
              />
            </View>
            <View style={styles.half}>
              <AppField
                label="CVV"
                value={cvv}
                onChangeText={value =>
                  setCvv(value.replace(/\D/g, '').slice(0, 4))
                }
                keyboardType="number-pad"
                secureTextEntry
              />
            </View>
          </View>
          <Text style={styles.note}>
            Pago simulado: MobileLab no procesa datos bancarios reales.
          </Text>
          {!!(validation || error) && (
            <Text style={styles.error}>{validation || error}</Text>
          )}
          <AppButton
            title={`Pagar S/ ${total.toFixed(2)}`}
            onPress={pay}
            loading={busy}
          />
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  flex: { flex: 1, backgroundColor: colors.background },
  content: { padding: 18 },
  summary: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    padding: 18,
    borderRadius: 18,
    backgroundColor: colors.primarySoft,
  },
  summaryTotal: { fontWeight: '900' },
  title: { fontSize: 23, fontWeight: '800', color: colors.text, marginTop: 26 },
  form: { gap: 14, marginTop: 16 },
  row: { flexDirection: 'row', gap: 12 },
  half: { flex: 1 },
  note: { fontSize: 12, color: colors.muted },
  error: { color: colors.error, textAlign: 'center' },
});
