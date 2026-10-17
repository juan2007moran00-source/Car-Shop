# AutoMarket — tarjetas para GitHub Projects

Este archivo prepara las tarjetas del tablero Kanban. Al crear el Project en GitHub, usa las columnas `Pendiente`, `En curso`, `En revisión` y `Terminado`, y crea una tarjeta por cada fila pendiente.

## En revisión

- [ ] Confirmar con el grupo que los campos del vehículo (`make`, `model`, `year`, `price_cop`, `description`) coinciden con el diseño acordado.
- [ ] Revisar las rutas y respuestas JSON con el docente o el equipo.
- [ ] Ejecutar el flujo completo localmente y guardar evidencia para el video.

## Pendiente

- [ ] Crear el Project Kanban en GitHub y pasar estas tarjetas al tablero.
- [ ] Completar el diagrama estructural con la arquitectura implementada.
- [ ] Corregir el diagrama de comportamiento/casos de uso con notación UML válida.
- [ ] Grabar el video de demostración y sustentación.
- [ ] Preparar el PDF final con portada, integrantes, repositorio, hash del commit, invitación, enlace del video, diagramas y declaración de uso de IA.

## Implementación local lista para revisión

- [x] Separar arranque, HTTP, reglas de negocio, dominio, persistencia y auditoría.
- [x] Usar Gin y conservar Gin como única dependencia directa.
- [x] Exponer consulta pública de catálogo y contacto del vendedor.
- [x] Registrar vendedores, iniciar sesión y proteger operaciones por rol.
- [x] Crear publicaciones pendientes, aprobarlas, reportar ventas y administrar publicaciones/usuarios.
- [x] Persistir datos en JSON y registrar solicitudes en una cadena de auditoría verificable.

Las tarjetas completadas reflejan la implementación actual; las de revisión siguen pendientes hasta confirmar campos y demostrar los flujos de aceptación.
