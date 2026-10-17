# Diagramas de AutoMarket

Los diagramas están en Mermaid para que GitHub los renderice en este Markdown. Los campos del vehículo reflejan el modelo implementado y deben confirmarse con el equipo antes de cerrar la entrega.

## Estructura

```mermaid
flowchart LR
    visitante[Visitante]
    vendedor[Vendedor]
    administrador[Administrador]

    subgraph app[AutoMarket]
        main[cmd/web\nArranque y configuración]
        router[internal/httpapi\nGin, rutas, autenticación]
        service[internal/service\nReglas de negocio]
        domain[internal/domain\nUser, Vehicle, estados]
        store[internal/store\nPersistencia JSON]
        audit[internal/audit\nRegistro encadenado]
        archivos[(data/automarket.json)]
        bitacora[(data/audit.jsonl)]

        main --> router
        router --> service
        service --> domain
        service --> store
        store --> archivos
        router --> audit
        audit --> bitacora
    end

    visitante -->|consulta catálogo y contacto| router
    vendedor -->|registro, publicación y reporte de venta| router
    administrador -->|aprobación y eliminación| router
```

## Comportamiento: publicar, aprobar y consultar

```mermaid
sequenceDiagram
    actor V as Vendedor
    participant API as API Gin
    participant S as Servicio
    participant D as Archivo de datos
    actor A as Administrador
    actor X as Visitante
    participant L as Auditoría

    V->>API: POST /api/sellers
    API->>S: Validar y registrar vendedor
    S->>D: Guardar vendedor
    API->>L: Registrar operación
    API-->>V: 201 + id del vendedor

    V->>API: POST /api/sessions
    API->>S: Verificar credenciales
    API-->>V: Token Bearer

    V->>API: POST /api/vehicles + token
    API->>S: Crear publicación
    S->>D: Guardar estado pending
    API->>L: Registrar actor, recurso y resultado
    API-->>V: 201 + publicación pendiente

    A->>API: GET /api/admin/vehicles/pending + token
    API->>S: Listar pendientes
    S->>D: Leer publicaciones y vendedores
    API-->>A: Lista para moderación
    A->>API: POST /api/admin/vehicles/{id}/approve + token
    API->>S: Autorizar publicación
    S->>D: Cambiar estado a published
    API->>L: Registrar actor, recurso y resultado
    API-->>A: 204

    X->>API: GET /api/vehicles
    API->>S: Consultar publicaciones publicadas
    S->>D: Leer vehículo y contacto del vendedor
    API->>L: Registrar consulta pública
    API-->>X: Catálogo con datos de contacto
```
