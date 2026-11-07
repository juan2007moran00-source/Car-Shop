// Lógica del catálogo de vehículos, panel de vendedor y administración para AutoMarket
let allVehicles = [];
let adminAllVehicles = [];
let adminAllUsers = [];
let adminViewMode = 'grid'; // 'grid' | 'list'

document.addEventListener('DOMContentLoaded', () => {
    initAuthUI();
    loadCatalog();
    setupFilters();
    setupEscapeKey();

    // Abrir modal si se redirigió con ?auth=login o ?auth=register
    const urlParams = new URLSearchParams(window.location.search);
    if (urlParams.get('auth') === 'login') {
        openAuthModal('login');
    } else if (urlParams.get('auth') === 'register') {
        openAuthModal('register');
    }
});

// ==================== 1. AUTENTICACIÓN Y SESIÓN ====================

function getStoredAuth() {
    const token = localStorage.getItem('automarket_token') || localStorage.getItem('token') || sessionStorage.getItem('automarket-token');
    let user = null;
    try {
        const userRaw = localStorage.getItem('automarket_user') || localStorage.getItem('user');
        if (userRaw) user = JSON.parse(userRaw);
    } catch (_) {}
    return { token, user };
}

function saveAuth(token, user) {
    localStorage.setItem('automarket_token', token);
    localStorage.setItem('token', token);
    localStorage.setItem('automarket_user', JSON.stringify(user));
    localStorage.setItem('user', JSON.stringify(user));
}

function clearAuth() {
    localStorage.removeItem('automarket_token');
    localStorage.removeItem('token');
    localStorage.removeItem('automarket_user');
    localStorage.removeItem('user');
    sessionStorage.removeItem('automarket-token');
}

async function refreshUserProfile() {
    const { token } = getStoredAuth();
    if (!token) return;

    try {
        const res = await fetch('/api/me', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if (res.status === 401 || res.status === 403) {
            const data = await res.json().catch(() => ({}));
            clearAuth();
            initAuthUI();
            if (data.error && data.error.includes('bloqueada')) {
                alert('Tu cuenta ha sido bloqueada por un administrador.');
            }
            return;
        }
        if (res.ok) {
            const freshUser = await res.json();
            saveAuth(token, freshUser);
            renderUserBar(freshUser);
        }
    } catch (_) {}
}

function initAuthUI() {
    const { token, user } = getStoredAuth();
    const guestNav = document.getElementById('topbar-guest');
    const userNav = document.getElementById('topbar-user');

    if (token && user) {
        guestNav.style.display = 'none';
        userNav.style.display = 'flex';
        renderUserBar(user);
        refreshUserProfile();
    } else {
        guestNav.style.display = 'flex';
        userNav.style.display = 'none';
    }
}

function renderUserBar(user) {
    const userNameEl = document.getElementById('user-display-name');
    const userRoleEl = document.getElementById('user-display-role');
    const publishBtn = document.getElementById('btn-publish-vehicle');
    const requestsBtn = document.getElementById('btn-seller-requests');
    const warningsBtn = document.getElementById('btn-seller-warnings');
    const adminBtn = document.getElementById('btn-admin-panel');
    const warningsCountEl = document.getElementById('seller-warnings-count');

    userNameEl.textContent = user.name || 'Usuario';
    const role = (user.role || 'seller').toLowerCase();

    if (role === 'admin') {
        userRoleEl.textContent = 'ADMIN';
        userRoleEl.className = 'badge badge-admin';
        adminBtn.style.display = 'inline-flex';
        publishBtn.style.display = 'none';
        requestsBtn.style.display = 'none';
        warningsBtn.style.display = 'none';
    } else {
        userRoleEl.textContent = 'VENDEDOR';
        userRoleEl.className = 'badge badge-seller';
        adminBtn.style.display = 'none';
        publishBtn.style.display = 'inline-flex';
        requestsBtn.style.display = 'inline-flex';

        const warnings = Array.isArray(user.warnings) ? user.warnings : [];
        if (warnings.length > 0) {
            warningsBtn.style.display = 'inline-flex';
            warningsCountEl.textContent = warnings.length;
        } else {
            warningsBtn.style.display = 'none';
        }
    }
}

function logout() {
    clearAuth();
    initAuthUI();
    closeAllModals();
}

function closeAllModals() {
    closeAuthModal();
    closePublishModal();
    closeSellerRequestsModal();
    closeEditVehicleModal();
    closeAdminModal();
    closeWarnModal();
}

// ==================== 2. CATÁLOGO PÚBLICO ====================

async function loadCatalog() {
    const loadingEl = document.getElementById('catalog-loading');
    const emptyEl = document.getElementById('catalog-empty');
    const gridEl = document.getElementById('vehicles-grid');
    const countEl = document.getElementById('vehicles-count');

    loadingEl.style.display = 'block';
    emptyEl.style.display = 'none';
    gridEl.innerHTML = '';

    try {
        const response = await fetch('/api/vehicles');
        if (!response.ok) throw new Error('No se pudo obtener el catálogo.');
        
        const data = await response.json();
        allVehicles = Array.isArray(data.vehicles) ? data.vehicles : [];
        
        loadingEl.style.display = 'none';
        applyFiltersAndRender();
    } catch (err) {
        loadingEl.style.display = 'none';
        emptyEl.style.display = 'block';
        emptyEl.querySelector('h3').textContent = 'Error al cargar vehículos';
        emptyEl.querySelector('p').textContent = err.message || 'Intente nuevamente más tarde.';
        countEl.textContent = '0';
    }
}

function applyFiltersAndRender() {
    const searchInput = document.getElementById('search-input');
    const sortSelect = document.getElementById('sort-select');
    const countEl = document.getElementById('vehicles-count');
    const gridEl = document.getElementById('vehicles-grid');
    const emptyEl = document.getElementById('catalog-empty');

    const query = (searchInput ? searchInput.value : '').toLowerCase().trim();
    const sortMode = sortSelect ? sortSelect.value : 'recent';

    let filtered = allVehicles.filter(v => {
        const make = (v.make || '').toLowerCase();
        const model = (v.model || '').toLowerCase();
        const year = String(v.year || '');
        const desc = (v.description || '').toLowerCase();
        return make.includes(query) || model.includes(query) || year.includes(query) || desc.includes(query);
    });

    if (sortMode === 'price-asc') {
        filtered.sort((a, b) => a.price_cop - b.price_cop);
    } else if (sortMode === 'price-desc') {
        filtered.sort((a, b) => b.price_cop - a.price_cop);
    } else if (sortMode === 'year-desc') {
        filtered.sort((a, b) => b.year - a.year);
    } else {
        filtered.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
    }

    countEl.textContent = `${filtered.length} disponible${filtered.length === 1 ? '' : 's'}`;

    if (filtered.length === 0) {
        gridEl.innerHTML = '';
        emptyEl.style.display = 'block';
        if (query) {
            emptyEl.querySelector('h3').textContent = 'Sin resultados';
            emptyEl.querySelector('p').textContent = `No se encontraron vehículos que coincidan con "${query}".`;
        } else {
            emptyEl.querySelector('h3').textContent = 'No hay vehículos publicados';
            emptyEl.querySelector('p').textContent = 'Aún no hay publicaciones aprobadas en el catálogo. ¡Vuelve pronto o sé el primero en registrar un vehículo!';
        }
    } else {
        emptyEl.style.display = 'none';
        gridEl.innerHTML = filtered.map(renderVehicleCard).join('');
    }
}

function renderVehicleCard(v) {
    const formattedPrice = formatCOP(v.price_cop);
    const sellerName = escapeHTML(v.seller?.name || 'Vendedor particular');
    const sellerPhone = escapeHTML(v.seller?.phone || '');
    const sellerEmail = escapeHTML(v.seller?.email || '');
    const phoneClean = sellerPhone.replace(/[^0-9+]/g, '');

    const images = Array.isArray(v.images) && v.images.length > 0 ? v.images : ['/static/placeholder-car.svg'];
    const primaryImg = images[0];
    const photoCount = images.length;

    return `
        <article class="vehicle-card" id="vehicle-${v.id}">
            <div class="card-img-wrap" onclick="window.location.href='/vehiculo/${v.id}'" title="Ver detalle de ${escapeAttr(v.make)} ${escapeAttr(v.model)}">
                <img src="${primaryImg}" alt="${escapeAttr(v.make)} ${escapeAttr(v.model)}" onerror="this.src='/static/placeholder-car.svg'">
                ${photoCount > 1 ? `<span class="photo-count-tag">📷 ${photoCount} fotos</span>` : ''}
            </div>

            <div>
                <div class="card-top">
                    <h3 class="vehicle-name" style="cursor:pointer;" onclick="window.location.href='/vehiculo/${v.id}'">
                        ${escapeHTML(v.make)} ${escapeHTML(v.model)}
                    </h3>
                    <div style="display:flex; gap:4px;">
                        <span class="year-badge">${v.year}</span>
                    </div>
                </div>
                <div class="vehicle-price">${formattedPrice} COP</div>
                <p class="vehicle-desc">${escapeHTML(v.description || 'Sin descripción detallada.')}</p>
            </div>

            <div class="seller-card-box">
                <span class="seller-box-title">Contacto del vendedor</span>
                <div class="seller-box-name">${sellerName}</div>
                <div class="seller-links">
                    ${sellerPhone ? `
                        <a href="tel:${phoneClean}" class="seller-link-item" title="Llamar al vendedor">
                            <span>📞</span> ${sellerPhone}
                        </a>
                    ` : ''}
                    ${sellerEmail ? `
                        <a href="mailto:${sellerEmail}" class="seller-link-item" title="Enviar correo">
                            <span>✉️</span> ${sellerEmail}
                        </a>
                    ` : ''}
                </div>
            </div>

            <a href="/vehiculo/${v.id}" class="btn-view-detail">
                🔍 Ver detalles y ficha técnica →
            </a>
        </article>
    `;
}

function setupFilters() {
    const searchInput = document.getElementById('search-input');
    const sortSelect = document.getElementById('sort-select');
    if (searchInput) searchInput.addEventListener('input', applyFiltersAndRender);
    if (sortSelect) sortSelect.addEventListener('change', applyFiltersAndRender);
}

// ==================== 3. MODAL DE LOGIN / REGISTRO ====================

function openAuthModal(tab = 'login') {
    const modal = document.getElementById('auth-modal');
    modal.classList.add('active');
    switchAuthTab(tab);
    clearAlert('auth-alert');
}

function closeAuthModal() {
    const modal = document.getElementById('auth-modal');
    modal.classList.remove('active');
    clearAlert('auth-alert');
}

function switchAuthTab(tab) {
    const btnLogin = document.getElementById('tab-btn-login');
    const btnRegister = document.getElementById('tab-btn-register');
    const formLogin = document.getElementById('modal-form-login');
    const formRegister = document.getElementById('modal-form-register');
    clearAlert('auth-alert');

    if (tab === 'login') {
        btnLogin.classList.add('active');
        btnRegister.classList.remove('active');
        formLogin.style.display = 'grid';
        formRegister.style.display = 'none';
        document.getElementById('login-email').focus();
    } else {
        btnRegister.classList.add('active');
        btnLogin.classList.remove('active');
        formRegister.style.display = 'grid';
        formLogin.style.display = 'none';
        document.getElementById('reg-name').focus();
    }
}

async function handleModalLogin(e) {
    e.preventDefault();
    const btn = e.target.querySelector('button[type="submit"]');
    btn.disabled = true;
    btn.textContent = 'Ingresando...';
    clearAlert('auth-alert');

    const payload = {
        email: document.getElementById('login-email').value.trim(),
        password: document.getElementById('login-password').value
    };

    try {
        const res = await fetch('/api/sessions', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
        const data = await res.json();

        if (res.ok) {
            saveAuth(data.token, data.user);
            initAuthUI();
            showAlert('auth-alert', 'success', `¡Bienvenido, ${data.user.name}!`);
            setTimeout(() => {
                closeAuthModal();
                btn.disabled = false;
                btn.textContent = 'Ingresar';
                document.getElementById('modal-form-login').reset();
            }, 800);
        } else {
            showAlert('auth-alert', 'error', data.error || 'Correo o contraseña incorrectos.');
            btn.disabled = false;
            btn.textContent = 'Ingresar';
        }
    } catch (err) {
        showAlert('auth-alert', 'error', 'No se pudo conectar con el servidor.');
        btn.disabled = false;
        btn.textContent = 'Ingresar';
    }
}

async function handleModalRegister(e) {
    e.preventDefault();
    const btn = e.target.querySelector('button[type="submit"]');
    btn.disabled = true;
    btn.textContent = 'Creando cuenta...';
    clearAlert('auth-alert');

    const payload = {
        name: document.getElementById('reg-name').value.trim(),
        email: document.getElementById('reg-email').value.trim(),
        phone: document.getElementById('reg-phone').value.trim(),
        password: document.getElementById('reg-password').value
    };

    try {
        const res = await fetch('/api/sellers', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
        const data = await res.json();

        if (res.ok) {
            showAlert('auth-alert', 'success', '¡Cuenta creada con éxito! Iniciando sesión...');
            const loginRes = await fetch('/api/sessions', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email: payload.email, password: payload.password })
            });
            if (loginRes.ok) {
                const loginData = await loginRes.json();
                saveAuth(loginData.token, loginData.user);
                initAuthUI();
                setTimeout(() => {
                    closeAuthModal();
                    btn.disabled = false;
                    btn.textContent = 'Crear cuenta de vendedor';
                    document.getElementById('modal-form-register').reset();
                }, 1000);
            } else {
                switchAuthTab('login');
                showAlert('auth-alert', 'success', 'Cuenta creada. Por favor ingresa tus datos.');
                btn.disabled = false;
                btn.textContent = 'Crear cuenta de vendedor';
            }
        } else {
            showAlert('auth-alert', 'error', data.error || 'No se pudo completar el registro.');
            btn.disabled = false;
            btn.textContent = 'Crear cuenta de vendedor';
        }
    } catch (err) {
        showAlert('auth-alert', 'error', 'Error al conectar con el servidor.');
        btn.disabled = false;
        btn.textContent = 'Crear cuenta de vendedor';
    }
}

// ==================== 4. PUBLICAR VEHÍCULO ====================

function openPublishModal() {
    const modal = document.getElementById('publish-modal');
    modal.classList.add('active');
    clearAlert('publish-alert');
    document.getElementById('pub-make').focus();
}

function closePublishModal() {
    const modal = document.getElementById('publish-modal');
    modal.classList.remove('active');
    clearAlert('publish-alert');
}

function handleImageSelection(e) {
    const previewContainer = document.getElementById('pub-images-preview');
    previewContainer.innerHTML = '';
    const files = e.target.files;
    if (!files || files.length === 0) {
        previewContainer.style.display = 'none';
        return;
    }
    previewContainer.style.display = 'flex';
    for (const file of files) {
        const reader = new FileReader();
        reader.onload = (event) => {
            const img = document.createElement('img');
            img.src = event.target.result;
            img.className = 'preview-thumb';
            img.title = file.name;
            previewContainer.appendChild(img);
        };
        reader.readAsDataURL(file);
    }
}

async function handlePublishVehicle(e) {
    e.preventDefault();
    const { token } = getStoredAuth();
    if (!token) {
        alert('Debes iniciar sesión para publicar un vehículo.');
        closePublishModal();
        openAuthModal('login');
        return;
    }

    const btn = e.target.querySelector('button[type="submit"]');
    btn.disabled = true;
    btn.textContent = 'Subiendo fotografías...';
    clearAlert('publish-alert');

    const imageInput = document.getElementById('pub-images');
    if (!imageInput.files || imageInput.files.length === 0) {
        showAlert('publish-alert', 'error', 'Debes seleccionar al menos una fotografía del vehículo.');
        btn.disabled = false;
        btn.textContent = 'Enviar Publicación a Revisión';
        return;
    }

    // 1. Subir Fotografías al servidor
    let imageUrls = [];
    try {
        const formData = new FormData();
        for (const file of imageInput.files) {
            formData.append('images', file);
        }
        const uploadRes = await fetch('/api/uploads', {
            method: 'POST',
            headers: { 'Authorization': `Bearer ${token}` },
            body: formData
        });
        const uploadData = await uploadRes.json();
        if (!uploadRes.ok) {
            throw new Error(uploadData.error || 'No se pudieron subir las fotografías.');
        }
        imageUrls = uploadData.urls;
    } catch (uploadErr) {
        showAlert('publish-alert', 'error', uploadErr.message);
        btn.disabled = false;
        btn.textContent = 'Enviar Publicación a Revisión';
        return;
    }

    btn.textContent = 'Guardando vehículo y ficha técnica...';

    // 2. Construir payload con especificaciones técnicas
    const payload = {
        make: document.getElementById('pub-make').value.trim(),
        model: document.getElementById('pub-model').value.trim(),
        year: parseInt(document.getElementById('pub-year').value, 10),
        price_cop: parseInt(document.getElementById('pub-price').value, 10),
        description: document.getElementById('pub-desc').value.trim(),
        images: imageUrls,
        specs: {
            engine_displacement: document.getElementById('pub-engine').value.trim(),
            fuel_type: document.getElementById('pub-fuel').value.trim(),
            horsepower: parseInt(document.getElementById('pub-hp').value, 10),
            torque_nm: parseInt(document.getElementById('pub-torque').value, 10),
            transmission: document.getElementById('pub-transmission').value.trim(),
            body_type: document.getElementById('pub-body').value.trim(),
            doors: parseInt(document.getElementById('pub-doors').value, 10),
            mileage_km: parseInt(document.getElementById('pub-mileage').value, 10),
            color: document.getElementById('pub-color').value.trim(),
            traction: document.getElementById('pub-traction').value.trim()
        }
    };

    try {
        const res = await fetch('/api/vehicles', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify(payload)
        });
        const data = await res.json();

        if (res.ok) {
            showAlert('publish-alert', 'success', '¡Vehículo publicado con éxito con sus fotos y ficha técnica! Quedó en estado pendiente de aprobación por el administrador.');
            document.getElementById('form-publish-vehicle').reset();
            const previewEl = document.getElementById('pub-images-preview');
            if (previewEl) {
                previewEl.innerHTML = '';
                previewEl.style.display = 'none';
            }
            setTimeout(() => {
                closePublishModal();
                btn.disabled = false;
                btn.textContent = 'Enviar Publicación a Revisión';
            }, 2500);
        } else {
            showAlert('publish-alert', 'error', data.error || 'No se pudo registrar el vehículo.');
            btn.disabled = false;
            btn.textContent = 'Enviar Publicación a Revisión';
        }
    } catch (err) {
        showAlert('publish-alert', 'error', 'Error de conexión con el servidor.');
        btn.disabled = false;
        btn.textContent = 'Enviar Publicación a Revisión';
    }
}

// ==================== 5. MIS SOLICITUDES (VENDEDOR) ====================

function openSellerRequestsModal() {
    const modal = document.getElementById('seller-requests-modal');
    modal.classList.add('active');
    loadSellerRequests();
}

function closeSellerRequestsModal() {
    const modal = document.getElementById('seller-requests-modal');
    modal.classList.remove('active');
}

async function loadSellerRequests() {
    const { token, user } = getStoredAuth();
    if (!token) return;

    const loadingEl = document.getElementById('seller-requests-loading');
    const emptyEl = document.getElementById('seller-requests-empty');
    const listEl = document.getElementById('seller-requests-list');
    const warningsBanner = document.getElementById('seller-warnings-banner');
    const warningsList = document.getElementById('seller-warnings-list');

    loadingEl.style.display = 'block';
    emptyEl.style.display = 'none';
    listEl.innerHTML = '';

    // Renderizar advertencias recibidas del admin si existen
    if (user && Array.isArray(user.warnings) && user.warnings.length > 0) {
        warningsBanner.style.display = 'block';
        warningsList.innerHTML = user.warnings.map(w => `<li><strong>Advertencia:</strong> ${escapeHTML(w)}</li>`).join('');
    } else {
        warningsBanner.style.display = 'none';
    }

    try {
        const res = await fetch('/api/sellers/vehicles', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if (!res.ok) throw new Error('No se pudieron obtener las solicitudes.');

        const data = await res.json();
        const vehicles = Array.isArray(data.vehicles) ? data.vehicles : [];
        loadingEl.style.display = 'none';

        if (vehicles.length === 0) {
            emptyEl.style.display = 'block';
        } else {
            emptyEl.style.display = 'none';
            listEl.innerHTML = vehicles.map(renderSellerRequestCard).join('');
        }
    } catch (err) {
        loadingEl.style.display = 'none';
        emptyEl.style.display = 'block';
        emptyEl.querySelector('h3').textContent = 'Error al cargar';
        emptyEl.querySelector('p').textContent = err.message || 'Intente nuevamente más tarde.';
    }
}

function renderSellerRequestCard(v) {
    const statusLabels = {
        'pending': '<span class="status-pill pending">🟡 Pendiente de Aprobación</span>',
        'published': '<span class="status-pill published">🟢 Publicado en Catálogo</span>',
        'sold': '<span class="status-pill sold">⚪ Vendido</span>'
    };
    const statusPill = statusLabels[v.status] || `<span class="status-pill">${v.status}</span>`;
    const formattedPrice = formatCOP(v.price_cop);
    const dateFormatted = new Date(v.created_at).toLocaleDateString('es-CO', {
        year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit'
    });

    const isSold = v.status === 'sold';

    return `
        <article class="request-card" id="req-card-${v.id}">
            <div class="request-main">
                <div class="request-title-row">
                    <h3 class="request-title">${escapeHTML(v.make)} ${escapeHTML(v.model)} (${v.year})</h3>
                    ${statusPill}
                </div>
                <div class="request-price">${formattedPrice} COP</div>
                <p class="request-desc">${escapeHTML(v.description || 'Sin descripción.')}</p>
                <div class="request-date">Registrado el: ${dateFormatted}</div>
            </div>

            <div class="action-btns">
                ${!isSold ? `
                    <button type="button" class="btn-action btn-edit" onclick="openEditVehicleModal('${v.id}', '${escapeAttr(v.make)}', '${escapeAttr(v.model)}', ${v.year}, ${v.price_cop}, '${escapeAttr(v.description || '')}')">
                        ✏️ Editar Precio / Descripción
                    </button>
                ` : ''}

                ${v.status === 'published' ? `
                    <a href="/vehiculo/${v.id}" target="_blank" class="btn-action" style="background:#f1f5f9; color:#155eef; border-color:#cbd5e1;">
                        👁️ Ver en Catálogo
                    </a>
                    <button type="button" class="btn-action btn-sold" onclick="handleSellerReportSale('${v.id}')">
                        🤝 Marcar Vendido
                    </button>
                ` : ''}

                <button type="button" class="btn-action btn-delete" onclick="handleSellerDeleteVehicle('${v.id}', '${escapeAttr(v.make)} ${escapeAttr(v.model)}')">
                    🗑️ Eliminar Solicitud
                </button>
            </div>
        </article>
    `;
}

// 5.1 Editar Vehículo del Vendedor
function openEditVehicleModal(id, make, model, year, price, desc) {
    const modal = document.getElementById('edit-vehicle-modal');
    modal.classList.add('active');
    clearAlert('edit-vehicle-alert');

    document.getElementById('edit-vehicle-id').value = id;
    document.getElementById('edit-vehicle-subtitle').textContent = `Vehículo: ${make} ${model} (${year})`;
    document.getElementById('edit-pub-price').value = price;
    document.getElementById('edit-pub-desc').value = desc;
    document.getElementById('edit-pub-price').focus();
}

function closeEditVehicleModal() {
    const modal = document.getElementById('edit-vehicle-modal');
    modal.classList.remove('active');
    clearAlert('edit-vehicle-alert');
}

async function handleEditVehicleSubmit(e) {
    e.preventDefault();
    const { token } = getStoredAuth();
    if (!token) return;

    const id = document.getElementById('edit-vehicle-id').value;
    const price = parseInt(document.getElementById('edit-pub-price').value, 10);
    const desc = document.getElementById('edit-pub-desc').value.trim();

    const btn = e.target.querySelector('button[type="submit"]');
    btn.disabled = true;
    btn.textContent = 'Guardando...';
    clearAlert('edit-vehicle-alert');

    try {
        const res = await fetch(`/api/vehicles/${id}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({ price_cop: price, description: desc })
        });
        const data = await res.json();

        if (res.ok) {
            showAlert('edit-vehicle-alert', 'success', '¡Publicación actualizada con éxito!');
            setTimeout(() => {
                closeEditVehicleModal();
                btn.disabled = false;
                btn.textContent = 'Guardar Cambios';
                loadSellerRequests();
                loadCatalog(); // Actualizar catálogo si estaba publicado
            }, 900);
        } else {
            showAlert('edit-vehicle-alert', 'error', data.error || 'No se pudo actualizar.');
            btn.disabled = false;
            btn.textContent = 'Guardar Cambios';
        }
    } catch (err) {
        showAlert('edit-vehicle-alert', 'error', 'Error al conectar con el servidor.');
        btn.disabled = false;
        btn.textContent = 'Guardar Cambios';
    }
}

// 5.2 Eliminar Vehículo (Vendedor)
async function handleSellerDeleteVehicle(id, title) {
    if (!confirm(`¿Estás seguro de que deseas eliminar permanentemente la solicitud de "${title}"? Esta acción no se puede deshacer.`)) {
        return;
    }
    const { token } = getStoredAuth();
    if (!token) return;

    try {
        const res = await fetch(`/api/vehicles/${id}`, {
            method: 'DELETE',
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if (res.ok || res.status === 204) {
            loadSellerRequests();
            loadCatalog();
        } else {
            const data = await res.json().catch(() => ({}));
            alert(data.error || 'No se pudo eliminar la publicación.');
        }
    } catch (err) {
        alert('Error al conectar con el servidor.');
    }
}

// 5.3 Reportar Venta (Vendedor)
async function handleSellerReportSale(id) {
    if (!confirm('¿Deseas reportar este vehículo como vendido? Una vez marcado, ya no se mostrará en el catálogo público.')) {
        return;
    }
    const { token } = getStoredAuth();
    if (!token) return;

    try {
        const res = await fetch(`/api/vehicles/${id}/sale`, {
            method: 'POST',
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if (res.ok || res.status === 204) {
            loadSellerRequests();
            loadCatalog();
        } else {
            const data = await res.json().catch(() => ({}));
            alert(data.error || 'No se pudo registrar la venta.');
        }
    } catch (err) {
        alert('Error al conectar con el servidor.');
    }
}

// ==================== 6. PANEL DE ADMINISTRADOR ====================

function openAdminModal() {
    const modal = document.getElementById('admin-panel-modal');
    modal.classList.add('active');
    switchAdminTab('vehicles');
    loadAdminData();
    setupAdminSearchAndFilter();
}

function closeAdminModal() {
    const modal = document.getElementById('admin-panel-modal');
    modal.classList.remove('active');
}

function switchAdminTab(tab) {
    const tabVehicles = document.getElementById('admin-tab-btn-vehicles');
    const tabSellers = document.getElementById('admin-tab-btn-sellers');
    const secVehicles = document.getElementById('admin-sec-vehicles');
    const secSellers = document.getElementById('admin-sec-sellers');

    if (tab === 'vehicles') {
        tabVehicles.classList.add('active');
        tabSellers.classList.remove('active');
        secVehicles.style.display = 'block';
        secSellers.style.display = 'none';
    } else {
        tabSellers.classList.add('active');
        tabVehicles.classList.remove('active');
        secSellers.style.display = 'block';
        secVehicles.style.display = 'none';
    }
}

function setAdminViewMode(mode) {
    adminViewMode = mode;
    document.getElementById('admin-btn-grid').classList.toggle('active', mode === 'grid');
    document.getElementById('admin-btn-list').classList.toggle('active', mode === 'list');
    renderAdminVehicles();
}

async function loadAdminData() {
    const { token } = getStoredAuth();
    if (!token) return;

    const vehiclesLoading = document.getElementById('admin-vehicles-loading');
    const sellersLoading = document.getElementById('admin-sellers-loading');
    vehiclesLoading.style.display = 'block';
    sellersLoading.style.display = 'block';

    try {
        // 1. Cargar Todos los Vehículos
        const resVehicles = await fetch('/api/admin/vehicles', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if (resVehicles.ok) {
            const data = await resVehicles.json();
            adminAllVehicles = Array.isArray(data.vehicles) ? data.vehicles : [];
            document.getElementById('admin-count-vehicles').textContent = adminAllVehicles.length;
        }

        // 2. Cargar Todos los Usuarios / Vendedores
        const resUsers = await fetch('/api/admin/users', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if (resUsers.ok) {
            const data = await resUsers.json();
            adminAllUsers = Array.isArray(data.users) ? data.users : [];
            const sellersCount = adminAllUsers.filter(u => u.role === 'seller').length;
            document.getElementById('admin-count-sellers').textContent = sellersCount;
        }

        vehiclesLoading.style.display = 'none';
        sellersLoading.style.display = 'none';
        renderAdminVehicles();
        renderAdminSellers();
    } catch (err) {
        vehiclesLoading.style.display = 'none';
        sellersLoading.style.display = 'none';
        alert('Error al cargar datos del panel de administración.');
    }
}

function setupAdminSearchAndFilter() {
    const searchInput = document.getElementById('admin-search-vehicle');
    const filterSelect = document.getElementById('admin-filter-status');

    if (searchInput) {
        searchInput.oninput = () => renderAdminVehicles();
    }
    if (filterSelect) {
        filterSelect.onchange = () => renderAdminVehicles();
    }
}

function renderAdminVehicles() {
    const searchInput = document.getElementById('admin-search-vehicle');
    const filterSelect = document.getElementById('admin-filter-status');
    const contentEl = document.getElementById('admin-vehicles-content');
    const emptyEl = document.getElementById('admin-vehicles-empty');

    const query = (searchInput ? searchInput.value : '').toLowerCase().trim();
    const statusFilter = filterSelect ? filterSelect.value : 'all';

    let filtered = adminAllVehicles.filter(v => {
        const matchesStatus = (statusFilter === 'all') || (v.status === statusFilter);
        const make = (v.make || '').toLowerCase();
        const model = (v.model || '').toLowerCase();
        const year = String(v.year || '');
        const sellerName = (v.seller?.name || '').toLowerCase();
        const sellerEmail = (v.seller?.email || '').toLowerCase();

        const matchesQuery = !query || make.includes(query) || model.includes(query) || year.includes(query) ||
            sellerName.includes(query) || sellerEmail.includes(query);

        return matchesStatus && matchesQuery;
    });

    if (filtered.length === 0) {
        contentEl.innerHTML = '';
        emptyEl.style.display = 'block';
        return;
    }

    emptyEl.style.display = 'none';

    if (adminViewMode === 'grid') {
        // Vista Cuadrícula
        contentEl.innerHTML = `
            <div class="vehicles-grid">
                ${filtered.map(renderAdminVehicleGridCard).join('')}
            </div>
        `;
    } else {
        // Vista Lista / Tabla
        contentEl.innerHTML = `
            <div class="table-responsive">
                <table class="data-table">
                    <thead>
                        <tr>
                            <th>Vehículo</th>
                            <th>Año</th>
                            <th>Precio (COP)</th>
                            <th>Estado</th>
                            <th>Vendedor</th>
                            <th>Acciones</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${filtered.map(renderAdminVehicleTableRow).join('')}
                    </tbody>
                </table>
            </div>
        `;
    }
}

function renderAdminVehicleGridCard(v) {
    const statusPill = getStatusPill(v.status);
    const formattedPrice = formatCOP(v.price_cop);

    return `
        <article class="vehicle-card" style="border-left: 4px solid ${v.status === 'pending' ? '#eab308' : (v.status === 'published' ? '#22c55e' : '#94a3b8')};">
            <div>
                <div class="card-top">
                    <h3 class="vehicle-name">${escapeHTML(v.make)} ${escapeHTML(v.model)}</h3>
                    ${statusPill}
                </div>
                <div class="vehicle-price">${formattedPrice} COP</div>
                <p class="vehicle-desc">${escapeHTML(v.description || 'Sin descripción.')}</p>
            </div>

            <div class="seller-card-box" style="margin-bottom: 12px;">
                <span class="seller-box-title">Vendedor</span>
                <div class="seller-box-name">${escapeHTML(v.seller?.name || 'N/A')}</div>
                <div style="font-size: 0.82rem; color:#64748b;">${escapeHTML(v.seller?.email || '')} | ${escapeHTML(v.seller?.phone || '')}</div>
            </div>

            <div class="action-btns" style="justify-content: flex-end;">
                ${v.status === 'published' ? `
                    <a href="/vehiculo/${v.id}" target="_blank" class="btn-action" style="background:#f1f5f9; color:#155eef; border-color:#cbd5e1;">
                        👁️ Ver
                    </a>
                ` : ''}
                ${v.status === 'pending' ? `
                    <button type="button" class="btn-action btn-approve" onclick="handleAdminApproveVehicle('${v.id}')">
                        ✅ Aprobar Publicación
                    </button>
                ` : ''}
                <button type="button" class="btn-action btn-delete" onclick="handleAdminDeleteVehicle('${v.id}', '${escapeAttr(v.make)} ${escapeAttr(v.model)}')">
                    🗑️ Eliminar
                </button>
            </div>
        </article>
    `;
}

function renderAdminVehicleTableRow(v) {
    const statusPill = getStatusPill(v.status);
    const formattedPrice = formatCOP(v.price_cop);

    return `
        <tr>
            <td><strong>${escapeHTML(v.make)} ${escapeHTML(v.model)}</strong></td>
            <td>${v.year}</td>
            <td style="font-weight: 700; color: #155eef;">${formattedPrice}</td>
            <td>${statusPill}</td>
            <td>
                <div><strong>${escapeHTML(v.seller?.name || 'N/A')}</strong></div>
                <div style="font-size:0.8rem; color:#64748b;">${escapeHTML(v.seller?.email || '')}</div>
            </td>
            <td>
                <div class="action-btns">
                    ${v.status === 'published' ? `
                        <a href="/vehiculo/${v.id}" target="_blank" class="btn-action" style="background:#f1f5f9; color:#155eef; border-color:#cbd5e1;">
                            👁️ Ver
                        </a>
                    ` : ''}
                    ${v.status === 'pending' ? `
                        <button type="button" class="btn-action btn-approve" onclick="handleAdminApproveVehicle('${v.id}')">
                            ✅ Aprobar
                        </button>
                    ` : ''}
                    <button type="button" class="btn-action btn-delete" onclick="handleAdminDeleteVehicle('${v.id}', '${escapeAttr(v.make)} ${escapeAttr(v.model)}')">
                        🗑️ Eliminar
                    </button>
                </div>
            </td>
        </tr>
    `;
}

function renderAdminSellers() {
    const tbody = document.getElementById('admin-sellers-table-body');
    const sellers = adminAllUsers.filter(u => u.role === 'seller');

    if (sellers.length === 0) {
        tbody.innerHTML = `<tr><td colspan="6" style="text-align:center; padding: 24px; color:#64748b;">No hay vendedores registrados todavía.</td></tr>`;
        return;
    }

    tbody.innerHTML = sellers.map(u => {
        const isBlocked = !!u.is_blocked;
        const statusPill = isBlocked
            ? '<span class="status-pill blocked-user">🔴 Bloqueado</span>'
            : '<span class="status-pill active-user">🟢 Activo</span>';

        const warnings = Array.isArray(u.warnings) ? u.warnings : [];
        const warningsCount = warnings.length;

        return `
            <tr style="${isBlocked ? 'background: #fff5f5;' : ''}">
                <td>
                    <strong>${escapeHTML(u.name)}</strong>
                    <div style="font-size:0.75rem; color:#94a3b8;">ID: ${u.id.substring(0, 8)}...</div>
                </td>
                <td>
                    <div>${escapeHTML(u.email)}</div>
                    <div style="font-size:0.82rem; color:#64748b;">${escapeHTML(u.phone || 'Sin teléfono')}</div>
                </td>
                <td><span class="badge badge-seller">Vendedor</span></td>
                <td>${statusPill}</td>
                <td>
                    <span style="font-weight:700; color: ${warningsCount > 0 ? '#b45309' : '#64748b'};">
                        ${warningsCount} aviso${warningsCount === 1 ? '' : 's'}
                    </span>
                </td>
                <td>
                    <div class="action-btns">
                        <button type="button" class="btn-action btn-warn" onclick="openWarnUserModal('${u.id}', '${escapeAttr(u.name)}', '${escapeAttr(u.email)}')">
                            ⚠️ Notificar
                        </button>

                        ${isBlocked ? `
                            <button type="button" class="btn-action btn-unblock" onclick="handleToggleBlockUser('${u.id}', true)">
                                🟢 Desbloquear
                            </button>
                        ` : `
                            <button type="button" class="btn-action btn-block" onclick="handleToggleBlockUser('${u.id}', false)">
                                🚫 Bloquear
                            </button>
                        `}

                        <button type="button" class="btn-action btn-delete" onclick="handleAdminDeleteUser('${u.id}', '${escapeAttr(u.name)}')">
                            🗑️ Eliminar
                        </button>
                    </div>
                </td>
            </tr>
        `;
    }).join('');
}

// 6.1 Acciones Admin sobre Vehículos
async function handleAdminApproveVehicle(id) {
    const { token } = getStoredAuth();
    if (!token) return;

    try {
        const res = await fetch(`/api/admin/vehicles/${id}/approve`, {
            method: 'POST',
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if (res.ok || res.status === 204) {
            loadAdminData();
            loadCatalog();
        } else {
            const data = await res.json().catch(() => ({}));
            alert(data.error || 'No se pudo aprobar el vehículo.');
        }
    } catch (_) {
        alert('Error al conectar con el servidor.');
    }
}

async function handleAdminDeleteVehicle(id, title) {
    if (!confirm(`¿Eliminar la publicación de "${title}"? Se removerá del catálogo y del almacén.`)) return;
    const { token } = getStoredAuth();
    if (!token) return;

    try {
        const res = await fetch(`/api/admin/vehicles/${id}`, {
            method: 'DELETE',
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if (res.ok || res.status === 204) {
            loadAdminData();
            loadCatalog();
        } else {
            const data = await res.json().catch(() => ({}));
            alert(data.error || 'No se pudo eliminar el vehículo.');
        }
    } catch (_) {
        alert('Error al conectar con el servidor.');
    }
}

// 6.2 Acciones Admin sobre Vendedores
function openWarnUserModal(id, name, email) {
    const modal = document.getElementById('warn-user-modal');
    modal.classList.add('active');
    clearAlert('warn-user-alert');

    document.getElementById('warn-user-id').value = id;
    document.getElementById('warn-user-target').textContent = `Destinatario: ${name} (${email})`;
    document.getElementById('warn-user-message').value = '';
    document.getElementById('warn-user-message').focus();
}

function closeWarnModal() {
    const modal = document.getElementById('warn-user-modal');
    modal.classList.remove('active');
    clearAlert('warn-user-alert');
}

async function handleWarnUserSubmit(e) {
    e.preventDefault();
    const { token } = getStoredAuth();
    if (!token) return;

    const id = document.getElementById('warn-user-id').value;
    const message = document.getElementById('warn-user-message').value.trim();
    const btn = e.target.querySelector('button[type="submit"]');

    btn.disabled = true;
    btn.textContent = 'Enviando...';
    clearAlert('warn-user-alert');

    try {
        const res = await fetch(`/api/admin/users/${id}/warn`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({ message })
        });
        const data = await res.json();

        if (res.ok) {
            showAlert('warn-user-alert', 'success', '¡Notificación enviada con éxito al vendedor!');
            setTimeout(() => {
                closeWarnModal();
                btn.disabled = false;
                btn.textContent = 'Enviar Notificación Formal';
                loadAdminData();
            }, 900);
        } else {
            showAlert('warn-user-alert', 'error', data.error || 'No se pudo enviar la advertencia.');
            btn.disabled = false;
            btn.textContent = 'Enviar Notificación Formal';
        }
    } catch (_) {
        showAlert('warn-user-alert', 'error', 'Error al conectar con el servidor.');
        btn.disabled = false;
        btn.textContent = 'Enviar Notificación Formal';
    }
}

async function handleToggleBlockUser(id, currentlyBlocked) {
    const action = currentlyBlocked ? 'desbloquear' : 'bloquear';
    const msg = currentlyBlocked
        ? '¿Deseas desbloquear la cuenta de este vendedor para permitirle ingresar y publicar de nuevo?'
        : '¿Estás seguro de bloquear esta cuenta? El vendedor no podrá iniciar sesión ni publicar y sus sesiones activas se cerrarán inmediatamente.';

    if (!confirm(msg)) return;
    const { token } = getStoredAuth();
    if (!token) return;

    const endpoint = currentlyBlocked ? `/api/admin/users/${id}/unblock` : `/api/admin/users/${id}/block`;

    try {
        const res = await fetch(endpoint, {
            method: 'POST',
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if (res.ok) {
            loadAdminData();
        } else {
            const data = await res.json().catch(() => ({}));
            alert(data.error || `No se pudo ${action} la cuenta.`);
        }
    } catch (_) {
        alert('Error al conectar con el servidor.');
    }
}

async function handleAdminDeleteUser(id, name) {
    if (!confirm(`¿Eliminar al vendedor "${name}"? Esta acción borrará al usuario y TODOS sus vehículos registrados en cascada.`)) {
        return;
    }
    const { token } = getStoredAuth();
    if (!token) return;

    try {
        const res = await fetch(`/api/admin/users/${id}`, {
            method: 'DELETE',
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if (res.ok || res.status === 204) {
            loadAdminData();
            loadCatalog();
        } else {
            const data = await res.json().catch(() => ({}));
            alert(data.error || 'No se pudo eliminar el usuario.');
        }
    } catch (_) {
        alert('Error al conectar con el servidor.');
    }
}

// ==================== UTILIDADES ====================

function formatCOP(amount) {
    return new Intl.NumberFormat('es-CO', {
        style: 'currency',
        currency: 'COP',
        maximumFractionDigits: 0
    }).format(amount);
}

function getStatusPill(status) {
    switch (status) {
        case 'pending': return '<span class="status-pill pending">🟡 Pendiente</span>';
        case 'published': return '<span class="status-pill published">🟢 Publicado</span>';
        case 'sold': return '<span class="status-pill sold">⚪ Vendido</span>';
        default: return `<span class="status-pill">${status}</span>`;
    }
}

function showAlert(elementId, type, message) {
    const el = document.getElementById(elementId);
    if (!el) return;
    el.className = `alert-message ${type}`;
    el.textContent = message;
    el.style.display = 'block';
}

function clearAlert(elementId) {
    const el = document.getElementById(elementId);
    if (!el) return;
    el.style.display = 'none';
    el.textContent = '';
}

function escapeHTML(str) {
    return String(str || '')
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;');
}

function escapeAttr(str) {
    return String(str || '')
        .replace(/'/g, "\\'")
        .replace(/"/g, '&quot;');
}

function setupEscapeKey() {
    window.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            closeAllModals();
        }
    });
}
