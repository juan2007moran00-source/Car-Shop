# AutoMarket

Aplicación web en Go para consultar un catálogo público de vehículos usados, registrar vendedores, publicar vehículos con fotografías y moderar las publicaciones desde un panel de administración.

## Funcionalidad

- **Catálogo público sin registro (HU-VIS-01):** cualquier visitante puede ver los vehículos publicados y los datos de contacto del vendedor.
- **Vendedor (HU-VEN-01):** se registra, publica vehículos con fotos (quedan pendientes de aprobación), ve el estado de sus publicaciones y reporta una venta para retirarla del catálogo.
- **Administrador (HU-ADM-01):** autoriza las publicaciones pendientes y elimina publicaciones o usuarios.
- **Seguridad AAA (RF-01):** autenticación con token Bearer, autorización por rol y bitácora de auditoría con cadena de hashes que registra cada operación.
- **HTTPS:** el servidor usa TLS. En desarrollo genera un certificado autofirmado.
- **Persistencia:** todos los usuarios y vehículos se guardan en PostgreSQL.

## Imágenes

- Se aceptan archivos **JPG o PNG**; se valida el contenido real del archivo, no solo la extensión.
- Una de las imágenes cargadas se marca como portada.
- Las imágenes se guardan en la base de datos PostgreSQL (columna `data BYTEA` en `vehicle_images`) y se sirven mediante `/media/{vehicleID}/{archivo}`. La URL pública de las imágenes no cambia.
- El almacenamiento en disco `data/uploads/` ya no se usa.

## Iniciar el servidor

Necesitas Go y una base de datos PostgreSQL. Crea la base y define las variables antes de arrancar (PowerShell, desde la raíz del repositorio):

```powershell
$env:AUTOMARKET_DATABASE_URL = "postgres://usuario:clave@host:puerto/nombre_basedatos?sslmode=require"
$env:ADMIN_EMAIL = "admin@automarket.local"
$env:ADMIN_PASSWORD = "una-clave-de-al-menos-8-caracteres"
go run ./cmd/web
```

La aplicación se conectará a PostgreSQL y creará las tablas automáticamente si no existen.

Si se define la variable `PORT`, la aplicación escuchará HTTP plano en `0.0.0.0:PORT` (modo Render, sin TLS). Si no, usará HTTPS con certificado autofirmado en `:8443`.

### Variables de entorno

| Variable | Valor por defecto | Uso |
| --- | --- | --- |
| `AUTOMARKET_DATABASE_URL` | — (obligatoria) | Cadena de conexión a PostgreSQL. **Debe incluir `sslmode=require`** y usar la URL directa de Neon (no pooler), o añadir `default_query_exec_mode=simple_protocol`. |
| `ADMIN_EMAIL` / `ADMIN_PASSWORD` | — | Credenciales de la cuenta de administrador (se definen juntas) |
| `AUTOMARKET_ADDR` | `:8443` | Dirección del servidor HTTPS (local). Ignorado en modo Render |
| `AUTOMARKET_CERT_DIR` | `data/certs` | Carpeta del certificado de desarrollo (modo local solo) |

## Administrador

La cuenta se crea la primera vez que arrancas con `ADMIN_EMAIL` y `ADMIN_PASSWORD`. Si el correo ya existe, su contraseña no se modifica; para cambiarla hay que actualizarla directamente en la tabla `users`.

## Rutas principales

| Método | Ruta | Acceso | Descripción |
| --- | --- | --- | --- |
| `GET` | `/api/vehicles` | Público | Catálogo de vehículos publicados con contacto |
| `GET` | `/api/vehicles/:id` | Público | Detalle de un vehículo publicado |
| `GET` | `/media/:vehicleID/:archivo` | Público | Imagen de una publicación |
| `POST` | `/api/sellers` | Público | Registro de vendedor |
| `POST` | `/api/sessions` | Público | Inicio de sesión |
| `GET` | `/api/session` | Autenticado | Usuario de la sesión actual |
| `DELETE` | `/api/session` | Autenticado | Cierre de sesión |
| `GET` | `/api/seller/vehicles` | Vendedor | Publicaciones propias y su estado |
| `POST` | `/api/vehicles` | Vendedor | Publicar vehículo (`multipart/form-data` con `images` y `cover_index`) |
| `POST` | `/api/vehicles/:id/sale` | Vendedor | Reportar venta |
| `GET` | `/api/admin/vehicles/pending` | Admin | Publicaciones pendientes |
| `GET` | `/api/admin/vehicles` | Admin | Todas las publicaciones |
| `POST` | `/api/admin/vehicles/:id/approve` | Admin | Aprobar publicación |
| `DELETE` | `/api/admin/vehicles/:id` | Admin | Eliminar publicación |
| `GET` | `/api/admin/users` | Admin | Listar usuarios |
| `DELETE` | `/api/admin/users/:id` | Admin | Eliminar usuario y sus publicaciones |

La publicación recibe `multipart/form-data` con los campos `make`, `model`, `year`, `price_cop`, `description`, `cover_index` y uno o más archivos en `images`.

## Demostración

El flujo completo (registro, publicación con fotos, aprobación, contacto y venta) está descrito en [`docs/DEMO.md`](docs/DEMO.md).

## Diagramas

En `docs/plantuml/`: `estructura.puml` (diagrama de estructura) y `comportamiento.puml` (diagrama de comportamiento).

## Organización

- `cmd/web`: arranque del servidor. Si hay `PORT`, usa HTTP plano; si no, HTTPS con certificado autofirmado.
- `internal/httpapi`: rutas, autenticación, autorización y formularios multipart.
- `internal/service`: reglas del sistema.
- `internal/domain`: modelos de datos.
- `internal/store`: persistencia en PostgreSQL.
- `internal/media`: validación y almacenamiento de imágenes en PostgreSQL.
- `internal/audit`: registro de operaciones en base de datos.
- `web/templates`: páginas HTML.
- `web/static`: CSS y JavaScript.

## Despliegue en Render + Neon

Para desplegar en Render con base de datos en Neon:

**Build Command:**
```powershell
go build -tags netgo -ldflags '-s -w' -o app ./cmd/web
```

**Start Command:**
```
./app
```

**Variables de entorno:**
- `AUTOMARKET_DATABASE_URL`: Cadena de conexión a Neon con `sslmode=require` (usar URL directa, no pooler) o añadir `?default_query_exec_mode=simple_protocol`.
- `ADMIN_EMAIL`: Correo de la cuenta administrador.
- `ADMIN_PASSWORD`: Contraseña de la cuenta administrador.

**Notas:**
- El directorio `data/` no necesita estar configurado; las imágenes se guardan en PostgreSQL y la bitácora también.
- Las sesiones se mantienen en memoria y se pierden al reiniciar la aplicación.
- Git debe excluir `data/` pero no debe versionar certificados ni claves.

## Despliegue en Render + Neon

1. En Neon crea un proyecto y copia la connection string **directa** (no pooler). Debe terminar en `?sslmode=require`.
2. En Render crea un Web Service desde el repositorio:
   - Build Command: `go build -tags netgo -ldflags '-s -w' -o app ./cmd/web`
   - Start Command: `./app`
3. Variables de entorno:

| Variable | Uso |
| --- | --- |
| `AUTOMARKET_DATABASE_URL` | Conexión a PostgreSQL en Neon |
| `ADMIN_EMAIL` | Correo de la cuenta de administrador |
| `ADMIN_PASSWORD` | Contraseña del administrador (mínimo 8 caracteres) |

Render define `PORT` automáticamente; con esa variable la app escucha HTTP plano y Render entrega HTTPS por fuera. Las imágenes y la bitácora de auditoría se guardan en PostgreSQL. En el plan gratuito la app se duerme tras inactividad y las sesiones en memoria se pierden al reiniciar.
