# CONTEXT.md — Contexto Técnico y Estado Actual del Proyecto

> **Propósito de este archivo**: Mantener una radiografía técnica actualizada de **AutoMarket (`Car-Shop`)**. Debe consultarse antes de realizar cambios y actualizarse cada vez que se modifique la arquitectura, modelos, endpoints o vistas del sistema.

---

## 1. Descripción General

**AutoMarket** es una aplicación web y API REST desarrollada en **Go (Gin)** para la compra y venta de vehículos usados en Colombia. Permite:
- A los **visitantes** consultar el catálogo público de vehículos aprobados y ver los datos de contacto del vendedor.
- A los **vendedores** registrarse, iniciar sesión, publicar vehículos (que quedan pendientes de moderación) y reportar la venta de sus vehículos publicados.
- A los **administradores** revisar vehículos pendientes, aprobarlos para su publicación en el catálogo, y eliminar vehículos o usuarios.

---

## 2. Estructura del Proyecto

```text
Car-Shop/
├── cmd/
│   └── web/
│       └── main.go                 # Punto de entrada, configuración por variables de entorno y arranque HTTP
├── internal/
│   ├── audit/
│   │   └── audit.go                # Registro de auditoría JSONL encadenado con hashes SHA-256
│   ├── domain/
│   │   └── models.go               # Entidades de dominio (User, Vehicle, PublicVehicle, SellerContact) y estados
│   ├── httpapi/
│   │   ├── router.go               # Definición de rutas Gin, grupos por rol y middlewares globales
│   │   └── handlers.go             # Controladores HTTP, validación de tokens Bearer y mapeo de errores
│   ├── service/
│   │   └── service.go              # Lógica de negocio, validaciones, sesiones en memoria y PBKDF2-SHA256
│   └── store/
│       └── store.go                # Persistencia en memoria sincronizada en disco (data/automarket.json)
├── web/
│   ├── static/
│   │   ├── catalogo.js             # Lógica cliente del catálogo, búsqueda, modales y sesión
│   │   ├── estilos.css             # Estilos globales, barra superior, catálogo, modales, slider y ficha técnica
│   │   ├── login.js                # Lógica cliente para login.html
│   │   ├── placeholder-car.svg     # Gráfico vectorial de respaldo para vehículos sin foto
│   │   └── vehiculo.js             # Lógica cliente de la vista de detalle y slider deslizable de fotos
│   └── templates/
│       ├── inicio.html             # Catálogo público, modales de login/registro, panel admin y solicitudes
│       ├── login.html              # Vista dedicada de inicio de sesión
│       └── vehiculo.html           # Vista individual de detalle con slider deslizable y ficha técnica
├── docs/
│   ├── DEMO.md                     # Guion paso a paso en PowerShell para demostrar el flujo completo
│   ├── DIAGRAMAS.md                # Diagramas C4, Clases, Casos de Uso y Secuencia en Mermaid
│   ├── KANBAN.md                   # Estado de tareas del proyecto para GitHub Projects
│   └── plantuml/                   # Fuentes PlantUML de los diagramas arquitectónicos
├── openspec/                       # Configuración para especificaciones de cambios (OpenSpec)
├── GEMINI.md                       # Reglas y directrices para el asistente de IA
├── CONTEXT.md                      # Este archivo: estado y arquitectura actual del proyecto
└── MEMORY.md                       # Bitácora histórica de cambios y modificaciones
```

---

## 3. Configuración y Variables de Entorno

El binario principal (`cmd/web/main.go`) lee las siguientes variables de entorno al iniciar:

| Variable | Valor por defecto | Descripción |
| :--- | :--- | :--- |
| `AUTOMARKET_DATA_FILE` | `data/automarket.json` | Ruta del archivo JSON de persistencia de usuarios y vehículos. |
| `AUTOMARKET_AUDIT_FILE` | `data/audit.jsonl` | Ruta del archivo JSONL para la cadena de auditoría HTTP. |
| `AUTOMARKET_ADDR` | `:8080` | Dirección y puerto donde escucha el servidor HTTP. |
| `ADMIN_EMAIL` | `admin@automarket.local` | Correo del administrador inicial (definido por defecto o variable de entorno). |
| `ADMIN_PASSWORD` | `AdminAutoMarket2026*` | Contraseña del administrador inicial (ver `ADMIN_CREDENTIALS.md`). |

---

## 4. Modelos de Dominio (`internal/domain/models.go`)

### Roles (`domain.Role`)
- `seller`: Vendedor registrado.
- `admin`: Administrador del sistema.

### Estados del Vehículo (`domain.VehicleStatus`)
- `pending`: Creado por el vendedor; oculto al público hasta que un administrador lo apruebe.
- `published`: Aprobado por un administrador; visible en el catálogo público.
- `sold`: Marcado como vendido por el vendedor propietario; deja de aparecer en el catálogo público.

### Entidades Principales
- **`User`**: `ID` (UUID v4), `Name`, `Email`, `Phone`, `Role`, `PasswordSalt` (hex), `PasswordHash` (hex PBKDF2-SHA256), `IsBlocked` (`bool`), `Warnings` (`[]string`), `CreatedAt` (UTC).
- **`TechnicalSpecs`**: `EngineDisplacement` (cilindrada), `FuelType` (combustible), `Horsepower` (HP), `TorqueNM` (Nm), `Transmission` (transmisión), `BodyType` (carrocería), `Doors` (puertas), `MileageKM` (kilometraje), `Color` (color exterior), `Traction` (tracción).
- **`Vehicle`**: `ID` (UUID v4), `Make`, `Model`, `Year`, `PriceCOP` (`int64`), `Description`, `Images` (`[]string`), `Specs` (`TechnicalSpecs`), `SellerID`, `Status`, `CreatedAt` (UTC).
- **`PublicVehicle`**: Proyección pública de `Vehicle` con `Images`, `Specs` y datos de contacto del vendedor `Seller` (`SellerContact`: `Name`, `Email`, `Phone`).

---

## 5. Rutas HTTP y API REST (`internal/httpapi/`)

### Vistas Web y Recursos Estáticos
| Método | Ruta | Descripción |
| :--- | :--- | :--- |
| `GET` | `/` | Renderiza `web/templates/inicio.html` (catálogo, login/registro, solicitudes de vendedor y panel admin). |
| `GET` | `/login` | Renderiza `web/templates/login.html` (página dedicada de login). |
| `GET` | `/vehiculo/:id` | Renderiza `web/templates/vehiculo.html` (detalle individual con slider de fotos y ficha técnica). |
| `GET` | `/static/*filepath` | Sirve archivos estáticos desde `web/static/`. |
| `GET` | `/uploads/*filepath` | Sirve fotografías subidas desde `data/uploads/`. |

### API Pública (`/api`)
| Método | Ruta | Auth | Descripción | Respuesta Exitosa |
| :--- | :--- | :--- | :--- | :--- |
| `POST` | `/api/sellers` | Pública | Registra un nuevo vendedor (`name`, `email`, `phone`, `password`). | `201 Created` |
| `POST` | `/api/sessions` | Pública | Inicia sesión (`email`, `password`) y devuelve token Bearer. | `200 OK` |
| `GET` | `/api/vehicles` | Pública | Lista todos los vehículos en estado `published` con fotos, specs y contacto. | `200 OK` |
| `GET` | `/api/vehicles/:id` | Pública | Obtiene el detalle de un vehículo en estado `published` con fotos y ficha técnica. | `200 OK` |
| `GET` | `/api/me` | Bearer | Devuelve los datos del usuario autenticado (incluye advertencias y estado de bloqueo). | `200 OK` |
| `POST` | `/api/uploads` | Bearer | Sube 1 o más fotos multipart (`JPG`, `PNG`, `WEBP`) y retorna sus URLs relativas. | `200 OK` |

### API de Vendedor (`/api`)
Requiere encabezado `Authorization: Bearer <token>` con rol `seller`.
| Método | Ruta | Auth | Descripción | Respuesta Exitosa |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/sellers/vehicles` | `seller` | Lista todas las solicitudes y vehículos del vendedor autenticado (`pending`, `published`, `sold`). | `200 OK` |
| `POST` | `/api/vehicles` | `seller` | Crea una publicación de vehículo en estado `pending`. | `201 Created` |
| `PUT` | `/api/vehicles/:id` | `seller` | Modifica el precio (`price_cop`) y descripción de un vehículo propio. | `200 OK` |
| `DELETE` | `/api/vehicles/:id` | `seller` | Elimina permanentemente una solicitud/vehículo propio. | `204 No Content` |
| `POST` | `/api/vehicles/:id/sale` | `seller` | Marca como `sold` un vehículo `published` propio del vendedor. | `204 No Content` |

### API de Administración (`/api/admin`)
Requiere encabezado `Authorization: Bearer <token>` con rol `admin`.
| Método | Ruta | Auth | Descripción | Respuesta Exitosa |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/admin/users` | `admin` | Lista todos los usuarios con estado de bloqueo y advertencias. | `200 OK` |
| `POST` | `/api/admin/users/:id/warn` | `admin` | Notifica / advierte a un vendedor por faltas éticas o incumplimiento de normas. | `200 OK` |
| `POST` | `/api/admin/users/:id/block` | `admin` | Bloquea la cuenta de un vendedor e invalida sus sesiones inmediatamente. | `200 OK` |
| `POST` | `/api/admin/users/:id/unblock` | `admin` | Desbloquea la cuenta de un vendedor. | `200 OK` |
| `DELETE` | `/api/admin/users/:id` | `admin` | Elimina un usuario, sus vehículos asociados y sus sesiones activas. | `204 No Content` |
| `GET` | `/api/admin/vehicles` | `admin` | Lista todos los vehículos de la plataforma (`pending`, `published`, `sold`) con vendedor. | `200 OK` |
| `GET` | `/api/admin/vehicles/pending` | `admin` | Lista todos los vehículos en estado `pending` con datos del vendedor. | `200 OK` |
| `POST` | `/api/admin/vehicles/:id/approve` | `admin` | Cambia el estado de un vehículo `pending` a `published`. | `204 No Content` |
| `DELETE` | `/api/admin/vehicles/:id` | `admin` | Elimina un vehículo del almacén sin importar su estado. | `204 No Content` |

---

## 6. Detalles de Implementación y Estado Actual

- **Persistencia (`internal/store`)**: Mantiene mapas en memoria (`users`, `vehicles`) sincronizados en `data/automarket.json`. El guardado utiliza escritura atómica con archivo temporal.
- **Sesiones y Bloqueo (`internal/service`)**: Sesiones en memoria. Si un usuario es bloqueado (`is_blocked: true`), sus sesiones activas se purgan al instante y el login rechaza sus credenciales con código `403 Forbidden`.
- **Auditoría (`internal/audit`)**: Registro criptográfico SHA-256 encadenado en `data/audit.jsonl`.
- **Interfaz Web (`web/templates/`)**:
  - `inicio.html`: Catálogo público con imágenes de portada, buscador en vivo, selector de ordenamiento y enlace directo a `/vehiculo/:id`. Modales para autenticación ("Iniciar sesión" / "Registrarse"), panel de solicitudes de vendedor y panel de control del administrador.
  - `vehiculo.html`: Vista individual de detalle (`/vehiculo/:id`) con slider de fotos interactivo (soporte para flechas, miniaturas, teclado y gestos táctiles), información principal, sección de ficha técnica exhaustiva y botones directos de contacto con el vendedor (llamada, correo y WhatsApp).
  - `login.html`: Pantalla secundaria de acceso con redirección hacia el catálogo.

