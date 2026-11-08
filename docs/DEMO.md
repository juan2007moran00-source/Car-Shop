# Guion para demostrar AutoMarket

Abre una ventana de PowerShell en la raíz del proyecto y deja el servidor en ejecución. Necesitas una base PostgreSQL creada (por ejemplo, `automarket`):

```powershell
$env:AUTOMARKET_DATABASE_URL = "postgres://root:root@localhost:5432/automarket?sslmode=disable"
$env:ADMIN_EMAIL = "admin@automarket.local"
$env:ADMIN_PASSWORD = "cambia-esta-clave-antes-de-usarla"
go run ./cmd/web
```

Abre **https://localhost:8443** y acepta la advertencia del certificado autofirmado. La página muestra el catálogo público sin pedir registro.

La aplicación crea las tablas automáticamente. Puedes ver los datos en pgAdmin o con `psql` (tablas `users`, `vehicles` y `vehicle_images`).

## 1. Catálogo público y contacto (HU-VIS-01)

En la página principal, sin iniciar sesión:

1. Verifica que el catálogo carga y que el botón **Ingresar** es opcional.
2. Abre un vehículo publicado.
3. Comprueba que se muestran los datos del vehículo y el nombre, teléfono y correo del vendedor.

También puedes comprobarlo por API:

```powershell
$base = "https://localhost:8443"
curl.exe -k -s "$base/api/vehicles"
```

## 2. Registro y publicación con fotos (HU-VEN-01)

Registra un vendedor con un correo que no exista en la base de datos (tabla `users`):

```powershell
$seller = curl.exe -k -s -X POST "$base/api/sellers" `
  -H "Content-Type: application/json" `
  -d '{"name":"Ana Pérez","email":"ana@example.com","phone":"3001234567","password":"clave-segura"}' |
  ConvertFrom-Json

$session = curl.exe -k -s -X POST "$base/api/sessions" `
  -H "Content-Type: application/json" `
  -d '{"email":"ana@example.com","password":"clave-segura"}' |
  ConvertFrom-Json

$token = $session.token
```

Publica un vehículo con dos imágenes. Cambia las rutas por fotos reales en formato JPG o PNG. `cover_index=1` indica que la segunda imagen será la portada:

```powershell
$vehicle = curl.exe -k -s -X POST "$base/api/vehicles" `
  -H "Authorization: Bearer $token" `
  -F "make=Renault" `
  -F "model=Logan" `
  -F "year=2020" `
  -F "price_cop=42000000" `
  -F "description=Buen estado, único dueño." `
  -F "cover_index=1" `
  -F "images=@C:\ruta\foto-1.jpg;type=image/jpeg" `
  -F "images=@C:\ruta\foto-2.jpg;type=image/jpeg" |
  ConvertFrom-Json

$vehicle | Select-Object id, status, cover_image_id
$vehicle.images | Select-Object file_name, width, height, size_bytes
```

La respuesta debe mostrar `status: pending`. Antes de aprobarla, el catálogo público no debe incluirla:

```powershell
((curl.exe -k -s "$base/api/vehicles" | ConvertFrom-Json).vehicles).Count
```

Prueba también enviar un archivo que no sea JPG ni PNG (por ejemplo, un `.txt`). La API debe responder `400` con el motivo.

En la interfaz web, el formulario permite seleccionar varias imágenes, muestra una vista previa y deja marcar cuál será la portada.

## 3. Moderación (HU-ADM-01)

Inicia sesión como administrador:

```powershell
$admin = curl.exe -k -s -X POST "$base/api/sessions" `
  -H "Content-Type: application/json" `
  -d '{"email":"admin@automarket.local","password":"cambia-esta-clave-antes-de-usarla"}' |
  ConvertFrom-Json

$adminHeaders = "Authorization: Bearer $($admin.token)"
```

Consulta pendientes y aprueba la publicación:

```powershell
$pending = curl.exe -k -s "$base/api/admin/vehicles/pending" -H $adminHeaders | ConvertFrom-Json
$vehicleId = $pending.vehicles[0].id

curl.exe -k -s -X POST "$base/api/admin/vehicles/$vehicleId/approve" -H $adminHeaders
```

El catálogo público ahora debe incluir el vehículo con sus imágenes y el contacto de Ana:

```powershell
curl.exe -k -s "$base/api/vehicles"
```

En la interfaz, entra con la cuenta de administrador y usa las pestañas **Pendientes**, **Publicaciones** y **Usuarios** para aprobar o eliminar.

## 4. Reportar la venta

Vuelve a usar el token del vendedor:

```powershell
curl.exe -k -s -X POST "$base/api/vehicles/$vehicleId/sale" -H "Authorization: Bearer $token"
curl.exe -k -s "$base/api/vehicles"
```

La publicación vendida deja de aparecer en el catálogo y el vendedor la ve como **Vendida** en su panel. Los usuarios y vehículos se conservan en PostgreSQL entre reinicios; las imágenes y la bitácora de auditoría se guardan en la base de datos PostgreSQL.

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
