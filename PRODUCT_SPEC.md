Actúa como Principal Software Architect, Staff Engineer y especialista en Developer Tooling, Mobile Development, Go, Flutter, React Native, Android, iOS, Ionic/Capacitor, networking, testing y CLI development.

Quiero que diseñes e implementes desde cero un proyecto open source llamado provisionalmente:

MobileLab

============================================================
1. VISIÓN DEL PRODUCTO
============================================================

MobileLab será una plataforma local-first para desarrollo, simulación y pruebas de aplicaciones móviles.

Objetivo principal:

Permitir que cualquier desarrollador mobile pueda levantar en su computadora un entorno completo para desarrollar y probar aplicaciones sin depender inicialmente de:

- backend real
- ambientes cloud
- cuentas cloud
- tarjetas de crédito
- tokens externos
- infraestructura remota
- Docker obligatorio

Debe funcionar con:

- Flutter
- React Native
- Android / Kotlin
- iOS / Swift
- Ionic / Capacitor

IMPORTANTE:

MobileLab NO debe implementar cinco productos diferentes.

Debe existir un Core universal y extensible.

Las diferencias entre plataformas deben resolverse mediante adapters/drivers/plugins.

Arquitectura conceptual:

                         MobileLab
                             |
          +------------------+------------------+
          |                  |                  |
      API Sandbox       Scenario Engine      Device Engine
          |                  |                  |
       REST              Scenarios           Android
       Auth              Assertions          iOS
       JWT               Replay              Network
       Errors            Reports             Location
       Latency                               Deep Links
       Fixtures                              App Lifecycle
       WebSocket                             Push
          |
          +------------------+
                             |
                     Platform Adapters
                             |
          +---------+--------+--------+---------+
          |         |        |        |         |
       Flutter     RN      Kotlin    Swift    Capacitor


Principio fundamental:

"Write the mobile scenario once. Run it everywhere."

MobileLab debe ser:

- local-first
- cross-platform
- framework-agnostic
- modular
- extensible
- rápido
- developer-first
- CLI-first
- automatizable
- CI-friendly
- open source friendly

============================================================
2. DECISIÓN TECNOLÓGICA
============================================================

Analiza primero la arquitectura y luego implementa.

Para el Core y CLI, prioriza Go salvo que exista una razón técnica fuerte para utilizar Rust.

Mi preferencia inicial es:

Core / CLI:
Go

Persistencia:
SQLite

Configuración:
YAML

API:
HTTP

Realtime:
WebSocket

Dashboard:
React + TypeScript

Flutter integration:
Dart package

React Native integration:
TypeScript package

Capacitor integration:
Capacitor plugin

Android:
Kotlin adapter/SDK

iOS:
Swift adapter/SDK

No introduzcas microservicios innecesarios.

MobileLab inicialmente debe poder distribuirse como un único binario siempre que sea técnicamente razonable.

Ejemplo:

mobilelab

No quiero que Node.js sea obligatorio para utilizar el Core.

No quiero que Docker sea obligatorio para utilizar MobileLab.

Docker podrá añadirse posteriormente como integración opcional.

============================================================
3. ESTRUCTURA DEL REPOSITORIO
============================================================

Diseña un monorepo profesional.

Una estructura inicial posible sería:

mobilelab/
|
+-- cmd/
|   +-- mobilelab/
|
+-- internal/
|   +-- core/
|   +-- config/
|   +-- server/
|   +-- sandbox/
|   +-- scenario/
|   +-- device/
|   +-- platform/
|   +-- storage/
|   +-- reporting/
|   +-- logging/
|
+-- pkg/
|   +-- mobilelab/
|
+-- adapters/
|   +-- android/
|   +-- ios/
|   +-- flutter/
|   +-- react-native/
|   +-- capacitor/
|
+-- dashboard/
|
+-- sdks/
|   +-- flutter/
|   +-- react-native/
|   +-- android/
|   +-- ios/
|   +-- capacitor/
|
+-- examples/
|   +-- flutter/
|   +-- react-native/
|   +-- android/
|   +-- ios/
|   +-- ionic/
|
+-- docs/
|
+-- schemas/
|
+-- scripts/
|
+-- tests/
|
+-- .github/
|   +-- workflows/
|
+-- mobilelab.example.yaml
+-- LICENSE
+-- README.md
+-- CONTRIBUTING.md
+-- SECURITY.md
+-- Makefile

No sigas esta estructura ciegamente.

Si encuentras una arquitectura mejor, propónla y justifica brevemente la decisión.

============================================================
4. CLEAN ARCHITECTURE
============================================================

Quiero separación estricta de responsabilidades.

El dominio no debe depender de:

- CLI
- HTTP
- SQLite
- Android
- iOS
- Flutter
- React Native
- UI
- infraestructura

Utiliza conceptos como:

Domain
Application
Infrastructure
Adapters

Aplica:

SOLID
Dependency Inversion
Interfaces
Ports & Adapters
Clean Architecture

Evita overengineering.

============================================================
5. CLI
============================================================

Quiero una CLI moderna.

Comando principal:

mobilelab

Debe soportar progresivamente:

mobilelab init
mobilelab start
mobilelab stop
mobilelab status
mobilelab doctor
mobilelab detect

mobilelab api
mobilelab api mock
mobilelab api error
mobilelab api latency

mobilelab auth
mobilelab auth expire
mobilelab auth reset

mobilelab network
mobilelab network offline
mobilelab network online
mobilelab network slow

mobilelab location
mobilelab location set

mobilelab deeplink
mobilelab deeplink open

mobilelab scenario
mobilelab scenario list
mobilelab scenario run

mobilelab run

mobilelab dashboard

Ejemplo:

mobilelab start

Salida esperada:

MobileLab

Starting local environment...

✓ API Server        http://localhost:4566
✓ Auth Server       http://localhost:4567
✓ WebSocket         ws://localhost:4568
✓ Dashboard         http://localhost:4569

Environment ready.

No copies exactamente esta salida si encuentras una UX mejor.

============================================================
6. MOBILELAB INIT
============================================================

Quiero que:

mobilelab init

analice automáticamente el proyecto actual.

Debe detectar cuando sea posible:

Flutter
React Native
Android
iOS
Ionic
Capacitor

Por ejemplo:

Analyzing project...

✓ Framework: Flutter
✓ Android detected
✓ iOS detected
✓ OpenAPI specification found

Generating MobileLab environment...

✓ mobilelab.yaml
✓ mobilelab/scenarios/
✓ mobilelab/fixtures/
✓ mobilelab/mocks/

Ready.

La detección debe realizarse mediante archivos reales del proyecto.

Ejemplos:

Flutter:
pubspec.yaml

React Native:
package.json + react-native dependency

Android:
build.gradle
build.gradle.kts
settings.gradle
AndroidManifest.xml

iOS:
.xcodeproj
.xcworkspace
Package.swift cuando corresponda

Ionic:
ionic.config.json

Capacitor:
capacitor.config.ts/json

Implementa detectores desacoplados mediante interfaces.

============================================================
7. CONFIGURACIÓN
============================================================

MobileLab utilizará:

mobilelab.yaml

Ejemplo conceptual:

project:
  name: my-mobile-app

server:
  host: 127.0.0.1
  port: 4566

dashboard:
  enabled: true
  port: 4569

sandbox:
  latency: 0
  error_rate: 0

auth:
  enabled: true

device:
  auto_detect: true

No guardes secretos reales.

Implementa validación estricta.

Los errores de configuración deben ser claros.

============================================================
8. API SANDBOX
============================================================

Una de las funciones principales será simular APIs.

Debe permitir definir endpoints mediante YAML.

Ejemplo:

endpoints:

  - path: /api/users
    method: GET
    response:
      status: 200
      body:
        users:
          - id: 1
            name: Gian

  - path: /api/payments
    method: POST
    response:
      status: 200
      body:
        success: true
        transactionId: ABC123

Los endpoints deben poder configurar:

status
headers
body
delay
fixtures
errors

Ejemplo:

delay: 2000

También:

mobilelab api latency 5000

debe permitir aplicar latencia global.

Y:

mobilelab api error 500

debe permitir simular temporalmente errores.

Diseña el sistema para poder posteriormente soportar:

GraphQL
gRPC
SSE

pero NO es obligatorio implementarlos en la primera versión.

============================================================
9. OPENAPI
============================================================

Esta característica es prioritaria.

Quiero poder ejecutar:

mobilelab api import openapi.yaml

o:

mobilelab init --openapi openapi.yaml

MobileLab debe analizar OpenAPI y generar automáticamente mocks básicos.

Ejemplo:

34 endpoints detected
12 schemas detected
8 initial scenarios generated

Debe generar:

mobilelab/
    mocks/
    fixtures/
    scenarios/

Soporta inicialmente OpenAPI 3.x.

No intentes implementar el estándar completo desde cero si existe una librería madura y mantenida para Go.

============================================================
10. AUTH SANDBOX
============================================================

Implementa un sistema local básico de autenticación.

Debe permitir simular:

login exitoso
credenciales inválidas
JWT válido
JWT expirado
refresh token
401
403

Ejemplo:

mobilelab auth expire

A partir de ese momento las APIs protegidas deben comportarse como sesión expirada según la configuración.

No quiero implementar un Identity Provider empresarial completo en la v0.1.

Diseña interfaces que permitan posteriormente soportar:

OAuth2
OIDC
PKCE
SSO

============================================================
11. LATENCIA Y ERRORES
============================================================

Debe poder simular condiciones reales.

Ejemplos:

mobilelab api latency 3000

mobilelab api error 500

mobilelab api error 401

mobilelab api reset

También quiero configuración por endpoint.

Ejemplo:

endpoints:
  - path: /payments
    latency: 5000

  - path: /profile
    error:
      status: 500

Debe ser posible activar/desactivar condiciones durante la ejecución.

============================================================
12. REQUEST INSPECTOR
============================================================

Cada request recibido debe registrarse.

Ejemplo:

12:32:01 GET  /api/profile    200   32ms
12:32:04 POST /api/payment    500   2.1s

Guardar cuando corresponda:

method
path
query
headers sanitizados
body sanitizado
status
duration
timestamp

NUNCA imprimir automáticamente:

Authorization
cookies sensibles
passwords
tokens

Implementa redacción de secretos.

============================================================
13. SCENARIO ENGINE
============================================================

Esta es una característica central del producto.

Quiero definir escenarios mediante YAML.

Ejemplo:

name: Payment with expired session

backend:
  latency: 2000

auth:
  token: expired

device:
  network: online

steps:
  - launch_app
  - open_deeplink: /payments

expect:
  - request:
      method: POST
      path: /payments

  - response:
      status: 200

Debe existir:

mobilelab scenario list

mobilelab scenario run payment-expired

y:

mobilelab run payment-expired.yaml

Diseña un modelo de dominio sólido para escenarios.

No mezcles el parser YAML directamente con el motor.

Debe existir algo equivalente a:

ScenarioDefinition
ScenarioParser
ScenarioRunner
ScenarioStep
ScenarioAssertion
ScenarioResult

============================================================
14. PORTABILIDAD DE ESCENARIOS
============================================================

Principio:

Write once. Run everywhere.

Un escenario no debe contener lógica específica de Flutter o React Native salvo que sea estrictamente necesaria.

Debe poder ejecutarse:

mobilelab run checkout.yaml --platform flutter

mobilelab run checkout.yaml --platform react-native

mobilelab run checkout.yaml --platform android

mobilelab run checkout.yaml --platform ios

mobilelab run checkout.yaml --platform capacitor

Y posteriormente:

mobilelab run checkout.yaml --all

La sintaxis del escenario debe ser independiente de la plataforma.

============================================================
15. DEVICE ENGINE
============================================================

Crear una abstracción:

DeviceAdapter

Ejemplo conceptual:

type DeviceAdapter interface {
    Detect(...)
    LaunchApp(...)
    StopApp(...)
    SetLocation(...)
    OpenDeepLink(...)
    SetNetworkCondition(...)
    GetDeviceInfo(...)
}

No acoples ScenarioEngine directamente a ADB o simctl.

Implementaciones:

AndroidDeviceAdapter
IOSDeviceAdapter

============================================================
16. ANDROID
============================================================

Utiliza herramientas oficiales disponibles localmente.

Principalmente:

adb

Debe poder detectar:

dispositivos físicos
emuladores

Inicialmente soportar cuando sea técnicamente viable:

launch app
stop app
clear app
deep links
location en emulator
device information

Para network simulation, investiga cuidadosamente qué puede hacerse realmente mediante ADB/emulador.

NO inventes comandos inexistentes.

Si alguna capacidad requiere privilegios, emulator console o configuración especial, documenta la limitación.

============================================================
17. iOS
============================================================

Utiliza:

xcrun simctl

Inicialmente soportar:

detectar simuladores
boot
launch
terminate
deep links
location cuando esté disponible

Recuerda:

iOS tooling requiere macOS + Xcode.

La aplicación principal debe seguir pudiendo compilar/funcionar en otros sistemas aunque el adapter iOS no esté disponible.

Implementa detección de capacidades.

Ejemplo:

mobilelab doctor

Android SDK       ✓
ADB               ✓
Xcode             unavailable
iOS Simulator     unavailable

============================================================
18. NETWORK CONDITIONS
============================================================

Quiero comandos conceptuales:

mobilelab network offline
mobilelab network online
mobilelab network slow

Pero NO asumas que Android e iOS permiten exactamente las mismas capacidades.

Diseña:

NetworkCondition

con capabilities.

Ejemplo:

Offline
Online
Latency
Bandwidth
PacketLoss

Cada adapter debe informar cuáles soporta.

No simules falsamente algo que el sistema operativo no permite.

============================================================
19. LOCATION
============================================================

Ejemplo:

mobilelab location set -5.1945 -80.6328

Debe utilizar Android Emulator / iOS Simulator cuando sea posible.

No quiero que la app tenga que modificar su código para pruebas básicas de ubicación.

============================================================
20. DEEP LINKS
============================================================

Ejemplo:

mobilelab deeplink open "myapp://payments/123"

Debe funcionar mediante:

Android adapter
iOS adapter

cuando el dispositivo lo permita.

============================================================
21. PUSH NOTIFICATIONS
============================================================

Diseña la abstracción desde v0.1.

No es obligatorio implementar soporte completo para FCM/APNs inicialmente.

Quiero eventualmente:

mobilelab push send payment-success

y fixtures:

push:
  payment-success:
    title: Payment completed
    body: Your payment was processed
    data:
      transactionId: ABC123

Implementa únicamente lo que sea técnicamente correcto localmente.

Documenta las limitaciones de Android/iOS.

============================================================
22. FRAMEWORK ADAPTERS
============================================================

MobileLab debe ser framework-agnostic.

Pero crea interfaces para integraciones avanzadas.

Flutter:

mobilelab_flutter

React Native:

@mobilelab/react-native

Android:

mobilelab-android

iOS:

MobileLabKit

Capacitor:

@mobilelab/capacitor

IMPORTANTE:

No obligues a instalar estos SDK para utilizar API Sandbox.

La mayoría de funciones básicas deben funcionar sin modificar la aplicación.

SDKs = capacidades avanzadas.

============================================================
23. DASHBOARD
============================================================

Crear un dashboard local sencillo.

Inicialmente mostrar:

MobileLab status

API server status
connected devices
requests
responses
latency
active errors
active scenario
logs

Ejemplo conceptual:

MobileLab Dashboard

Environment
API          ONLINE
Auth         ONLINE
Device       Pixel Emulator
Network      ONLINE
Latency      300ms

Recent Requests

GET  /profile       200
POST /payments      500
GET  /insurance     200

No inviertas demasiado tiempo inicialmente en diseño visual.

Prioriza arquitectura y funcionalidad.

============================================================
24. WEBSOCKET
============================================================

Implementa infraestructura básica para comunicación realtime entre:

Core
Dashboard
CLI cuando corresponda

Esto permitirá actualizar logs y estados en tiempo real.

Diseña el protocolo/eventos de manera tipada.

============================================================
25. PERSISTENCIA
============================================================

Utiliza SQLite cuando sea necesario.

Por ejemplo:

request history
scenario runs
settings
environment state

No guardes secretos.

Crea repositories/interfaces para evitar acoplamiento directo.

============================================================
26. MOBILELAB DOCTOR
============================================================

Quiero:

mobilelab doctor

Ejemplo:

MobileLab Doctor

Core
✓ Configuration
✓ Local ports

Android
✓ Android SDK
✓ ADB
✓ Emulator

iOS
✗ Xcode unavailable
✗ simctl unavailable

Frameworks
✓ Flutter
✓ Node
✓ React Native tooling

Environment ready with 2 warnings.

Debe funcionar incluso si ciertas plataformas no están instaladas.

============================================================
27. MOBILELAB DETECT
============================================================

Ejemplo:

mobilelab detect

Detected:

Flutter
Android SDK
Android Emulator
React Native
Node.js

Available devices:

Pixel 9
emulator-5554

Debe utilizar detectores desacoplados.

============================================================
28. REPORTES
============================================================

Al ejecutar escenarios:

mobilelab run scenario.yaml

mostrar:

Scenario: Payment failure

✓ Environment configured
✓ Application launched
✓ POST /payments detected
✓ Expected 500 received
✓ Application remained responsive

PASSED

Duration: 4.82s

Implementa inicialmente:

terminal report
JSON report

Diseña la abstracción para posteriormente agregar:

HTML
JUnit XML

para CI/CD.

============================================================
29. CI/CD
============================================================

Debe existir modo headless.

Ejemplo:

mobilelab start --headless

mobilelab run ./scenarios --report junit

No necesitas implementar device farm.

Pero diseña MobileLab para poder ejecutarse posteriormente en:

GitHub Actions
GitLab CI
Azure DevOps
Jenkins

============================================================
30. OBSERVABILIDAD
============================================================

Implementa logging estructurado.

Niveles:

debug
info
warn
error

Evita logs innecesariamente ruidosos.

Agregar:

--verbose

cuando corresponda.

============================================================
31. SEGURIDAD
============================================================

MobileLab es una herramienta de desarrollo LOCAL.

Por defecto:

bind únicamente a 127.0.0.1.

NO:

0.0.0.0

salvo que el usuario lo solicite explícitamente.

Nunca guardar:

passwords reales
tokens reales
Authorization headers completos

Sanitizar logs.

Mostrar warning si el usuario expone MobileLab a la red.

============================================================
32. PERFORMANCE
============================================================

Quiero que el Core sea ligero.

Objetivos orientativos, no promesas:

- startup rápido
- bajo consumo de RAM
- binario razonablemente pequeño
- cientos de requests locales sin problemas
- shutdown limpio

Mide antes de optimizar.

Incluye benchmarks para componentes críticos cuando tenga sentido.

============================================================
33. UX DEL CLI
============================================================

La experiencia de terminal es importantísima.

Errores malos:

error code 34 config parser failed

Errores buenos:

Unable to start MobileLab.

Port 4566 is already in use.

Try:
mobilelab start --port 4567

Los mensajes deben indicar cómo resolver problemas.

============================================================
34. MOBILELAB START
============================================================

El comando:

mobilelab start

debe:

1. localizar configuración
2. validar configuración
3. cargar fixtures
4. cargar mocks
5. iniciar API server
6. iniciar auth sandbox
7. iniciar event bus
8. iniciar dashboard si está habilitado
9. detectar devices opcionalmente
10. mostrar estado

Debe responder correctamente a:

Ctrl+C

y cerrar recursos limpiamente.

============================================================
35. MOBILELAB STOP
============================================================

Diseña una estrategia robusta para detener una instancia existente.

No dependas exclusivamente de matar procesos arbitrariamente.

Utiliza PID/state/control endpoint o una estrategia apropiada.

============================================================
36. FIXTURES
============================================================

Ejemplo:

mobilelab/fixtures/user.json

{
  "id": "123",
  "name": "Test User",
  "plan": "premium"
}

Desde YAML:

response:
  fixture: user.json

Debe validar que no exista path traversal.

============================================================
37. VARIABLES
============================================================

Permite variables locales seguras.

Ejemplo:

variables:
  userId: "123"

Response:

{
  "id": "{{userId}}"
}

NO construyas un template engine excesivamente complejo.

============================================================
38. SCENARIO RECORDER
============================================================

Diseña la arquitectura para una futura funcionalidad:

mobilelab record login

El objetivo futuro será registrar:

requests
responses
deeplinks
environment changes

y generar un escenario reproducible.

NO es obligatorio implementar completamente el recorder en v0.1.

Sí debes dejar interfaces/extensiones preparadas.

============================================================
39. PLUGINS
============================================================

Diseña una futura arquitectura de plugins.

Ejemplos futuros:

mobilelab-plugin-firebase
mobilelab-plugin-supabase
mobilelab-plugin-graphql
mobilelab-plugin-grpc

NO implementes todos ahora.

Evita una arquitectura que impida agregarlos posteriormente.

============================================================
40. VERSIONADO
============================================================

Utiliza Semantic Versioning.

Inicial:

0.1.0

Comandos:

mobilelab version

y:

mobilelab --version

============================================================
41. TESTING
============================================================

Testing es obligatorio.

Incluye:

unit tests
integration tests

Especialmente para:

config parser
scenario parser
API sandbox
routing
latency
errors
auth
secret redaction
fixtures
OpenAPI importer
device adapter interfaces

Los tests no deben depender siempre de tener Android/iOS disponibles.

Utiliza FakeDeviceAdapter.

Ejemplo:

FakeDeviceAdapter

para ScenarioEngine.

============================================================
42. DOCUMENTACIÓN
============================================================

README profesional.

Debe explicar:

What is MobileLab?
Why MobileLab?
Installation
Quick Start
Architecture
API Sandbox
Scenarios
Android
iOS
Flutter
React Native
Ionic
CI
Roadmap
Contributing

Quick Start ideal:

mobilelab init
mobilelab start

y listo.

============================================================
43. EJEMPLO FLUTTER
============================================================

Crear una app mínima de ejemplo.

Debe consumir:

http://localhost:<port>

IMPORTANTE Android Emulator:

localhost desde Android Emulator no siempre apunta al host.

Gestiona/documenta correctamente:

10.0.2.2

o una estrategia equivalente.

No ocultes este problema.

MobileLab debería eventualmente proporcionar:

mobilelab endpoint

que devuelva el endpoint correcto según device.

============================================================
44. EJEMPLO REACT NATIVE
============================================================

Crear ejemplo mínimo equivalente.

Debe demostrar que el Sandbox no depende de Flutter.

============================================================
45. ANDROID/KOTLIN
============================================================

Crear ejemplo mínimo nativo.

Demostrar consumo del mismo endpoint y escenario.

============================================================
46. IOS/SWIFT
============================================================

Crear ejemplo mínimo si el entorno de desarrollo actual permite generar/verificarlo.

No bloquees todo el proyecto si estamos trabajando fuera de macOS.

============================================================
47. IONIC/CAPACITOR
============================================================

Crear ejemplo mínimo o scaffolding documentado.

Debe consumir el mismo Sandbox.

============================================================
48. PRINCIPIO FUNDAMENTAL DE COMPATIBILIDAD
============================================================

ESTO ES MUY IMPORTANTE:

NO quiero código como:

if flutter ...
else if reactNative ...
else if ionic ...

disperso por todo el Core.

Quiero:

interfaces
adapters
capability detection
dependency inversion

El Core no debe saber detalles innecesarios del framework.

============================================================
49. CAPABILITY SYSTEM
============================================================

Implementa un sistema para consultar capacidades.

Ejemplo:

mobilelab capabilities

Android Emulator:

launch             ✓
deepLink           ✓
location           ✓
networkOffline     ✓
networkLatency     partial
push               partial

iOS Simulator:

launch             ✓
deepLink           ✓
location           ✓
networkOffline     unavailable
...

Los valores reales deben basarse en capacidades implementadas.

Nunca anuncies una funcionalidad que realmente no existe.

============================================================
50. MVP 0.1
============================================================

No intentes terminar todas las ideas anteriores de golpe.

El MVP funcional obligatorio debe contener:

1. CLI
2. init
3. start
4. stop
5. status
6. doctor
7. detect
8. YAML config
9. API mock server
10. fixtures
11. configurable responses
12. latency simulation
13. HTTP error simulation
14. basic auth/JWT sandbox
15. request logging
16. secret redaction
17. scenario parser
18. scenario runner
19. assertions básicas
20. FakeDeviceAdapter
21. Android adapter básico
22. iOS adapter básico cuando el SO lo permita
23. deep links
24. location cuando esté soportado
25. JSON reporting
26. basic dashboard
27. OpenAPI import básico
28. unit tests
29. integration tests
30. README
31. GitHub Actions CI

Las características más complejas pueden quedar correctamente documentadas en ROADMAP.md.

============================================================
51. ROADMAP
============================================================

Genera:

ROADMAP.md

Propuesta:

0.1
Core + Sandbox + scenarios

0.2
Advanced Android/iOS Device Engine

0.3
Flutter + React Native advanced adapters

0.4
Capacitor + native SDKs

0.5
Recorder + replay

0.6
CI integrations

0.7
Plugin ecosystem

1.0
Stable MobileLab platform

Puedes modificar este roadmap si técnicamente tiene más sentido.

============================================================
52. LICENCIA
============================================================

Preparar el proyecto para licencia MIT inicialmente.

Crear LICENSE.

============================================================
53. GITHUB
============================================================

Preparar:

.github/workflows/

CI para:

build
test
lint

Considerar matrices para:

Linux
macOS
Windows

No hacer que los tests normales fallen porque Xcode no exista en Linux/Windows.

============================================================
54. MAKEFILE
============================================================

Agregar comandos útiles:

make build
make test
make lint
make run
make clean

Si propones Taskfile u otra alternativa, Makefile debe seguir existiendo inicialmente.

============================================================
55. CALIDAD
============================================================

No quiero:

TODOs masivos
métodos vacíos fingiendo funcionalidad
hardcodes innecesarios
arquitectura monolítica
dependencias globales
errores ignorados
panic para errores recuperables
secretos en código
funcionalidades falsas

Prefiero una funcionalidad NO implementada que diga:

Capability not available on this platform.

antes que simular que funciona.

============================================================
56. FORMA DE TRABAJAR
============================================================

IMPORTANTE:

No quiero que simplemente me expliques cómo construirlo.

QUIERO QUE LO CONSTRUYAS.

Trabaja directamente sobre el repositorio.

Primero:

1. inspecciona el repositorio actual
2. identifica el estado
3. crea ARCHITECTURE.md
4. define el plan técnico
5. crea estructura
6. implementa el Core
7. implementa CLI
8. implementa Sandbox
9. implementa Scenario Engine
10. implementa Device abstractions
11. implementa adapters
12. implementa dashboard mínimo
13. crea ejemplos
14. crea tests
15. ejecuta tests
16. corrige errores
17. ejecuta build
18. corrige errores
19. actualiza documentación

NO te detengas después de generar solamente la arquitectura.

Continúa implementando.

============================================================
57. REGLA DE ITERACIÓN
============================================================

Trabaja incrementalmente.

Después de cada módulo importante:

- compila
- ejecuta tests
- corrige errores
- continúa

No escribas 100 archivos antes de comprobar que compilan.

============================================================
58. DEPENDENCIAS
============================================================

Antes de añadir una dependencia:

- verifica que realmente sea necesaria
- prioriza librerías maduras
- evita dependencias abandonadas
- minimiza dependency footprint

Documenta las dependencias principales en ARCHITECTURE.md.

============================================================
59. CROSS PLATFORM
============================================================

El Core debe poder compilar idealmente en:

macOS
Linux
Windows

Android tooling dependerá del Android SDK.

iOS tooling únicamente estará disponible en macOS.

Utiliza build constraints/build tags cuando sean apropiados.

============================================================
60. NOMBRE Y BRANDING
============================================================

Usa "MobileLab" provisionalmente en código y documentación.

No construyas lógica dependiente del nombre.

Debe ser relativamente sencillo renombrarlo posteriormente.

============================================================
61. DEFINITION OF DONE DEL MVP
============================================================

Consideraré el primer MVP exitoso cuando pueda hacer:

git clone <repository>

cd mobilelab

make build

./bin/mobilelab doctor

./bin/mobilelab init

./bin/mobilelab start

y desde una aplicación mobile pueda realizar:

GET /api/profile

contra MobileLab y recibir un fixture.

Después pueda ejecutar:

mobilelab api latency 3000

y observar que el endpoint tarda aproximadamente 3 segundos.

Después:

mobilelab api error 500

y recibir HTTP 500.

Después:

mobilelab api reset

y volver al comportamiento normal.

Después pueda ejecutar:

mobilelab run scenarios/payment-error.yaml

y obtener:

PASS / FAIL

con un reporte reproducible.

Y si existe Android Emulator:

mobilelab detect

debe detectarlo.

Además:

mobilelab deeplink open "myapp://test"

debe utilizar el adapter Android cuando corresponda.

============================================================
62. PRIMERA EJECUCIÓN
============================================================

Empieza ahora.

Antes de modificar código:

A. inspecciona completamente el repositorio;
B. determina si está vacío o ya existe código;
C. crea un plan corto;
D. define decisiones arquitectónicas;
E. comienza inmediatamente la implementación.

No me pidas confirmación para decisiones técnicas normales.

Toma decisiones razonables como Principal Engineer.

Solo detente para preguntarme cuando exista una decisión que:

- cambie radicalmente el producto,
- implique credenciales,
- pueda destruir datos,
- requiera un servicio de pago,
- o sea imposible determinar técnicamente.

============================================================
63. RESTRICCIÓN IMPORTANTE
============================================================

No construyas una copia de LocalStack, Floci, Firebase Emulator Suite u otra herramienta.

MobileLab tiene otro objetivo:

desarrollo y testing LOCAL específicamente orientado al ecosistema MOBILE.

Su diferenciación debe ser:

API Sandbox
+
Device Simulation
+
Scenario Engine
+
Cross-framework Mobile Testing

Todo integrado bajo una sola experiencia.

============================================================
64. NORTH STAR
============================================================

Cada decisión arquitectónica debe responder a esta pregunta:

"¿Esto ayuda a que un desarrollador mobile pueda reproducir localmente un escenario complejo de su aplicación de manera sencilla?"

Si la respuesta es no, probablemente no pertenece al MVP.

============================================================
65. RESULTADO ESPERADO DE TU TRABAJO
============================================================

No quiero únicamente pseudocódigo.

Al finalizar cada iteración quiero:

- código real
- proyecto compilable
- tests pasando
- documentación actualizada
- ejemplos ejecutables
- lista corta de lo implementado
- lista honesta de limitaciones
- siguiente milestone recomendado

Empieza implementando MobileLab v0.1.