# MEMORY.md — Historial de Modificaciones y Decisiones del Proyecto

> **Propósito de este archivo**: Registrar cronológicamente todos los cambios realizados en el proyecto **AutoMarket (`Car-Shop`)**, las decisiones técnicas tomadas y los archivos modificados. En cada cambio futuro, se debe añadir una nueva entrada al inicio de la sección **Historial de Cambios** (orden cronológico inverso: del más reciente al más antiguo).

---

## Plantilla para Futuros Cambios

Cuando se realice una modificación en el proyecto, copia y completa el siguiente bloque al principio del historial:

```markdown
### [AAAA-MM-DD] — Título breve del cambio
- **Tipo**: (Feature | Bugfix | Refactor | Docs | Test | Config)
- **Resumen**: Descripción clara de lo que se hizo y el motivo.
- **Archivos creados/modificados**:
  - `ruta/al/archivo.go`: Qué cambió exactamente.
- **Impacto en Contexto / API / Datos**: Indicar si se añadieron endpoints, cambiaron modelos JSON o si se actualizó `CONTEXT.md`.
- **Notas / Pendientes**: Cualquier consideración para próximas iteraciones.
```

---

## Historial de Cambios

### [2026-11-07] — Subida de fotografías obligatorias, ficha técnica exhaustiva y página de detalle con slider deslizable
- **Tipo**: Feature (Full-Stack)
- **Resumen**:
  1. Se implementó la subida de fotografías obligatorias (mínimo 1 imagen) almacenadas en `data/uploads/` y servidas por Go/Gin bajo `/uploads/*filepath`. Se creó el endpoint `POST /api/uploads` para recepción multipart con validación de tipo y tamaño (15MB por foto).
  2. Se agregó la estructura obligatoria de **Ficha Técnica** (`TechnicalSpecs`) validando que ninguno de sus 10 campos quede vacío o con valores numéricos inválidos: cilindrada del motor, combustible, potencia (HP), torque (Nm), transmisión, carrocería, puertas, kilometraje, color y tracción.
  3. Se creó la nueva página de detalle del automóvil (`GET /vehiculo/:id`, plantilla `web/templates/vehiculo.html` y script `web/static/vehiculo.js`):
     - **Slider deslizable interactivo de fotografías**: visualizador con controles previo/siguiente, contador, tira de miniaturas activas clicables, soporte táctil (swipe) y navegación por teclado (`←`/`→`).
     - **Ficha técnica visual**: cuadrícula estética con iconos para cada especificación.
     - **Caja de contacto**: llamada directa, correo y botón de WhatsApp con mensaje personalizado.
  4. Se actualizó el catálogo público (`inicio.html` y `catalogo.js`) para mostrar la fotografía principal en las tarjetas y enlazar con el botón "🔍 Ver detalles y ficha técnica →".
  5. Se actualizó el modal de publicación de vendedor para permitir previsualización en miniatura de fotos y captura de todos los campos de la ficha técnica.
- **Archivos creados/modificados**:
  - `internal/domain/models.go`: Incorporación de `TechnicalSpecs` y campos `Images` y `Specs` en `Vehicle` y `PublicVehicle`.
  - `internal/store/store.go`: Manejo de `Images` y `Specs` al persistir y nuevo método `UpdateVehicleDetails`.
  - `internal/service/service.go`: Validación obligatoria de fotos y función `validateSpecs`, actualización de `VehicleInput`, `PublishVehicle` y `withSeller`.
  - `internal/httpapi/router.go`: Registro de `data/uploads`, ruta `/vehiculo/:id`, endpoint `POST /api/uploads` y ampliación del límite de petición a 30 MB.
  - `internal/httpapi/handlers.go`: Controlador `uploadImages`, actualización de `publishVehicleRequest` y `updateVehicleRequest`.
  - `web/templates/vehiculo.html`: Plantilla HTML de detalle individual con slider y ficha técnica.
  - `web/static/vehiculo.js`: Lógica cliente del detalle, slider interactivo, miniaturas y contacto.
  - `web/static/placeholder-car.svg`: Gráfico vectorial de respaldo para vehículos sin foto.
  - `web/templates/inicio.html`: Formulario ampliado con carga de fotos, previsualización y campos técnicos obligatorios.
  - `web/static/estilos.css`: Estilos para slider, miniaturas, ficha técnica, migas de pan y fotos en tarjetas.
  - `web/static/catalogo.js`: Renderizado de tarjetas con portada y previsualización/subida de fotos en publicación.
  - `CONTEXT.md` y `MEMORY.md`: Actualización de arquitectura viva y bitácora.
- **Impacto en Contexto / API / Datos**: Nuevos endpoints `GET /vehiculo/:id` y `POST /api/uploads`, modelos de datos ampliados y servido de archivos estáticos bajo `/uploads/`.
- **Notas / Pendientes**: Las publicaciones previas sin fotos utilizan el gráfico SVG de respaldo `placeholder-car.svg`.

---

### [2026-11-07] — Gestión de solicitudes de vendedor, panel de administración completo y usuario administrador predeterminado
- **Tipo**: Feature (Full-Stack)
- **Resumen**: 
  1. Se implementó el apartado "Mis Solicitudes" para vendedores, permitiendo consultar todas sus solicitudes de publicación (`pending`, `published`, `sold`), modificar precio en COP y descripción en vivo mediante un modal de edición, retirar/eliminar solicitudes y reportar ventas.
  2. Se configuró un usuario administrador predeterminado (`admin@automarket.local` / `AdminAutoMarket2026*`) y se separaron sus credenciales en `ADMIN_CREDENTIALS.md` y `ADMIN_CREDENTIALS.txt`.
  3. Se creó el "Panel de Administración" completo accesible desde la web con:
     - Vista de publicaciones con selector de modo (Cuadrícula vs Tabla/Lista), buscador predictivo, filtrado por estado (`Todos`, `Pendientes`, `Publicados`, `Vendidos`), aprobación de vehículos pendientes y eliminación de publicaciones.
     - Vista de gestión de vendedores para supervisar conductas éticas: envío de notificaciones/advertencias formales por infracción de normas (`POST /api/admin/users/:id/warn`), bloqueo de cuentas con revocación inmediata de sesiones activas (`POST /api/admin/users/:id/block`), reactivación de cuentas (`/unblock`) y eliminación en cascada.
  4. Los vendedores con advertencias ahora ven una alerta y banner informativo en su panel de solicitudes.
- **Archivos creados/modificados**:
  - `ADMIN_CREDENTIALS.md` / `ADMIN_CREDENTIALS.txt`: Archivos separados con las credenciales e instrucciones de acceso del administrador.
  - `cmd/web/main.go`: Fallback automático al administrador predeterminado al arrancar el servidor.
  - `internal/domain/models.go`: Se añadieron campos `IsBlocked` (`bool`) y `Warnings` (`[]string`) a la estructura `User`.
  - `internal/store/store.go`: Nuevos métodos `AllVehicles`, `VehiclesBySeller`, `UpdateVehiclePriceAndDescription`, `DeleteSellerVehicle`, `SetUserBlocked` y `AddUserWarning`.
  - `internal/service/service.go`: Reglas de negocio para bloqueo de cuentas (purga de sesiones activas y restricción en `Login`/`Authenticate`), advertencias formales a vendedores, consulta de solicitudes por vendedor y actualización/eliminación de vehículos propios.
  - `internal/httpapi/router.go` y `handlers.go`: Endpoints `GET /api/me`, `GET /api/sellers/vehicles`, `PUT /api/vehicles/:id`, `DELETE /api/vehicles/:id`, `GET /api/admin/vehicles`, `POST /api/admin/users/:id/warn`, `POST /api/admin/users/:id/block`, `POST /api/admin/users/:id/unblock`.
  - `web/templates/inicio.html`: Modales para "Mis Solicitudes", edición de vehículo, "Panel de Administración" (pestañas de publicaciones y vendedores, selector de cuadrícula/lista) y modal para redactar advertencias a vendedores.
  - `web/static/estilos.css`: Estilos para modales grandes con scroll, tablas responsivas, selector de vista cuadrícula/lista, tarjetas de solicitudes, banners de advertencia y botones de acción.
  - `web/static/catalogo.js`: Lógica cliente completa para refresco de perfil, solicitudes de vendedor con edición/eliminación, panel admin con filtrado dinámico en cuadrícula/tabla, emisión de advertencias y bloqueo/desbloqueo de cuentas.
  - `README.md`, `CONTEXT.md` y `MEMORY.md`: Actualización de documentación, credenciales y arquitectura.
- **Impacto en Contexto / API / Datos**: Nuevos endpoints REST protegidos y ampliación del modelo `User` en `data/automarket.json`. Se mantiene plena compatibilidad con registros existentes.
- **Notas / Pendientes**: Todo el flujo de historias de usuario y requerimientos administrativos ahora es 100% operable desde la interfaz web sin requerir comandos de terminal.

---

### [2026-11-07] — Implementación del catálogo público de vehículos en la página principal
- **Tipo**: Feature / Frontend
- **Resumen**: Se rediseñó la vista principal (`GET /`, `web/templates/inicio.html`) para que sea la primera pantalla del sistema mostrando el catálogo público de vehículos usados para todo visitante (incluso no registrados). En la esquina superior derecha se ubicaron los botones de "Iniciar sesión" y "Registrarse", los cuales abren un modal interactivo con pestañas sin salir del catálogo. Se incorporaron tarjetas responsivas con información detallada del vehículo (marca, modelo, año, precio en COP, descripción) y datos de contacto directo del vendedor (teléfono y correo clicables), además de búsqueda y ordenamiento en tiempo real.
- **Archivos creados/modificados**:
  - `web/templates/inicio.html`: Vista principal con barra de navegación, botones superiores de acceso/registro, banner, buscador en vivo, cuadrícula de tarjetas de vehículos, y modales para autenticación y publicación de vehículos para vendedores.
  - `web/static/estilos.css`: Estilos visuales completos, barra superior, tarjetas de catálogo, botones de acción, estados de carga y vacíos, y modales accesibles.
  - `web/static/catalogo.js`: Lógica en JavaScript Vanilla para consultar `/api/vehicles`, renderizar tarjetas con formato de moneda en COP, filtrar en tiempo real, gestionar modales y sesiones en memoria/almacenamiento local.
  - `web/static/login.js`: Sincronización de credenciales en `localStorage` y redirección automática hacia el catálogo principal tras inicio de sesión exitoso.
  - `CONTEXT.md`: Actualización del estado actual de la interfaz y estructura de archivos estáticos.
  - `MEMORY.md`: Registro de la modificación.
- **Impacto en Contexto / API / Datos**: No hubo cambios en las rutas del backend ni en la API REST existente; se aprovechó el endpoint ya disponible `GET /api/vehicles`.
- **Notas / Pendientes**: En próximas fases se puede implementar un panel visual para que los administradores aprueben publicaciones pendientes directamente desde la web sin necesidad de PowerShell/curl.

---

### [2026-11-07] — Creación de archivos de gobernanza IA (`GEMINI.md`, `CONTEXT.md`, `MEMORY.md`)
- **Tipo**: Docs / Config
- **Resumen**: Se analizó la totalidad del repositorio AutoMarket (`cmd`, `internal`, `web`, `docs`, `openspec`) y se crearon los tres documentos base para guiar al asistente en futuras sesiones y llevar trazabilidad de todas las modificaciones.
- **Archivos creados/modificados**:
  - `GEMINI.md`: Reglas del proyecto, arquitectura en capas, restricciones de dependencias (solo Gin), seguridad (PBKDF2, tokens Bearer, cadena de auditoría SHA-256) e invariantes de negocio.
  - `CONTEXT.md`: Documentación viva del estado técnico actual, estructura de carpetas, variables de entorno, modelos de dominio, tabla completa de rutas HTTP/REST y limitaciones conocidas.
  - `MEMORY.md`: Bitácora histórica de cambios con plantilla estandarizada y línea base del proyecto.
- **Impacto en Contexto / API / Datos**: Ningún cambio funcional en el código Go ni en las plantillas web. Se establece la obligación de actualizar `CONTEXT.md` y `MEMORY.md` en cada cambio futuro.
- **Notas / Pendientes**:
  - Revisar tareas pendientes en `docs/KANBAN.md` (confirmación de campos con el equipo, pruebas de flujo completo y preparación de entregables).

---

### [Línea Base] — Estado inicial del repositorio AutoMarket
- **Tipo**: Feature (Base)
- **Resumen**: Implementación inicial funcional del sistema AutoMarket en Go 1.26 con Gin.
- **Componentes existentes en la línea base**:
  - `cmd/web/main.go`: Arranque del servidor, carga de persistencia, auditoría y creación opcional de administrador vía `ADMIN_EMAIL` y `ADMIN_PASSWORD`.
  - `internal/domain/models.go`: Modelos `User`, `Vehicle`, `PublicVehicle`, `SellerContact` y estados (`pending`, `published`, `sold`).
  - `internal/service/service.go`: Lógica de negocio, hashing PBKDF2-SHA256 (600,000 iteraciones), sesiones en memoria y validaciones.
  - `internal/store/store.go`: Persistencia JSON atómica en `data/automarket.json` con `sync.RWMutex` y borrado en cascada de vehículos al eliminar un usuario.
  - `internal/audit/audit.go`: Bitácora append-only en `data/audit.jsonl` con verificación de cadena de hashes SHA-256.
  - `internal/httpapi/`: Rutas públicas, de vendedor y de administrador, más middlewares de autenticación, rol, auditoría y límite de tamaño de petición (1 MB).
  - `web/`: Plantillas `inicio.html` (portada, login/registro y panel básico de publicación para vendedores) y `login.html` con recursos estáticos en `web/static/`.
  - `docs/`: Guion de demostración (`DEMO.md`), diagramas C4/UML en Mermaid y PlantUML (`DIAGRAMAS.md`, `plantuml/`) y tablero de tareas (`KANBAN.md`).

