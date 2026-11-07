// Lógica de la página de detalle del vehículo y slider deslizable de fotografías
let currentVehicle = null;
let currentSlideIndex = 0;
let vehicleImages = [];

document.addEventListener('DOMContentLoaded', () => {
    initAuthBar();
    loadVehicleDetail();
    setupKeyboardNavigation();
    setupTouchSwipe();
});

// 1. Barra de Autenticación
function initAuthBar() {
    const token = localStorage.getItem('automarket_token') || localStorage.getItem('token') || sessionStorage.getItem('automarket-token');
    let user = null;
    try {
        const userRaw = localStorage.getItem('automarket_user') || localStorage.getItem('user');
        if (userRaw) user = JSON.parse(userRaw);
    } catch (_) {}

    const guestNav = document.getElementById('topbar-guest');
    const userNav = document.getElementById('topbar-user');

    if (token && user) {
        guestNav.style.display = 'none';
        userNav.style.display = 'flex';
        document.getElementById('user-display-name').textContent = user.name || 'Usuario';
        const role = (user.role || 'seller').toLowerCase();
        const roleBadge = document.getElementById('user-display-role');
        roleBadge.textContent = role === 'admin' ? 'ADMIN' : 'VENDEDOR';
        roleBadge.className = `badge ${role === 'admin' ? 'badge-admin' : 'badge-seller'}`;
    } else {
        guestNav.style.display = 'flex';
        userNav.style.display = 'none';
    }
}

function logout() {
    localStorage.removeItem('automarket_token');
    localStorage.removeItem('token');
    localStorage.removeItem('automarket_user');
    localStorage.removeItem('user');
    sessionStorage.removeItem('automarket-token');
    window.location.reload();
}

function openAuthModal(tab) {
    window.location.href = `/?auth=${tab}`;
}

// 2. Carga del Vehículo
async function loadVehicleDetail() {
    const loadingEl = document.getElementById('detail-loading');
    const errorEl = document.getElementById('detail-error');
    const contentEl = document.getElementById('detail-content');

    const pathSegments = window.location.pathname.split('/').filter(Boolean);
    const vehicleId = pathSegments[pathSegments.length - 1];

    if (!vehicleId) {
        showError('Identificador inválido', 'No se especificó un vehículo válido para consultar.');
        return;
    }

    try {
        const response = await fetch(`/api/vehicles/${vehicleId}`);
        if (!response.ok) {
            if (response.status === 404) {
                showError('Vehículo no encontrado', 'Esta publicación no existe o fue vendida y retirada del catálogo.');
            } else {
                throw new Error('Error al consultar el servidor.');
            }
            return;
        }

        currentVehicle = await response.json();
        renderVehicle(currentVehicle);

        loadingEl.style.display = 'none';
        errorEl.style.display = 'none';
        contentEl.style.display = 'block';
    } catch (err) {
        showError('Error de conexión', err.message || 'No se pudo conectar con el servidor.');
    }
}

function showError(title, msg) {
    document.getElementById('detail-loading').style.display = 'none';
    const errorEl = document.getElementById('detail-error');
    document.getElementById('detail-error-title').textContent = title;
    document.getElementById('detail-error-msg').textContent = msg;
    errorEl.style.display = 'block';
}

// 3. Renderizado de Datos y Ficha Técnica
function renderVehicle(v) {
    const title = `${v.make} ${v.model}`;
    document.title = `${title} (${v.year}) | AutoMarket`;
    document.getElementById('breadcrumb-vehicle-name').textContent = `${title} (${v.year})`;
    document.getElementById('detail-vehicle-title').textContent = title;
    document.getElementById('detail-badge-year').textContent = v.year;
    document.getElementById('detail-vehicle-price').textContent = `${formatCOP(v.price_cop)} COP`;

    const specs = v.specs || {};
    document.getElementById('detail-badge-body').textContent = specs.body_type || 'Vehículo';

    // Quick specs en la cabecera
    const kmFormatted = (specs.mileage_km !== undefined && specs.mileage_km !== null) 
        ? `${specs.mileage_km.toLocaleString('es-CO')} km` 
        : 'Kilometraje N/D';
    document.getElementById('quick-spec-km').textContent = `🛣️ ${kmFormatted}`;
    document.getElementById('quick-spec-fuel').textContent = `⛽ ${specs.fuel_type || 'Combustible N/D'}`;
    document.getElementById('quick-spec-trans').textContent = `🕹️ ${specs.transmission || 'Transmisión N/D'}`;

    // Descripción
    document.getElementById('detail-description-text').textContent = v.description || 'Sin descripción detallada disponible.';

    // Contacto del Vendedor
    const seller = v.seller || {};
    document.getElementById('detail-seller-name').textContent = seller.name || 'Vendedor Particular';

    const phoneClean = (seller.phone || '').replace(/[^0-9+]/g, '');
    const phoneBtn = document.getElementById('btn-seller-phone');
    const phoneSpan = document.getElementById('detail-seller-phone');
    if (phoneClean) {
        phoneBtn.href = `tel:${phoneClean}`;
        phoneSpan.textContent = `Llamar: ${seller.phone}`;
        phoneBtn.style.display = 'inline-flex';
    } else {
        phoneBtn.style.display = 'none';
    }

    const emailBtn = document.getElementById('btn-seller-email');
    const emailSpan = document.getElementById('detail-seller-email');
    if (seller.email) {
        emailBtn.href = `mailto:${seller.email}?subject=Consulta%20por%20${encodeURIComponent(title + ' en AutoMarket')}`;
        emailSpan.textContent = `Correo: ${seller.email}`;
        emailBtn.style.display = 'inline-flex';
    } else {
        emailBtn.style.display = 'none';
    }

    // Botón de WhatsApp
    const waBtn = document.getElementById('btn-seller-whatsapp');
    if (phoneClean) {
        const waNumber = phoneClean.replace(/\+/g, '');
        const message = `Hola ${seller.name || ''}, vi tu publicación de ${title} (${v.year}) en AutoMarket y me gustaría recibir más información.`;
        waBtn.href = `https://wa.me/${waNumber}?text=${encodeURIComponent(message)}`;
        waBtn.style.display = 'inline-flex';
    } else {
        waBtn.style.display = 'none';
    }

    // Slider de Fotos
    vehicleImages = Array.isArray(v.images) && v.images.length > 0 ? v.images : ['/static/placeholder-car.svg'];
    setupSlider(vehicleImages);

    // Ficha Técnica Exhaustiva
    renderTechnicalSpecs(specs);
}

function renderTechnicalSpecs(specs) {
    const grid = document.getElementById('specs-grid');
    const specItems = [
        { label: 'Cilindrada del Motor', value: specs.engine_displacement || 'N/D', icon: '⚙️' },
        { label: 'Tipo de Combustible', value: specs.fuel_type || 'N/D', icon: '⛽' },
        { label: 'Potencia (Caballos de Fuerza)', value: specs.horsepower ? `${specs.horsepower} HP` : 'N/D', icon: '🐎' },
        { label: 'Torque del Motor', value: specs.torque_nm ? `${specs.torque_nm} Nm` : 'N/D', icon: '🔄' },
        { label: 'Tipo de Transmisión', value: specs.transmission || 'N/D', icon: '🕹️' },
        { label: 'Tipo de Carrocería', value: specs.body_type || 'N/D', icon: '🚗' },
        { label: 'Número de Puertas', value: specs.doors ? `${specs.doors} puertas` : 'N/D', icon: '🚪' },
        { label: 'Kilometraje Recorrido', value: (specs.mileage_km !== undefined) ? `${specs.mileage_km.toLocaleString('es-CO')} km` : 'N/D', icon: '🛣️' },
        { label: 'Color Exterior', value: specs.color || 'N/D', icon: '🎨' },
        { label: 'Tracción', value: specs.traction || 'N/D', icon: '🛞' }
    ];

    grid.innerHTML = specItems.map(item => `
        <div class="spec-card">
            <span class="spec-icon" aria-hidden="true">${item.icon}</span>
            <div class="spec-data">
                <span class="spec-label">${escapeHTML(item.label)}</span>
                <strong class="spec-value">${escapeHTML(item.value)}</strong>
            </div>
        </div>
    `).join('');
}

// 4. Slider Deslizable de Fotos
function setupSlider(images) {
    currentSlideIndex = 0;
    const prevBtn = document.getElementById('slider-btn-prev');
    const nextBtn = document.getElementById('slider-btn-next');
    const strip = document.getElementById('thumbnails-strip');

    if (images.length <= 1) {
        prevBtn.style.display = 'none';
        nextBtn.style.display = 'none';
        strip.style.display = 'none';
    } else {
        prevBtn.style.display = 'flex';
        nextBtn.style.display = 'flex';
        strip.style.display = 'flex';

        // Generar miniaturas
        strip.innerHTML = images.map((src, idx) => `
            <button type="button" class="thumbnail-btn ${idx === 0 ? 'active' : ''}" onclick="goToSlide(${idx})" aria-label="Ver foto ${idx + 1}">
                <img src="${src}" alt="Miniatura ${idx + 1}" onerror="this.src='/static/placeholder-car.svg'">
            </button>
        `).join('');
    }

    updateSlideView();
}

function updateSlideView() {
    if (vehicleImages.length === 0) return;
    
    const mainImg = document.getElementById('slider-main-img');
    const counter = document.getElementById('slider-counter');
    const src = vehicleImages[currentSlideIndex];

    mainImg.src = src;
    mainImg.onerror = () => { mainImg.src = '/static/placeholder-car.svg'; };
    counter.textContent = `Foto ${currentSlideIndex + 1} de ${vehicleImages.length}`;

    // Actualizar miniaturas activas
    const thumbs = document.querySelectorAll('.thumbnail-btn');
    thumbs.forEach((thumb, idx) => {
        thumb.classList.toggle('active', idx === currentSlideIndex);
    });
}

function nextSlide() {
    if (vehicleImages.length <= 1) return;
    currentSlideIndex = (currentSlideIndex + 1) % vehicleImages.length;
    updateSlideView();
}

function prevSlide() {
    if (vehicleImages.length <= 1) return;
    currentSlideIndex = (currentSlideIndex - 1 + vehicleImages.length) % vehicleImages.length;
    updateSlideView();
}

function goToSlide(index) {
    if (index >= 0 && index < vehicleImages.length) {
        currentSlideIndex = index;
        updateSlideView();
    }
}

// 5. Soporte para Teclado y Gestos Táctiles
function setupKeyboardNavigation() {
    window.addEventListener('keydown', (e) => {
        if (e.key === 'ArrowLeft') {
            prevSlide();
        } else if (e.key === 'ArrowRight') {
            nextSlide();
        }
    });
}

function setupTouchSwipe() {
    const viewport = document.getElementById('slider-viewport');
    if (!viewport) return;

    let touchStartX = 0;
    let touchEndX = 0;

    viewport.addEventListener('touchstart', (e) => {
        touchStartX = e.changedTouches[0].screenX;
    }, { passive: true });

    viewport.addEventListener('touchend', (e) => {
        touchEndX = e.changedTouches[0].screenX;
        handleSwipe();
    }, { passive: true });

    function handleSwipe() {
        const threshold = 40; // Mínimo de pixeles para swipe
        if (touchEndX < touchStartX - threshold) {
            nextSlide(); // Deslizar hacia la izquierda -> siguiente
        }
        if (touchEndX > touchStartX + threshold) {
            prevSlide(); // Deslizar hacia la derecha -> anterior
        }
    }
}

// Utilidades
function formatCOP(amount) {
    return new Intl.NumberFormat('es-CO', {
        style: 'currency',
        currency: 'COP',
        maximumFractionDigits: 0
    }).format(amount);
}

function escapeHTML(str) {
    return String(str || '')
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;');
}

