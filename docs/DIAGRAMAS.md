# Diagramas de AutoMarket

Estos diagramas describen el alcance del proyecto de curso y la implementación que existe en el repositorio. GitHub puede renderizar los bloques Mermaid de este archivo.

## 1. Contexto C4

Muestra quién usa AutoMarket y para qué. En este nivel no se muestran paquetes ni archivos internos.

```mermaid
flowchart LR
    visitor["Visitante\nPersona no autenticada"]
    seller["Vendedor\nPersona registrada"]
    admin["Administrador\nPersona autorizada"]
    system["AutoMarket\nSistema web para consultar el catálogo, publicar vehículos y moderar publicaciones"]

    visitor -->|Consulta vehículos y contacto del vendedor| system
    seller -->|Se registra, publica vehículos y reporta ventas| system
    admin -->|Revisa pendientes y administra publicaciones o usuarios| system
```

## 2. Contenedores C4: separación estructural

El sistema se separa en interfaz web, backend y dos almacenes persistentes. La interfaz corre en el navegador; el backend Go/Gin sirve las páginas y la API HTTP. Los archivos JSON son almacenes de datos, no componentes de Go.

```mermaid
flowchart LR
    visitor["Visitante"]
    seller["Vendedor"]
    admin["Administrador"]

    subgraph automarket["Sistema AutoMarket"]
        direction LR
        web["Interfaz web\nContenedor: HTML, CSS y JavaScript\nEjecutado en el navegador"]
        api["Backend HTTP / API REST\nContenedor: Go + Gin\nSirve páginas, aplica reglas y controla accesos"]
        data[("Almacén de datos\nContenedor: data/automarket.json\nUsuarios y vehículos")]
        audit[("Bitácora de auditoría\nContenedor: data/audit.jsonl\nEventos encadenados con SHA-256")]

        api -->|Lee y guarda JSON| data
        api -->|Registra cada solicitud HTTP| audit
        api -->|Entrega HTML, CSS y JavaScript| web
    end

    visitor -->|Usa catálogo público| web
    seller -->|Usa registro, acceso y publicación| web
    web -->|Solicitudes HTTP y JSON| api
    admin -->|Solicitudes REST protegidas con token| api
```

### Responsabilidades que se separan

- **Interfaz web:** presenta la portada, registro, inicio de sesión y formulario del vendedor.
- **Backend Go/Gin:** contiene las rutas HTTP, autenticación, autorización por rol y reglas de negocio.
- **Almacén JSON:** conserva usuarios, vehículos y sus estados entre reinicios.
- **Bitácora JSONL:** guarda las solicitudes y enlaza eventos para detectar modificaciones.

**Límite actual:** el repositorio no tiene un panel visual de administración. El administrador usa las rutas REST protegidas, por ejemplo con un cliente HTTP. El diagrama no representa una pantalla que todavía no existe.

## 3. Componentes del backend Go

Este diagrama amplía el contenedor del backend. Los paquetes son componentes internos; no son contenedores independientes ni procesos separados.

```mermaid
flowchart LR
    client["Navegador o cliente HTTP"]

    subgraph backend["Backend AutoMarket: un proceso Go/Gin"]
        direction LR
        routes["Rutas y handlers\ninternal/httpapi\nInterpreta HTTP y crea respuestas"]
        auth["Autenticación y autorización\ninternal/httpapi\nValida token y rol"]
        service["Reglas de negocio\ninternal/service\nRegistro, sesión, publicación y moderación"]
        domain["Modelos del dominio\ninternal/domain\nUser, Vehicle y estados"]
        store["Persistencia\ninternal/store\nLee y escribe JSON"]
        logger["Auditoría\ninternal/audit\nEncadena eventos con SHA-256"]
        files[("data/automarket.json")]
        logs[("data/audit.jsonl")]

        routes --> auth
        auth --> service
        service --> domain
        service --> store
        store --> files
        routes -->|Middleware de accounting| logger
        logger --> logs
    end

    client -->|HTTP| routes
```

## 4. Diagrama de clases del dominio

Representa los modelos que aparecen en `internal/domain/models.go`. `PublicVehicle` es la vista pública de un vehículo e incluye los datos de contacto del vendedor.

```mermaid
classDiagram
    class Role {
        <<enumeration>>
        seller
        admin
    }

    class VehicleStatus {
        <<enumeration>>
        pending
        published
        sold
    }

    class User {
        +string ID
        +string Name
        +string Email
        +string Phone
        +Role Role
        +string PasswordSalt
        +string PasswordHash
        +time CreatedAt
    }

    class Vehicle {
        +string ID
        +string Make
        +string Model
        +int Year
        +int64 PriceCOP
        +string Description
        +string SellerID
        +VehicleStatus Status
        +time CreatedAt
    }

    class PublicVehicle {
        +string ID
        +string Make
        +string Model
        +int Year
        +int64 PriceCOP
        +string Description
        +VehicleStatus Status
        +time CreatedAt
    }

    class SellerContact {
        +string Name
        +string Email
        +string Phone
    }

    User --> Role : tiene
    Vehicle --> VehicleStatus : tiene
    User "1 vendedor" --> "0..* vehículos" Vehicle : SellerID
    Vehicle ..> PublicVehicle : se convierte en respuesta pública
    PublicVehicle *-- SellerContact : incluye contacto
```

## 5. Casos de uso

Los casos se limitan a las historias de usuario del curso. No se añaden relaciones `include` o `extend` porque los requisitos no definen esos vínculos obligatorios.

```mermaid
flowchart LR
    visitor["Visitante"]
    seller["Vendedor"]
    admin["Administrador"]

    subgraph system["AutoMarket"]
        catalog([Consultar catálogo])
        contact([Consultar contacto del vendedor])
        register([Registrarse como vendedor])
        login([Iniciar sesión])
        publish([Publicar vehículo])
        report([Reportar venta])
        pending([Consultar publicaciones pendientes])
        approve([Autorizar publicación])
        deleteVehicle([Eliminar publicación])
        deleteUser([Eliminar usuario])
    end

    visitor --- catalog
    visitor --- contact
    seller --- register
    seller --- login
    seller --- publish
    seller --- report
    admin --- login
    admin --- pending
    admin --- approve
    admin --- deleteVehicle
    admin --- deleteUser
```

## 6. Secuencia de comportamiento: publicar y aprobar

Flujo principal de las historias de usuario. La aprobación requiere que el administrador se autentique antes de invocar la ruta protegida.

```mermaid
sequenceDiagram
    actor V as Vendedor
    participant UI as Interfaz web
    participant API as API Go/Gin
    participant S as Servicio
    participant D as automarket.json
    participant L as audit.jsonl
    actor A as Administrador
    actor X as Visitante

    V->>UI: Completa registro
    UI->>API: POST /api/sellers
    API->>S: Validar y registrar vendedor
    S->>D: Guardar usuario
    API->>L: Registrar solicitud
    API-->>UI: 201 Created

    V->>UI: Ingresa correo y contraseña
    UI->>API: POST /api/sessions
    API->>S: Verificar credenciales
    S->>D: Consultar usuario
    API-->>UI: Token Bearer y rol

    V->>UI: Envía datos del vehículo
    UI->>API: POST /api/vehicles + token
    API->>API: Autenticar token y verificar rol seller
    API->>S: Validar y crear publicación
    S->>D: Guardar estado pending
    API->>L: Registrar actor y resultado
    API-->>UI: 201 Created + publicación pendiente

    A->>API: POST /api/sessions
    API->>S: Verificar credenciales de administrador
    API-->>A: Token Bearer y rol admin
    A->>API: GET /api/admin/vehicles/pending + token
    API->>API: Verificar rol admin
    API->>S: Consultar pendientes
    S->>D: Leer vehículos y vendedores
    API-->>A: Lista de publicaciones pendientes
    A->>API: POST /api/admin/vehicles/{id}/approve + token
    API->>S: Aprobar publicación
    S->>D: Cambiar estado a published
    API->>L: Registrar actor y resultado
    API-->>A: 204 No Content

    X->>UI: Abre catálogo
    UI->>API: GET /api/vehicles
    API->>S: Consultar vehículos publicados
    S->>D: Leer vehículos y vendedores
    API->>L: Registrar consulta pública
    API-->>UI: Catálogo y contacto del vendedor
    UI-->>X: Mostrar oferta y contacto

    V->>UI: Reporta la venta
    UI->>API: POST /api/vehicles/{id}/sale + token
    API->>API: Autenticar token y verificar rol seller
    API->>S: Marcar venta del vendedor
    S->>D: Cambiar estado a sold
    API->>L: Registrar actor y resultado
    API-->>UI: 204 No Content
```

## Diagramas para la entrega

La guía pide un diagrama de estructura y uno de comportamiento en el PDF. Para esa entrega se recomienda usar **Contenedores C4** (sección 2) como estructura y **Secuencia publicar y aprobar** (sección 6) como comportamiento. El diagrama de contexto sirve para explicar el sistema y el resto queda como apoyo para la sustentación.
