# Credenciales de Administrador — AutoMarket

Este archivo contiene las credenciales de la cuenta administradora configurada en el sistema para pruebas, demostración y gestión de la plataforma.

---

## Datos de Acceso del Administrador

- **Correo electrónico**: `admin@automarket.local`
- **Contraseña**: `AdminAutoMarket2026*`
- **Rol**: `admin`

---

## Cómo Iniciar Sesión como Administrador

1. Inicia el servidor de AutoMarket desde la terminal o PowerShell:
   ```powershell
   go run ./cmd/web
   ```
   *(El servidor asegura automáticamente la existencia de esta cuenta administradora al iniciar).*

2. Abre en tu navegador:
   [http://localhost:8080](http://localhost:8080)

3. En la esquina superior derecha, haz clic en el botón **"Iniciar sesión"**.

4. Ingresa las credenciales:
   - **Correo**: `admin@automarket.local`
   - **Contraseña**: `AdminAutoMarket2026*`

5. Al autenticarte como administrador:
   - Verás la insignia roja **ADMIN** en la barra superior.
   - Tendrás acceso al botón **"Panel de Administración"**.
   - Podrás ver todas las publicaciones en modo **Cuadrícula** o **Lista / Tabla**.
   - Podrás filtrar los vehículos por estado (`Pendientes`, `Publicados`, `Vendidos`).
   - Podrás aprobar publicaciones pendientes y eliminar cualquier vehículo.
   - Podrás gestionar a los vendedores: enviarles **notificaciones / advertencias** por conductas contrarias a las normas y **bloquear/desbloquear** sus cuentas.

