# MobileLab Flutter Shop Example

Demo de comercio electrónico en Flutter conectada a una API simulada con
MobileLab. Incluye autenticación, catálogo, negocios, carrito, pago, historial
de compras, edición de perfil y administración de productos propios.

## Funcionalidades

- Inicio de sesión y registro.
- Catálogo general de productos.
- Lista de negocios y productos de cada negocio.
- Carrito con cantidades, eliminación y total.
- Checkout y pago simulado con tarjeta.
- Historial y detalle de compras.
- Consulta y edición de perfil.
- Consulta, creación y edición de productos propios.
- Estados de carga, errores y reintentos.

## Arquitectura

El código usa Clean Architecture organizada por funcionalidades:

```text
lib/
├── core/
│   ├── config/       # URL de la API
│   ├── errors/       # Errores de aplicación
│   ├── network/      # Cliente HTTP
│   ├── providers/    # Inyección de dependencias
│   └── theme/        # Tema visual
└── features/
    ├── auth/
    │   ├── data/
    │   ├── domain/
    │   └── presentation/
    └── store/
        ├── data/
        ├── domain/
        └── presentation/
```

Riverpod administra el estado y la inyección de dependencias. La presentación
depende de casos de uso y repositorios; no realiza peticiones HTTP directamente.

## API de MobileLab

La configuración se encuentra en `mobilelab.yaml` y las respuestas en
`mobilelab/fixtures/`.

| Método | Ruta | Uso |
| --- | --- | --- |
| POST | `/api/auth/login` | Iniciar sesión |
| POST | `/api/auth/register` | Registrar usuario |
| GET, PUT | `/api/profile` | Consultar y editar perfil |
| GET | `/api/products` | Catálogo |
| GET | `/api/businesses` | Negocios |
| GET | `/api/businesses/{id}/products` | Productos del negocio |
| POST, PATCH, DELETE | `/api/cart/items` | Administrar carrito |
| POST | `/api/payments` | Procesar pago simulado |
| GET | `/api/purchases` | Historial de compras |
| GET, POST | `/api/my-products` | Listar y crear productos propios |
| PUT | `/api/my-products/{id}` | Editar un producto propio |

MobileLab devuelve fixtures estáticas. La app conserva localmente los datos
escritos durante la sesión para reflejar de inmediato cambios de carrito,
perfil, compras y productos.

## Ejecutar

Desde `examples/apps/flutter`, instala primero las dependencias:

```bash
flutter pub get
```

En la primera terminal, valida la configuración incluida y levanta MobileLab:

```bash
../../../bin/mobilelab doctor
../../../bin/mobilelab start
```

No es necesario ejecutar `mobilelab init`: este ejemplo ya incluye
`mobilelab.yaml`, fixtures y escenarios. MobileLab permanece en primer plano;
usa Ctrl+C para detenerlo.

En una segunda terminal ejecuta la aplicación:

```bash
flutter run
```

Puedes ingresar con los valores precargados:

- Correo: `demo@mobilelab.dev`
- Contraseña: `123456`

La URL se selecciona automáticamente:

- Android Emulator: `http://10.0.2.2:4566`
- iOS Simulator, web y escritorio: `http://127.0.0.1:4566`

Con MobileLab activo puedes abrir:

- Información de la API: `http://127.0.0.1:4566/`
- Productos de ejemplo: `http://127.0.0.1:4566/api/products`
- Dashboard: `http://127.0.0.1:4566/dashboard`

En un dispositivo físico indica la IP local de la computadora:

```bash
flutter run --dart-define=API_BASE_URL=http://192.168.1.10:4566
```

MobileLab deberá escuchar en una interfaz accesible desde la red local y el
teléfono deberá estar conectado a la misma red.

El mensaje `no mock configured` significa que se solicitó una ruta que no está
declarada en `mobilelab.yaml`; no indica que MobileLab se haya detenido.

## Verificar

```bash
flutter pub get
flutter analyze
flutter test
```
