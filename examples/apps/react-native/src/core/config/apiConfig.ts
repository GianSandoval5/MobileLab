import { Platform } from 'react-native';

// Para un dispositivo físico reemplaza esta URL por la IP local de tu Mac/PC.
export const API_BASE_URL = Platform.select({
  android: 'http://10.0.2.2:4566',
  default: 'http://127.0.0.1:4566',
}) as string;
