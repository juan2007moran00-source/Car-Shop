#  Car Shop

Sistema web para la gestión y consulta de vehículos usados.

## Descripción

AutoMarket es un sitio web orientado a la publicación y consulta de
vehículos usados en venta. La plataforma cuenta con diferentes
funcionalidades dependiendo del rol del usuario.

El sistema contempla tres tipos de usuarios:

- Visitante
- Vendedor
- Administrador

## Roles

### Visitante
Puede consultar públicamente el catálogo de vehículos disponibles
y acceder a la información de contacto del vendedor, sin necesidad
de registrarse.

### Vendedor
Puede registrarse en la plataforma, publicar vehículos en venta
y reportar cuando estos hayan sido vendidos.

### Administrador
Puede autorizar las publicaciones realizadas por los vendedores
y eliminar publicaciones o usuarios cuando sea necesario.

## Tecnologías

- Go
- Gin
- HTTP
- Persistencia según la implementación del equipo

El proyecto utiliza únicamente las dependencias permitidas para
la entrega académica.

## Requerimientos principales

El sistema implementa el marco de seguridad AAA:

- Authentication: autenticación de los usuarios.
- Authorization: autorización de operaciones según el rol.
- Accounting: registro trazable de las operaciones realizadas.

## Estado del proyecto

Actualmente se encuentra en desarrollo.

### Implementado

- [ ] Estructura inicial del proyecto
- [ ] Rol Visitante
- [ ] Catálogo público de vehículos
- [ ] Información de contacto del vendedor
- [ ] Rol Vendedor
- [ ] Rol Administrador
- [ ] Seguridad AAA

## Proyecto académico

**Asignatura:** Laboratorio de Software III  
**Proyecto:** AutoMarket  
**Universidad:** Universidad de San Buenaventura - Cali
