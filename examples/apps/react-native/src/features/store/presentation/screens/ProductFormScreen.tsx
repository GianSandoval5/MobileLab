import React, { useState } from 'react';
import { Alert, ScrollView, StyleSheet, Text, View } from 'react-native';
import { NativeStackScreenProps } from '@react-navigation/native-stack';
import { colors } from '../../../../core/theme/colors';
import { AppButton } from '../../../../presentation/components/AppButton';
import { AppField } from '../../../../presentation/components/AppField';
import { RootStackParamList } from '../../../../presentation/navigation/types';
import { Product } from '../../domain/entities/storeEntities';
import { useMyProductsStore } from '../state/myProductsStore';

type Props = NativeStackScreenProps<RootStackParamList, 'ProductForm'>;

export function ProductFormScreen({ navigation, route }: Props) {
  const product = route.params?.product;
  const isNew = !product;
  const { loading, error, save } = useMyProductsStore();
  const [name, setName] = useState(product?.name ?? '');
  const [description, setDescription] = useState(product?.description ?? '');
  const [price, setPrice] = useState(product?.price.toFixed(2) ?? '');
  const [stock, setStock] = useState(product?.stock.toString() ?? '');
  const [category, setCategory] = useState(product?.category ?? '');
  const [validation, setValidation] = useState('');

  const submit = async () => {
    const numericPrice = Number(price.replace(',', '.'));
    const numericStock = Number(stock);
    if (
      !name.trim() ||
      !description.trim() ||
      !category.trim() ||
      numericPrice <= 0 ||
      numericStock < 0
    ) {
      setValidation('Completa todos los campos con valores válidos.');
      return;
    }
    const value: Product = {
      id: product?.id ?? 'pending',
      name: name.trim(),
      description: description.trim(),
      price: numericPrice,
      stock: numericStock,
      category: category.trim(),
      businessId: 'my-business',
      businessName: 'Mi negocio',
    };
    if (await save(value, isNew)) {
      Alert.alert(
        isNew ? 'Producto publicado' : 'Producto actualizado',
        'Los cambios fueron guardados.',
        [{ text: 'Continuar', onPress: () => navigation.goBack() }],
      );
    }
  };

  return (
    <ScrollView
      contentContainerStyle={styles.content}
      keyboardShouldPersistTaps="handled"
    >
      <AppField label="Nombre" value={name} onChangeText={setName} />
      <AppField
        label="Descripción"
        value={description}
        onChangeText={setDescription}
        multiline
        style={styles.description}
      />
      <View style={styles.row}>
        <View style={styles.half}>
          <AppField
            label="Precio (S/)"
            value={price}
            onChangeText={setPrice}
            keyboardType="decimal-pad"
          />
        </View>
        <View style={styles.half}>
          <AppField
            label="Stock"
            value={stock}
            onChangeText={setStock}
            keyboardType="number-pad"
          />
        </View>
      </View>
      <AppField label="Categoría" value={category} onChangeText={setCategory} />
      {!!(validation || error) && (
        <Text style={styles.error}>{validation || error}</Text>
      )}
      <AppButton
        title={isNew ? 'Publicar' : 'Guardar cambios'}
        onPress={submit}
        loading={loading}
      />
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
  description: { height: 100, textAlignVertical: 'top', paddingTop: 14 },
  row: { flexDirection: 'row', gap: 12 },
  half: { flex: 1 },
  error: { color: colors.error, textAlign: 'center' },
});
