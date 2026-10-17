# AutoMarket

Backend HTTP para consultar y administrar publicaciones de vehículos usados. El proyecto usa Go y Gin. Las demás funciones usan solo la biblioteca estándar de Go.

## Ejecutar

Requiere Go 1.26. Desde la raíz del repositorio, instala la dependencia permitida e inicia el servidor:

```powershell
go mod tidy
$env:ADMIN_EMAIL = "admin@automarket.local"
$env:ADMIN_PASSWORD = "una-clave-local-de-al-menos-8-caracteres"
go run ./cmd/web
```

La aplicación escucha en `http://localhost:8080`. Las credenciales del administrador se configuran con variables de entorno para no guardarlas en el código ni en Git. Si no se configuran, el servidor funciona sin una cuenta administradora y las rutas de administración no se pueden usar.

La aplicación guarda usuarios y vehículos en `data/automarket.json` y las operaciones en `data/audit.jsonl`. La auditoría encadena hashes SHA-256 para permitir detectar cambios en registros anteriores. Los datos y las credenciales no deben subirse al repositorio.

## Rutas

| Método y ruta | Acceso | Función |
| --- | --- | --- |
| `POST /api/sellers` | Público | Registrar una cuenta de vendedor |
| `POST /api/sessions` | Público | Iniciar sesión y recibir un token Bearer |
| `GET /api/vehicles` | Público | Consultar el catálogo publicado y los contactos |
| `GET /api/vehicles/:id` | Público | Consultar una publicación y el contacto del vendedor |
| `POST /api/vehicles` | Vendedor | Crear una publicación pendiente de aprobación |
| `POST /api/vehicles/:id/sale` | Vendedor propietario | Reportar la venta de una publicación publicada |
| `GET /api/admin/vehicles/pending` | Administrador | Consultar publicaciones pendientes |
| `POST /api/admin/vehicles/:id/approve` | Administrador | Aprobar una publicación pendiente |
| `DELETE /api/admin/vehicles/:id` | Administrador | Eliminar una publicación |
| `GET /api/admin/users` | Administrador | Consultar usuarios para moderación |
| `DELETE /api/admin/users/:id` | Administrador | Eliminar una cuenta y sus publicaciones |

Las rutas protegidas reciben `Authorization: Bearer <token>`. El registro crea únicamente vendedores; el rol no se acepta desde el cliente. Las contraseñas se guardan derivadas con PBKDF2-HMAC-SHA-256 y sal aleatoria. Los tokens de sesión viven en memoria y se invalidan al reiniciar el servidor.

## Ejemplos

Registro de vendedor:

```json
{
  "name": "Ana Pérez",
  "email": "ana@example.com",
  "phone": "+57 300 000 0000",
  "password": "clave-segura"
}
```

Inicio de sesión:

```json
{
  "email": "ana@example.com",
  "password": "clave-segura"
}
```

Publicación de vehículo, usando el token de vendedor:

```json
{
  "make": "Renault",
  "model": "Logan",
  "year": 2020,
  "price_cop": 42000000,
  "description": "Buen estado"
}
```

La publicación se crea como `pending`; solo aparece en el catálogo público después de la aprobación del administrador. Cuando el vendedor reporta la venta, pasa a `sold` y deja de aparecer en el catálogo.

## Estructura

- `cmd/web`: configuración y arranque del servidor.
- `internal/httpapi`: rutas, autenticación HTTP y respuestas.
- `internal/service`: reglas de negocio y manejo de sesiones.
- `internal/domain`: entidades y estados del sistema.
- `internal/store`: persistencia JSON local.
- `internal/audit`: registro trazable de operaciones.
- `web`: página de inicio y archivos estáticos.

## Alcance académico

El backend cubre el catálogo público, registro y publicación de vendedores, reporte de venta, moderación administrativa y el marco AAA solicitado. La página de inicio existente es una portada mínima; las funcionalidades del sistema se exponen por la API.
