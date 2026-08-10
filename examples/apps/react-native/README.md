# MobileLab React Native Shop Example

Aplicación de comercio electrónico en React Native CLI y TypeScript conectada
a una API simulada con MobileLab.

## Funcionalidades

- Inicio de sesión y registro.
- Catálogo general de productos.
- Negocios y productos por negocio.
- Carrito con cantidades, eliminación y total.
- Checkout y pago simulado.
- Historial y detalle de compras.
- Consulta y edición de perfil.
- Consulta, creación y edición de productos propios.
- Estados de carga, error y reintento.

## Arquitectura

El proyecto usa Clean Architecture organizada por funcionalidades:

```text
src/
├── core/
│   ├── config/       # Dirección de MobileLab
│   ├── di/           # Inyección de dependencias
│   ├── errors/       # Errores de aplicación
│   ├── network/      # Cliente HTTP
│   └── theme/        # Colores compartidos
├── features/
│   ├── auth/
│   │   ├── data/
│   │   ├── domain/
│   │   └── presentation/
│   └── store/
│       ├── data/
│       ├── domain/
│       └── presentation/
└── presentation/
    ├── components/
    └── navigation/
```

- React Navigation administra stacks y pestañas.
- Zustand administra autenticación, catálogo, carrito, compras y productos.
- AsyncStorage conserva sesión, perfil, carrito, compras y productos en el dispositivo.
- Las pantallas dependen de stores y casos de uso, no directamente de `fetch`.
- Los repositorios abstraen MobileLab para poder sustituirlo por un backend real.

## Ejecutar

Desde `examples/apps/react-native`, prepara las dependencias JavaScript:

Este ejemplo usa npm y `package-lock.json` como fuente reproducible; no mezcles
el lockfile con Yarn.

```bash
npm ci
```

En la primera terminal, valida la configuración incluida y levanta MobileLab:

```bash
../../../bin/mobilelab doctor
../../../bin/mobilelab start
```

No es necesario ejecutar `mobilelab init`: este ejemplo ya incluye
`mobilelab.yaml`, fixtures y escenarios. MobileLab permanece en primer plano;
usa Ctrl+C para detenerlo.

En una segunda terminal inicia Metro:

```bash
npm start
```

En una tercera terminal ejecuta Android:

```bash
npm run android
```

Para iOS, después de instalar dependencias nativas:

```bash
cd ios
bundle exec pod install
cd ..
npm run ios
```

Con MobileLab activo puedes abrir:

- Información de la API: `http://127.0.0.1:4566/`
- Productos de ejemplo: `http://127.0.0.1:4566/api/products`
- Dashboard: `http://127.0.0.1:4566/dashboard`

Credenciales precargadas:

```text
Correo: demo@mobilelab.dev
Contraseña: 123456
```

La URL se define en `src/core/config/apiConfig.ts`:

- Android Emulator: `http://10.0.2.2:4566`
- iOS Simulator: `http://127.0.0.1:4566`

Para un dispositivo físico reemplázala por la IP local de la computadora y
configura MobileLab para escuchar en una interfaz accesible desde la red.

El mensaje `no mock configured` significa que se solicitó una ruta que no está
declarada en `mobilelab.yaml`; no indica que MobileLab se haya detenido.

## MobileLab

Los endpoints se encuentran en `mobilelab.yaml` y sus respuestas en
`mobilelab/fixtures/`. Las respuestas son estáticas, pero la app fusiona las
fixtures con los cambios persistidos en AsyncStorage. Estos datos sobreviven a
recargas y reinicios en el mismo dispositivo; no se sincronizan entre equipos.

## Verificar

```bash
npm ci
npx tsc --noEmit
npm run lint
npm test -- --runInBand --no-watchman
```
