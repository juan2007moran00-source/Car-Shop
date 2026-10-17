# Guion para demostrar AutoMarket

Abre dos ventanas de PowerShell. En la primera, desde la carpeta del proyecto, configura la cuenta administradora y deja el servidor en ejecución:

```powershell
cd "$HOME\Documents\Car-Shop"
$env:ADMIN_EMAIL = "admin@automarket.local"
$env:ADMIN_PASSWORD = "cambia-esta-clave-antes-de-usarla"
go run ./cmd/web
```

En la segunda ventana, registra un vendedor, inicia sesión y crea una publicación. Cambia el correo por uno que no esté guardado ya en `data/automarket.json`:

```powershell
$base = "http://localhost:8080"
$seller = Invoke-RestMethod -Method Post -Uri "$base/api/sellers" -ContentType "application/json" -Body (@{
    name = "Ana Pérez"; email = "ana@example.com"; phone = "+57 300 000 0000"; password = "clave-segura"
} | ConvertTo-Json)
$sellerSession = Invoke-RestMethod -Method Post -Uri "$base/api/sessions" -ContentType "application/json" -Body (@{
    email = "ana@example.com"; password = "clave-segura"
} | ConvertTo-Json)
$sellerHeaders = @{ Authorization = "Bearer $($sellerSession.token)" }
$vehicle = Invoke-RestMethod -Method Post -Uri "$base/api/vehicles" -Headers $sellerHeaders -ContentType "application/json" -Body (@{
    make = "Renault"; model = "Logan"; year = 2020; price_cop = 42000000; description = "Buen estado"
} | ConvertTo-Json)
$vehicle.id
```

La respuesta de publicación debe mostrar el estado `pending`. Antes de aprobarla, `GET http://localhost:8080/api/vehicles` no debe incluirla.

Inicia sesión como administrador y aprueba la publicación pendiente:

```powershell
$env:ADMIN_EMAIL = "admin@automarket.local"
$env:ADMIN_PASSWORD = "usa-aqui-la-misma-clave-de-la-primera-ventana"
$adminSession = Invoke-RestMethod -Method Post -Uri "$base/api/sessions" -ContentType "application/json" -Body (@{
    email = $env:ADMIN_EMAIL; password = $env:ADMIN_PASSWORD
} | ConvertTo-Json)
$adminHeaders = @{ Authorization = "Bearer $($adminSession.token)" }
$pending = Invoke-RestMethod -Method Get -Uri "$base/api/admin/vehicles/pending" -Headers $adminHeaders
$vehicleId = $pending.vehicles[0].id
Invoke-RestMethod -Method Post -Uri "$base/api/admin/vehicles/$vehicleId/approve" -Headers $adminHeaders
Invoke-RestMethod -Method Get -Uri "$base/api/vehicles"
```

El catálogo público debe mostrar el vehículo y el contacto de Ana. Para reportar la venta, vuelve a usar el token del vendedor:

```powershell
Invoke-RestMethod -Method Post -Uri "$base/api/vehicles/$vehicleId/sale" -Headers $sellerHeaders
Invoke-RestMethod -Method Get -Uri "$base/api/vehicles"
```

La publicación vendida deja de aparecer en el catálogo. En la carpeta `data`, cada petición queda registrada en la auditoría y el archivo de datos conserva vendedores y publicaciones después de reiniciar el servidor.
