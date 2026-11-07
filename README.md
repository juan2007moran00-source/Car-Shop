# AutoMarket

Página web sencilla para consultar carros y una API hecha en Go para registrar vendedores, publicar vehículos y administrarlos.

## Iniciar el servidor

Necesitas tener Go instalado. Abre PowerShell en la carpeta `Car-Shop` y ejecuta:

```powershell
go run ./cmd/web
```

Abre [http://localhost:8080](http://localhost:8080). La pantalla de acceso está en [http://localhost:8080/login](http://localhost:8080/login).

Para detener el servidor, vuelve a PowerShell y presiona `Ctrl + C`.

## Administrador

El sistema cuenta con un usuario administrador predeterminado cuyas credenciales se encuentran en el archivo separado [`ADMIN_CREDENTIALS.md`](./ADMIN_CREDENTIALS.md) y [`ADMIN_CREDENTIALS.txt`](./ADMIN_CREDENTIALS.txt):

- **Correo**: `admin@automarket.local`
- **Contraseña**: `AdminAutoMarket2026*`

Si deseas personalizar las credenciales del administrador antes de iniciar Go:

```powershell
$env:ADMIN_EMAIL = "admin@automarket.local"
$env:ADMIN_PASSWORD = "escribe-una-clave-de-al-menos-8-caracteres"
go run ./cmd/web
```

## Organización

- `cmd/web`: inicia el servidor.
- `internal/httpapi`: rutas y comunicación web.
- `internal/service`: reglas del sistema.
- `internal/domain`: modelos de datos.
- `internal/store`: guardado local.
- `internal/audit`: registro de operaciones.
- `web/templates`: páginas HTML.
- `web/static`: CSS y JavaScript.

Los datos locales se guardan en `data/` y Git los excluye.
