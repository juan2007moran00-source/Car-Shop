# GEMINI.md — Reglas y Directrices del Proyecto AutoMarket (Car-Shop)

Este archivo define las reglas obligatorias, convenciones arquitectónicas y el flujo de trabajo que el asistente de IA debe seguir al analizar, modificar o extender este repositorio.

---

## 1. Flujo de Trabajo de Memoria y Contexto (Obligatorio)

1. **Lectura previa**: Antes de proponer o aplicar cualquier cambio en el código, lee siempre [`CONTEXT.md`](./CONTEXT.md) para entender la arquitectura actual y [`MEMORY.md`](./MEMORY.md) para conocer el historial de modificaciones recientes y decisiones previas.
2. **Registro posterior**: Al finalizar cualquier modificación funcional, corrección de bugs, refactorización o cambio arquitectónico:
   - Registra una nueva entrada cronológica en [`MEMORY.md`](./MEMORY.md) detallando qué cambió, qué archivos se tocaron y por qué.
   - Actualiza [`CONTEXT.md`](./CONTEXT.md) si el cambio altera rutas HTTP, modelos de datos, reglas de negocio, variables de entorno, vistas web o estructura de directorios.

---

## 2. Idioma y Estilo de Comunicación

- **Idioma del proyecto**: Español.
- Todos los mensajes de error de la API (`{"error": "..."}`), logs del servidor, comentarios relevantes, documentación (`.md`, diagramas Mermaid/PlantUML) y textos de la interfaz de usuario deben escribirse en **español claro y consistente**.
- Los identificadores de código en Go (tipos, funciones, paquetes y etiquetas JSON) se mantienen en **inglés** siguiendo las convenciones actuales del código (`User`, `Vehicle`, `PublishVehicle`, `price_cop`, etc.).

---

## 3. Restricciones Tecnológicas y Dependencias

- **Lenguaje**: Go `1.26`.
- **Única dependencia directa**: `github.com/gin-gonic/gin`.
  - **Regla estricta**: No agregues nuevas dependencias externas a `go.mod` a menos que el usuario lo solicite explícitamente. Utiliza la biblioteca estándar de Go (`crypto/*`, `encoding/json`, `net/http`, `sync`, `os`, etc.) para cualquier funcionalidad nueva.
- **Frontend sin frameworks**: Las vistas en `web/templates/` y recursos en `web/static/` usan únicamente HTML5, CSS3 y JavaScript Vanilla (`fetch`, `localStorage`/`sessionStorage`). No introduzcas bundlers ni librerías JS externas salvo petición expresa.

---

## 4. Arquitectura y Separación de Capas

Respeta estrictamente las responsabilidades de cada paquete; no mezcles lógica HTTP con persistencia ni reglas de negocio con acceso a archivos:

- **`cmd/web/main.go`**:
  - Punto de entrada exclusivo. Lee variables de entorno (`AUTOMARKET_DATA_FILE`, `AUTOMARKET_AUDIT_FILE`, `AUTOMARKET_ADDR`, `ADMIN_EMAIL`, `ADMIN_PASSWORD`), instancia `store`, `audit`, `service` y `httpapi`, y arranca el servidor.
- **`internal/domain/models.go`**:
  - Modelos puros del dominio (`User`, `Vehicle`, `PublicVehicle`, `SellerContact`) y tipos enumerados (`Role`: `seller` | `admin`; `VehicleStatus`: `pending` | `published` | `sold`).
  - No debe importar ningún otro paquete interno del proyecto.
- **`internal/service/service.go`**:
  - Contiene todas las reglas de negocio, validaciones de entrada, gestión de sesiones en memoria y criptografía de contraseñas.
  - No debe depender de `gin` ni de detalles HTTP.
- **`internal/store/store.go`**:
  - Persistencia local en archivo JSON (`data/automarket.json`).
  - Toda lectura/escritura debe ser concurrentemente segura mediante `sync.RWMutex`.
  - Toda escritura en disco (`saveLocked`) debe ser **atómica** (escribir en archivo temporal `.automarket-*.tmp` con permisos `0600`, hacer `Sync()`, cerrar y renombrar) y revertir el estado en memoria si falla el guardado en disco.
- **`internal/audit/audit.go`**:
  - Bitácora de auditoría en formato JSON Lines (`data/audit.jsonl`).
  - Mantiene una cadena de integridad criptográfica (`PrevHash` -> `Hash` calculado con `SHA-256`) protegida por `sync.Mutex`.
  - Al iniciar el servidor, verifica la integridad completa del archivo de auditoría existente.
- **`internal/httpapi/` (`router.go`, `handlers.go`)**:
  - Configuración de Gin, middlewares (`auditMiddleware`, `gin.Recovery()`, `requestSizeLimit` de 1 MB, `authentication`, `requireRole`) y traducción de errores de negocio/persistencia a códigos de estado HTTP (`writeServiceError`).

---

## 5. Reglas de Negocio y Seguridad Invariantes

1. **Seguridad de contraseñas**:
   - Longitud mínima de contraseña: **8 caracteres**.
   - Hashing obligatorio con **PBKDF2-SHA256** (`600,000` iteraciones, salt aleatorio de 16 bytes, clave de 32 bytes) y comparación en tiempo constante (`subtle.ConstantTimeCompare`).
   - Jamás incluir `PasswordSalt` ni `PasswordHash` en respuestas de la API o logs.
2. **Autenticación y Sesiones**:
   - Tokens de sesión de 32 bytes aleatorios codificados en Base64 URL-safe (`Bearer <token>`).
   - En memoria solo se almacena el hash `SHA-256` del token asociado al `UserID`, protegido con `sync.RWMutex`.
   - Al eliminar un usuario, todas sus sesiones activas deben invalidarse inmediatamente.
3. **Flujo de estados de vehículos**:
   - Todo vehículo publicado por un vendedor (`RoleSeller`) nace obligatoriamente en estado `pending`.
   - Solo un administrador (`RoleAdmin`) puede aprobar un vehículo (`pending` -> `published`). Intentar aprobar un vehículo en otro estado debe devolver `ErrInvalidState` (`409 Conflict`).
   - Solo el vendedor dueño del vehículo puede reportar su venta (`published` -> `sold`). Si el vehículo pertenece a otro vendedor se devuelve `ErrNotFound` (`404 Not Found`), y si no está en estado `published` se devuelve `ErrInvalidState` (`409 Conflict`).
   - El catálogo público (`GET /api/vehicles` y `GET /api/vehicles/:id`) **solo** expone vehículos en estado `published` junto con los datos de contacto del vendedor (`SellerContact`).
4. **Validaciones de datos**:
   - `User`: `Name` (1–120 caracteres), `Email` (formato válido según `net/mail`, normalizado a minúsculas y único sin importar mayúsculas/minúsculas), `Phone` (1–40 caracteres para vendedores).
   - `Vehicle`: `Make` (1–80 caracteres), `Model` (1–120 caracteres), `Year` (entre `1886` y `año actual + 1`), `PriceCOP` (`> 0`), `Description` (máximo 1000 caracteres).
5. **Integridad al eliminar usuarios**:
   - Eliminar un usuario (`DELETE /api/admin/users/:id`) elimina en cascada todos los vehículos asociados a su `SellerID` dentro de la misma operación atómica de persistencia.
6. **Datos locales y Git**:
   - La carpeta `data/` (`automarket.json` y `audit.jsonl`) contiene datos de ejecución local y está excluida en `.gitignore`. Nunca comitear archivos dentro de `data/` ni credenciales reales.

---

## 6. Documentación y OpenSpec

- Si un cambio modifica la estructura de paquetes, modelos o flujos principales, verifica si es necesario actualizar [`docs/DIAGRAMAS.md`](./docs/DIAGRAMAS.md), los archivos PlantUML en [`docs/plantuml/`](./docs/plantuml/) o el guion en [`docs/DEMO.md`](./docs/DEMO.md).
- Cuando el usuario utilice flujos de trabajo **OpenSpec** (`/opsx-propose`, `/opsx-apply`, `/opsx-explore`, etc.), sigue las habilidades definidas en `.agents/skills/openspec-*` y la configuración de `openspec/config.yaml`.

