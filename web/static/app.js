'use strict';

const STORAGE_KEY = 'automarket.session';

const state = {
    token: null,
    user: null,
    catalog: [],
    mine: [],
    admin: { pending: [], vehicles: [], users: [] },
    adminTab: 'pending',
    authIntent: null,
    selectedFiles: [],
    previewURLs: [],
    coverIndex: 0,
    toastTimer: null,
};

const $ = (selector, root = document) => root.querySelector(selector);
const $$ = (selector, root = document) => Array.from(root.querySelectorAll(selector));

const catalogView = $('#catalog-view');
const sellerView = $('#seller-view');
const adminView = $('#admin-view');
const catalogGrid = $('#catalog-grid');
const catalogEmpty = $('#catalog-empty');
const catalogCount = $('#catalog-count');
const catalogSearch = $('#catalog-search');
const navAuth = $('#nav-auth');
const navPublish = $('#nav-publish');
const navAccount = $('#nav-account');
const sellerName = $('#seller-name');
const myVehicles = $('#my-vehicles');
const adminStats = $('#admin-stats');
const adminPanel = $('#admin-panel');
const pendingCount = $('#pending-count');
const authModal = $('#auth-modal');
const detailModal = $('#detail-modal');
const authAlert = $('#auth-alert');
const authContext = $('#auth-context');
const detailContent = $('#detail-content');
const toast = $('#toast');
const publishForm = $('#publish-form');
const publishAlert = $('#publish-alert');
const publishSubmit = $('#publish-submit');
const imageInput = $('#v-images');
const imagePreview = $('#image-preview');
const imageError = $('#image-error');

/* Utilidades de interfaz -------------------------------------------------- */

function el(tag, attributes = {}, ...children) {
    const node = document.createElement(tag);
    for (const [key, value] of Object.entries(attributes)) {
        if (value === null || value === undefined || value === false) continue;
        if (key === 'class') {
            node.className = value;
        } else if (key === 'text') {
            node.textContent = value;
        } else if (key === 'dataset') {
            Object.assign(node.dataset, value);
        } else if (key.startsWith('on') && typeof value === 'function') {
            node.addEventListener(key.slice(2).toLowerCase(), value);
        } else if (key === 'value') {
            node.value = value;
        } else if (value === true) {
            node.setAttribute(key, '');
        } else {
            node.setAttribute(key, value);
        }
    }
    for (const child of children.flat(Infinity)) {
        if (child === null || child === undefined || child === false) continue;
        node.append(child instanceof Node ? child : document.createTextNode(String(child)));
    }
    return node;
}

function clear(node) {
    node.replaceChildren();
}

function formatCOP(value) {
    const amount = Number(value);
    if (!Number.isFinite(amount)) return '—';
    return new Intl.NumberFormat('es-CO', {
        style: 'currency', currency: 'COP', maximumFractionDigits: 0,
    }).format(amount);
}

function formatDate(value) {
    if (!value) return '';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return '';
    return new Intl.DateTimeFormat('es-CO', {
        day: '2-digit', month: 'short', year: 'numeric',
    }).format(date);
}

function formatBytes(bytes) {
    const value = Number(bytes);
    if (!Number.isFinite(value)) return '';
    if (value < 1024 * 1024) return `${Math.round(value / 1024)} KB`;
    return `${(value / (1024 * 1024)).toFixed(1)} MB`;
}

function statusLabel(status) {
    return {
        pending: 'Pendiente',
        published: 'Publicada',
        sold: 'Vendida',
    }[status] || status || 'Sin estado';
}

function statusClass(status) {
    return `badge badge-${status || 'pending'}`;
}

function showToast(message, type = 'success') {
    if (!toast) return;
    toast.textContent = message;
    toast.className = `toast${type === 'error' ? ' error' : ''}`;
    toast.hidden = false;
    window.clearTimeout(state.toastTimer);
    state.toastTimer = window.setTimeout(() => {
        toast.hidden = true;
    }, 4600);
}

function setButtonBusy(button, busy, busyLabel = 'Procesando…') {
    if (!button) return;
    if (busy) {
        if (!button.dataset.originalLabel) button.dataset.originalLabel = button.textContent;
        button.textContent = busyLabel;
        button.disabled = true;
    } else {
        button.textContent = button.dataset.originalLabel || button.textContent;
        button.disabled = false;
    }
}

function updateBodyLock() {
    const anyOpen = [authModal, detailModal].some((modal) => modal && !modal.hidden);
    document.body.style.overflow = anyOpen ? 'hidden' : '';
}

function openModal(id) {
    const modal = document.getElementById(id);
    if (!modal) return;
    modal.hidden = false;
    updateBodyLock();
}

function closeModal(id) {
    const modal = document.getElementById(id);
    if (!modal) return;
    modal.hidden = true;
    updateBodyLock();
}

/* Sesión ---------------------------------------------------------------- */

function loadStoredSession() {
    try {
        const raw = localStorage.getItem(STORAGE_KEY);
        if (!raw) return null;
        const parsed = JSON.parse(raw);
        if (!parsed || typeof parsed.token !== 'string' || !parsed.user) return null;
        return parsed;
    } catch {
        return null;
    }
}

function persistSession() {
    if (state.token && state.user) {
        localStorage.setItem(STORAGE_KEY, JSON.stringify({ token: state.token, user: state.user }));
    } else {
        localStorage.removeItem(STORAGE_KEY);
    }
}

function clearSession() {
    state.token = null;
    state.user = null;
    localStorage.removeItem(STORAGE_KEY);
}

function handleExpiredSession() {
    clearSession();
    renderNav();
    if (!sellerView.hidden) showView('catalog');
    if (!adminView.hidden) showView('catalog');
}

async function api(path, options = {}) {
    const { method = 'GET', body, auth = true, silent = false, headers = {} } = options;
    const requestHeaders = new Headers(headers);
    if (auth && state.token) requestHeaders.set('Authorization', `Bearer ${state.token}`);
    if (body !== undefined && !(body instanceof FormData)) requestHeaders.set('Content-Type', 'application/json');

    let response;
    try {
        response = await fetch(path, { method, headers: requestHeaders, body });
    } catch {
        throw new Error('No se pudo conectar con el servidor.');
    }

    let data = null;
    if (response.status !== 204) {
        const text = await response.text();
        if (text) {
            try {
                data = JSON.parse(text);
            } catch {
                data = null;
            }
        }
    }
    if (!response.ok) {
        if (response.status === 401 && auth && state.token) {
            handleExpiredSession();
            if (!silent) showToast('Tu sesión expiró. Ingresa de nuevo.', 'error');
        }
        const error = new Error((data && data.error) || `No se pudo completar la operación (${response.status}).`);
        error.status = response.status;
        throw error;
    }
    return data;
}

async function restoreSession() {
    const stored = loadStoredSession();
    if (!stored) return;
    state.token = stored.token;
    state.user = stored.user;
    try {
        const data = await api('/api/session', { silent: true });
        if (data && data.user) {
            state.user = data.user;
            persistSession();
        }
    } catch {
        clearSession();
    }
}

function renderNav() {
    const user = state.user;
    navAuth.hidden = Boolean(user);
    navPublish.hidden = Boolean(user && user.role !== 'seller');
    navAccount.hidden = !user;
    if (!user) {
        clear(navAccount);
        return;
    }
    const name = el('span', { class: 'nav-user-name', text: user.name || user.email || 'Mi cuenta' });
    const role = el('span', { class: 'nav-role', text: user.role === 'admin' ? 'Admin' : 'Vendedor' });
    const panel = el('button', {
        type: 'button', class: 'button button-sm button-ghost', text: 'Mi panel',
        onclick: openPanel,
    });
    const logoutButton = el('button', {
        type: 'button', class: 'button button-sm button-outline', text: 'Salir',
        onclick: handleLogout,
    });
    navAccount.replaceChildren(name, role, panel, logoutButton);
}

async function completeLogin(user, token) {
    state.token = token;
    state.user = user;
    persistSession();
    renderNav();
    const intent = state.authIntent;
    state.authIntent = null;
    closeModal('auth-modal');
    showToast(`Bienvenido, ${user.name || user.email}.`);
    if (user.role === 'admin') {
        await showAdmin();
    } else {
        await showSeller();
        if (intent === 'publish') {
            window.setTimeout(() => $('#v-make')?.focus(), 180);
        }
    }
}

async function handleLogout() {
    try {
        await api('/api/session', { method: 'DELETE', silent: true });
    } catch {
        // La sesión local se cierra igualmente.
    }
    clearSession();
    renderNav();
    resetImages();
    showView('catalog');
    showToast('Sesión cerrada.');
}

function openAuth(intent = 'account') {
    state.authIntent = intent;
    authAlert.hidden = true;
    authContext.textContent = intent === 'publish'
        ? 'Para publicar un vehículo necesitas una cuenta de vendedor. Consultar el catálogo sigue siendo libre.'
        : 'Solo necesitas una cuenta para publicar vehículos. Explorar el catálogo y ver el contacto es libre.';
    setAuthTab('login');
    openModal('auth-modal');
    window.setTimeout(() => $('#login-email')?.focus(), 80);
}

function setAuthTab(tab) {
    $$('[data-auth-tab]').forEach((button) => {
        const active = button.dataset.authTab === tab;
        button.classList.toggle('active', active);
        button.setAttribute('aria-selected', String(active));
    });
    $('#login-form').hidden = tab !== 'login';
    $('#register-form').hidden = tab !== 'register';
    authAlert.hidden = true;
    window.setTimeout(() => {
        if (tab === 'login') $('#login-email')?.focus();
        else $('#reg-name')?.focus();
    }, 30);
}

function showAuthAlert(message, type = 'error') {
    authAlert.textContent = message;
    authAlert.className = `alert ${type}`;
    authAlert.hidden = false;
}

/* Catálogo público ------------------------------------------------------ */

function getCover(vehicle) {
    const images = Array.isArray(vehicle.images) ? vehicle.images : [];
    return images.find((image) => image.cover) || images[0] || null;
}

function showView(name) {
    const views = { catalog: catalogView, seller: sellerView, admin: adminView };
    for (const [key, view] of Object.entries(views)) {
        if (view) view.hidden = key !== name;
    }
    $$('[data-nav]').forEach((button) => button.classList.toggle('active', button.dataset.nav === name));
}

function renderCatalogSkeleton() {
    const cards = Array.from({ length: 6 }, () => el('div', { class: 'skeleton-card' },
        el('div', { class: 'skeleton-media' }),
        el('div', { class: 'skeleton-line' }),
        el('div', { class: 'skeleton-line short' }),
    ));
    catalogGrid.replaceChildren(...cards);
    catalogEmpty.hidden = true;
}

function showCatalogMessage(title, message) {
    const heading = catalogEmpty.querySelector('h3');
    const paragraph = catalogEmpty.querySelector('p');
    if (heading) heading.textContent = title;
    if (paragraph) paragraph.textContent = message;
    catalogEmpty.hidden = false;
}

function renderCatalog() {
    const term = (catalogSearch.value || '').trim().toLocaleLowerCase('es');
    const vehicles = state.catalog.filter((vehicle) => {
        if (!term) return true;
        return `${vehicle.make} ${vehicle.model} ${vehicle.year}`.toLocaleLowerCase('es').includes(term);
    });
    catalogCount.textContent = `${vehicles.length} ${vehicles.length === 1 ? 'vehículo' : 'vehículos'}`;
    if (vehicles.length === 0) {
        catalogGrid.replaceChildren();
        showCatalogMessage(
            'No encontramos vehículos con ese criterio',
            'Prueba con otra búsqueda o vuelve más tarde: el catálogo se actualiza cuando el administrador aprueba publicaciones.',
        );
        return;
    }
    catalogEmpty.hidden = true;
    catalogGrid.replaceChildren(...vehicles.map(vehicleCard));
}

function vehicleCard(vehicle) {
    const cover = getCover(vehicle);
    const media = el('div', { class: 'vehicle-media' },
        cover
            ? el('img', { src: cover.url, alt: `${vehicle.make} ${vehicle.model}`, loading: 'lazy' })
            : el('div', { class: 'no-photo' }, el('span', { text: 'Sin foto' })),
        el('span', { class: 'vehicle-year', text: String(vehicle.year || '') }),
    );
    const body = el('div', { class: 'vehicle-body' },
        el('h3', { class: 'vehicle-title', text: `${vehicle.make} ${vehicle.model}` }),
        el('p', { class: 'vehicle-price', text: formatCOP(vehicle.price_cop) }),
        el('p', { class: 'vehicle-desc', text: vehicle.description || 'Sin descripción.' }),
        el('div', { class: 'vehicle-foot' },
            el('span', {}, 'Vendedor: ', el('strong', { text: vehicle.seller?.name || 'No disponible' })),
            el('span', { class: 'link-arrow', text: 'Ver contacto →' }),
        ),
    );
    const card = el('article', {
        class: 'vehicle-card', tabindex: '0', role: 'button',
        'aria-label': `Ver ${vehicle.make} ${vehicle.model}`,
        onclick: () => openDetail(vehicle),
    }, media, body);
    card.addEventListener('keydown', (event) => {
        if (event.key === 'Enter' || event.key === ' ') {
            event.preventDefault();
            openDetail(vehicle);
        }
    });
    return card;
}

async function loadCatalog() {
    renderCatalogSkeleton();
    try {
        const data = await api('/api/vehicles', { auth: false });
        state.catalog = Array.isArray(data?.vehicles) ? data.vehicles : [];
        renderCatalog();
    } catch (error) {
        state.catalog = [];
        catalogGrid.replaceChildren();
        catalogCount.textContent = '0 vehículos';
        showCatalogMessage('No se pudo cargar el catálogo', error.message || 'Intenta recargar la página.');
    }
}

/* Detalle y contacto ---------------------------------------------------- */

function openDetail(vehicle) {
    detailContent.replaceChildren(buildDetail(vehicle));
    openModal('detail-modal');
}

function buildDetail(vehicle) {
    const images = Array.isArray(vehicle.images) ? [...vehicle.images] : [];
    images.sort((a, b) => Number(Boolean(b.cover)) - Number(Boolean(a.cover)));
    let activeIndex = 0;

    const mainImage = el('div', { class: 'detail-main-image' });
    const thumbs = el('div', { class: 'detail-thumbs' });

    function paintImage(index) {
        activeIndex = index;
        clear(mainImage);
        if (images.length === 0) {
            mainImage.append(el('div', { class: 'no-photo', text: 'Sin fotografías' }));
        } else {
            const image = images[index] || images[0];
            mainImage.append(el('img', {
                src: image.url,
                alt: `${vehicle.make} ${vehicle.model} - imagen ${index + 1}`,
            }));
        }
        $$('.detail-thumb', thumbs).forEach((button, buttonIndex) => {
            button.classList.toggle('active', buttonIndex === activeIndex);
        });
    }

    paintImage(0);
    if (images.length > 1) {
        images.forEach((image, index) => {
            thumbs.append(el('button', {
                type: 'button', class: 'detail-thumb', 'aria-label': `Ver imagen ${index + 1}`,
                onclick: () => paintImage(index),
            }, el('img', { src: image.url, alt: '' })));
        });
        $$('.detail-thumb', thumbs)[0]?.classList.add('active');
    }

    const gallery = el('div', { class: 'detail-gallery' }, mainImage, images.length > 1 ? thumbs : null);

    const titleRow = el('div', { class: 'my-vehicle-top' },
        el('h2', { id: 'detail-title', text: `${vehicle.make} ${vehicle.model}` }),
        vehicle.status && vehicle.status !== 'published'
            ? el('span', { class: statusClass(vehicle.status), text: statusLabel(vehicle.status) })
            : null,
    );

    const meta = el('p', {
        class: 'moderation-meta',
        text: `${vehicle.year || '—'} · Publicado el ${formatDate(vehicle.created_at) || '—'}`,
    });

    const info = el('div', { class: 'detail-info' },
        titleRow,
        meta,
        el('p', { class: 'detail-price', text: formatCOP(vehicle.price_cop) }),
        el('p', { class: 'detail-description', text: vehicle.description || 'Sin descripción.' }),
        el('div', { class: 'detail-seller-contact' },
            el('p', { class: 'detail-seller-name', text: `Vendedor: ${vehicle.seller?.name || 'No disponible'}` }),
            el('p', { class: 'detail-seller-phone', text: `Teléfono: ${vehicle.seller?.phone || 'No disponible'}` }),
            el('p', { class: 'detail-seller-email', text: `Correo: ${vehicle.seller?.email || 'No disponible'}` }),
        ),
    );

    return el('div', { class: 'detail-layout' }, gallery, info);
}

/* Panel del vendedor ---------------------------------------------------- */

async function openPanel() {
    if (!state.user) {
        openAuth('account');
        return;
    }
    if (state.user.role === 'admin') {
        await showAdmin();
    } else {
        await showSeller();
    }
}

function requestPublish() {
    if (!state.user) {
        openAuth('publish');
        return;
    }
    if (state.user.role !== 'seller') {
        showToast('Solo las cuentas de vendedor pueden publicar vehículos.', 'error');
        if (state.user.role === 'admin') showAdmin();
        return;
    }
    showSeller();
    window.setTimeout(() => $('#v-make')?.focus(), 180);
}

async function showSeller() {
    if (!state.user) {
        openAuth('publish');
        return;
    }
    if (state.user.role !== 'seller') {
        await showAdmin();
        return;
    }
    showView('seller');
    sellerName.textContent = state.user.name || 'vendedor';
    await loadMine();
}

async function showAdmin() {
    if (!state.user) {
        openAuth('account');
        return;
    }
    if (state.user.role !== 'admin') {
        showToast('No tienes permiso para acceder al panel de administración.', 'error');
        showView('catalog');
        return;
    }
    showView('admin');
    await loadAdmin();
}

async function loadMine() {
    myVehicles.replaceChildren(el('p', { class: 'muted', text: 'Cargando tus publicaciones…' }));
    try {
        const data = await api('/api/seller/vehicles');
        state.mine = Array.isArray(data?.vehicles) ? data.vehicles : [];
        renderMine();
    } catch (error) {
        myVehicles.replaceChildren(el('p', { class: 'alert error', text: error.message }));
    }
}

function renderMine() {
    if (state.mine.length === 0) {
        myVehicles.replaceChildren(el('div', { class: 'empty-state' },
            el('h3', { text: 'Aún no tienes publicaciones' }),
            el('p', { text: 'Completa el formulario para enviar tu primer vehículo a revisión.' }),
        ));
        return;
    }
    myVehicles.replaceChildren(...state.mine.map((vehicle) => {
        const cover = getCover(vehicle);
        const media = el('div', { class: 'my-vehicle-media' },
            cover
                ? el('img', { src: cover.url, alt: `${vehicle.make} ${vehicle.model}` })
                : el('div', { class: 'no-photo', text: 'Sin foto' }),
        );
        const actions = el('div', { class: 'my-vehicle-actions' });
        if (vehicle.status === 'published') {
            actions.append(
                el('button', { type: 'button', class: 'button button-sm button-outline', text: 'Ver publicación', onclick: () => openDetail(vehicle) }),
                el('button', { type: 'button', class: 'button button-sm button-accent', text: 'Reportar venta', onclick: () => reportSale(vehicle) }),
            );
        }
        if (vehicle.status === 'pending') {
            actions.append(el('span', { class: 'hint', text: 'Esperando autorización del administrador.' }));
        }
        if (vehicle.status === 'sold') {
            actions.append(el('span', { class: 'hint', text: 'Ya no aparece en el catálogo público.' }));
        }
        return el('article', { class: 'my-vehicle' },
            media,
            el('div', { class: 'my-vehicle-body' },
                el('div', { class: 'my-vehicle-top' },
                    el('h3', { text: `${vehicle.make} ${vehicle.model}` }),
                    el('span', { class: statusClass(vehicle.status), text: statusLabel(vehicle.status) }),
                ),
                el('p', { class: 'price', text: formatCOP(vehicle.price_cop) }),
                el('p', { class: 'moderation-meta', text: `${vehicle.year} · ${formatDate(vehicle.created_at)}` }),
                actions,
            ),
        );
    }));
}

async function reportSale(vehicle) {
    const confirmed = window.confirm(
        `¿Confirmas que vendiste ${vehicle.make} ${vehicle.model}? La publicación dejará de aparecer en el catálogo.`,
    );
    if (!confirmed) return;
    try {
        await api(`/api/vehicles/${vehicle.id}/sale`, { method: 'POST' });
        showToast('Venta reportada. La publicación ya no aparece en el catálogo.');
        await Promise.all([loadMine(), loadCatalog()]);
    } catch (error) {
        showToast(error.message, 'error');
    }
}

/* Publicación con imágenes ---------------------------------------------- */

function revokePreviewURLs() {
    for (const url of state.previewURLs) URL.revokeObjectURL(url);
    state.previewURLs = [];
}

function resetImages() {
    revokePreviewURLs();
    state.selectedFiles = [];
    state.coverIndex = 0;
    imageInput.value = '';
    imagePreview.hidden = true;
    imagePreview.replaceChildren();
    imageError.hidden = true;
}

function showImageError(message) {
    imageError.textContent = message;
    imageError.hidden = false;
}

async function handleImageSelection(files) {
    imageError.hidden = true;
    if (!files || files.length === 0) return;
    const valid = [];
    try {
        for (const file of files) {
            const typeOK = ['image/jpeg', 'image/png'].includes(file.type) || /\.(jpe?g|png)$/i.test(file.name);
            if (!typeOK) throw new Error(`"${file.name}" no es JPG ni PNG.`);
            valid.push(file);
        }
    } catch (error) {
        imageInput.value = '';
        showImageError(error.message);
        return;
    }

    revokePreviewURLs();
    state.selectedFiles = valid;
    state.coverIndex = 0;
    renderImagePreview();
}

function renderImagePreview() {
    revokePreviewURLs();
    imagePreview.replaceChildren();
    if (state.selectedFiles.length === 0) {
        imagePreview.hidden = true;
        return;
    }
    imagePreview.hidden = false;
    state.selectedFiles.forEach((file, index) => {
        const url = URL.createObjectURL(file);
        state.previewURLs.push(url);
        const isCover = index === state.coverIndex;
        const radio = el('input', {
            type: 'radio', name: 'cover-image', value: String(index), checked: isCover,
            onchange: () => {
                state.coverIndex = index;
                renderImagePreview();
            },
        });
        const remove = el('button', {
            type: 'button', class: 'remove-image', 'aria-label': `Quitar ${file.name}`, text: '×',
            onclick: () => {
                state.selectedFiles.splice(index, 1);
                if (state.coverIndex >= state.selectedFiles.length) state.coverIndex = 0;
                else if (index < state.coverIndex) state.coverIndex -= 1;
                imageInput.value = '';
                renderImagePreview();
            },
        });
        imagePreview.append(el('div', { class: `preview-item${isCover ? ' is-cover' : ''}` },
            el('img', { src: url, alt: `Vista previa de ${file.name}` }),
            remove,
            el('div', { class: 'preview-meta' },
                el('p', { class: 'preview-name', title: file.name, text: file.name }),
                el('p', { class: 'hint', text: formatBytes(file.size) }),
                el('label', { class: 'cover-choice' }, radio, 'Portada'),
            ),
        ));
    });
}

async function handlePublish(event) {
    event.preventDefault();
    publishAlert.hidden = true;
    imageError.hidden = true;

    const make = $('#v-make').value.trim();
    const model = $('#v-model').value.trim();
    const year = Number.parseInt($('#v-year').value, 10);
    const price = Number.parseInt($('#v-price').value, 10);
    const description = $('#v-desc').value.trim();
    const maxYear = new Date().getFullYear() + 1;

    if (!make || !model || !description) {
        publishAlert.className = 'alert error';
        publishAlert.textContent = 'Completa la marca, el modelo y la descripción.';
        publishAlert.hidden = false;
        return;
    }
    if (!Number.isInteger(year) || year < 1886 || year > maxYear) {
        publishAlert.className = 'alert error';
        publishAlert.textContent = `El año debe estar entre 1886 y ${maxYear}.`;
        publishAlert.hidden = false;
        return;
    }
    if (!Number.isInteger(price) || price <= 0) {
        publishAlert.className = 'alert error';
        publishAlert.textContent = 'Ingresa un precio válido en pesos colombianos.';
        publishAlert.hidden = false;
        return;
    }
    if (state.selectedFiles.length === 0) {
        showImageError('Agrega al menos una imagen del vehículo.');
        return;
    }
    if (state.coverIndex < 0 || state.coverIndex >= state.selectedFiles.length) {
        showImageError('Selecciona una imagen de portada válida.');
        return;
    }

    const formData = new FormData();
    formData.append('make', make);
    formData.append('model', model);
    formData.append('year', String(year));
    formData.append('price_cop', String(price));
    formData.append('description', description);
    formData.append('cover_index', String(state.coverIndex));
    state.selectedFiles.forEach((file) => formData.append('images', file, file.name));

    setButtonBusy(publishSubmit, true, 'Enviando…');
    try {
        const vehicle = await api('/api/vehicles', { method: 'POST', body: formData });
        publishAlert.className = 'alert success';
        publishAlert.textContent = `Publicación enviada. Quedó en estado ${statusLabel(vehicle.status).toLowerCase()} y será revisada por un administrador.`;
        publishAlert.hidden = false;
        publishForm.reset();
        resetImages();
        showToast('Publicación enviada a revisión.');
        await loadMine();
    } catch (error) {
        publishAlert.className = 'alert error';
        publishAlert.textContent = error.message;
        publishAlert.hidden = false;
    } finally {
        setButtonBusy(publishSubmit, false);
    }
}

/* Administración -------------------------------------------------------- */

async function loadAdmin() {
    adminPanel.replaceChildren(el('p', { class: 'muted', text: 'Cargando información de moderación…' }));
    try {
        const [pending, vehicles, users] = await Promise.all([
            api('/api/admin/vehicles/pending'),
            api('/api/admin/vehicles'),
            api('/api/admin/users'),
        ]);
        state.admin.pending = Array.isArray(pending?.vehicles) ? pending.vehicles : [];
        state.admin.vehicles = Array.isArray(vehicles?.vehicles) ? vehicles.vehicles : [];
        state.admin.users = Array.isArray(users?.users) ? users.users : [];
        renderAdmin();
    } catch (error) {
        adminPanel.replaceChildren(el('p', { class: 'alert error', text: error.message }));
    }
}

function renderAdmin() {
    const published = state.admin.vehicles.filter((vehicle) => vehicle.status === 'published').length;
    const stats = [
        { label: 'Pendientes', value: state.admin.pending.length },
        { label: 'Publicadas', value: published },
        { label: 'Usuarios', value: state.admin.users.length },
    ];
    adminStats.replaceChildren(...stats.map((stat) => el('div', { class: 'stat' },
        el('span', { text: stat.label }),
        el('strong', { text: String(stat.value) }),
    )));
    pendingCount.textContent = String(state.admin.pending.length);

    $$('[data-admin-tab]').forEach((button) => {
        const active = button.dataset.adminTab === state.adminTab;
        button.classList.toggle('active', active);
        button.setAttribute('aria-selected', String(active));
    });

    adminPanel.replaceChildren();
    if (state.adminTab === 'pending') {
        if (state.admin.pending.length === 0) {
            adminPanel.append(emptyAdmin('No hay publicaciones pendientes', 'Cuando un vendedor envíe un vehículo, aparecerá aquí para su revisión.'));
            return;
        }
        adminPanel.append(...state.admin.pending.map((vehicle) => moderationCard(vehicle, true)));
        return;
    }
    if (state.adminTab === 'vehicles') {
        if (state.admin.vehicles.length === 0) {
            adminPanel.append(emptyAdmin('Todavía no hay publicaciones', 'Las publicaciones de los vendedores aparecerán aquí.'));
            return;
        }
        adminPanel.append(...state.admin.vehicles.map((vehicle) => moderationCard(vehicle, vehicle.status === 'pending')));
        return;
    }
    if (state.admin.users.length === 0) {
        adminPanel.append(emptyAdmin('No hay usuarios registrados', 'Las cuentas de vendedores aparecerán aquí.'));
        return;
    }
    adminPanel.append(...state.admin.users.map(userRow));
}

function emptyAdmin(title, message) {
    return el('div', { class: 'empty-state' },
        el('h3', { text: title }),
        el('p', { text: message }),
    );
}

function moderationCard(vehicle, allowModeration) {
    const cover = getCover(vehicle);
    const media = el('div', { class: 'moderation-media' },
        cover
            ? el('img', { src: cover.url, alt: `${vehicle.make} ${vehicle.model}` })
            : el('div', { class: 'no-photo', text: 'Sin foto' }),
    );
    const actions = el('div', { class: 'moderation-actions' });
    if (allowModeration) {
        actions.append(
            el('button', { type: 'button', class: 'button button-sm button-accent', text: 'Aprobar', onclick: () => approveVehicle(vehicle) }),
        );
    }
    if (vehicle.status === 'published') {
        actions.append(el('button', { type: 'button', class: 'button button-sm button-outline', text: 'Ver', onclick: () => openDetail(vehicle) }));
    }
    actions.append(el('button', { type: 'button', class: 'button button-sm button-outline', text: 'Eliminar', onclick: () => deleteVehicle(vehicle) }));

    return el('article', { class: 'moderation-card' },
        media,
        el('div', { class: 'moderation-body' },
            el('div', { class: 'my-vehicle-top' },
                el('h3', { text: `${vehicle.make} ${vehicle.model}` }),
                el('span', { class: statusClass(vehicle.status), text: statusLabel(vehicle.status) }),
            ),
            el('p', { class: 'moderation-meta', text: `${vehicle.year} · ${formatCOP(vehicle.price_cop)} · ${formatDate(vehicle.created_at)}` }),
            el('p', { class: 'moderation-meta', text: vehicle.description || 'Sin descripción.' }),
            el('p', { class: 'moderation-contact', text: `Vendedor: ${vehicle.seller?.name || '—'} · ${vehicle.seller?.phone || 'Sin teléfono'} · ${vehicle.seller?.email || 'Sin correo'}` }),
        ),
        actions,
    );
}

function userRow(user) {
    const isSelf = state.user && state.user.id === user.id;
    return el('div', { class: 'user-row' },
        el('div', {},
            el('h3', { text: `${user.name || 'Sin nombre'}${isSelf ? ' (tú)' : ''}` }),
            el('p', { text: `${user.email} · ${user.phone || 'Sin teléfono'}` }),
        ),
        el('div', { class: 'user-row-actions' },
            el('span', { class: user.role === 'admin' ? 'badge badge-published' : 'badge badge-pending', text: user.role === 'admin' ? 'Admin' : 'Vendedor' }),
            el('button', {
                type: 'button', class: 'button button-sm button-danger', text: 'Eliminar',
                disabled: isSelf, title: isSelf ? 'No puedes eliminar tu propia cuenta' : '',
                onclick: () => deleteUser(user),
            }),
        ),
    );
}

async function refreshAdminAndCatalog() {
    await Promise.all([loadAdmin(), loadCatalog()]);
}

async function approveVehicle(vehicle) {
    try {
        await api(`/api/admin/vehicles/${vehicle.id}/approve`, { method: 'POST' });
        showToast('Publicación aprobada y visible en el catálogo.');
        await refreshAdminAndCatalog();
    } catch (error) {
        showToast(error.message, 'error');
    }
}

async function deleteVehicle(vehicle) {
    const confirmed = window.confirm(`¿Eliminar definitivamente la publicación ${vehicle.make} ${vehicle.model}?`);
    if (!confirmed) return;
    try {
        await api(`/api/admin/vehicles/${vehicle.id}`, { method: 'DELETE' });
        showToast('Publicación eliminada.');
        await refreshAdminAndCatalog();
    } catch (error) {
        showToast(error.message, 'error');
    }
}

async function deleteUser(user) {
    const confirmed = window.confirm(`¿Eliminar la cuenta de ${user.name || user.email}? También se eliminarán sus publicaciones.`);
    if (!confirmed) return;
    try {
        await api(`/api/admin/users/${user.id}`, { method: 'DELETE' });
        showToast('Usuario eliminado.');
        await refreshAdminAndCatalog();
    } catch (error) {
        showToast(error.message, 'error');
    }
}

/* Eventos e inicio ------------------------------------------------------ */

function bindEvents() {
    $('#brand-link').addEventListener('click', (event) => {
        event.preventDefault();
        showView('catalog');
        window.scrollTo({ top: 0, behavior: 'smooth' });
    });
    $('#nav-catalog').addEventListener('click', () => showView('catalog'));
    navPublish.addEventListener('click', requestPublish);
    $('#hero-publish').addEventListener('click', requestPublish);
    navAuth.addEventListener('click', () => openAuth('account'));

    $('#seller-logout').addEventListener('click', handleLogout);
    $('#admin-logout').addEventListener('click', handleLogout);
    $('#refresh-mine').addEventListener('click', loadMine);

    catalogSearch.addEventListener('input', renderCatalog);

    $('#pick-images').addEventListener('click', () => {
        imageInput.value = '';
        imageInput.click();
    });
    imageInput.addEventListener('change', () => handleImageSelection(Array.from(imageInput.files || [])));
    publishForm.addEventListener('submit', handlePublish);

    $('#login-form').addEventListener('submit', async (event) => {
        event.preventDefault();
        authAlert.hidden = true;
        const button = event.target.querySelector('button[type="submit"]');
        const email = $('#login-email').value.trim();
        const password = $('#login-password').value;
        setButtonBusy(button, true, 'Entrando…');
        try {
            const data = await api('/api/sessions', {
                method: 'POST', auth: false,
                body: JSON.stringify({ email, password }),
            });
            await completeLogin(data.user, data.token);
            event.target.reset();
        } catch (error) {
            showAuthAlert(error.message);
        } finally {
            setButtonBusy(button, false);
        }
    });

    $('#register-form').addEventListener('submit', async (event) => {
        event.preventDefault();
        authAlert.hidden = true;
        const button = event.target.querySelector('button[type="submit"]');
        const name = $('#reg-name').value.trim();
        const phone = $('#reg-phone').value.trim();
        const email = $('#reg-email').value.trim();
        const password = $('#reg-password').value;
        setButtonBusy(button, true, 'Creando cuenta…');
        try {
            await api('/api/sellers', {
                method: 'POST', auth: false,
                body: JSON.stringify({ name, phone, email, password }),
            });
            const data = await api('/api/sessions', {
                method: 'POST', auth: false,
                body: JSON.stringify({ email, password }),
            });
            await completeLogin(data.user, data.token);
            event.target.reset();
        } catch (error) {
            showAuthAlert(error.message);
        } finally {
            setButtonBusy(button, false);
        }
    });

    $$('[data-auth-tab]').forEach((button) => {
        button.addEventListener('click', () => setAuthTab(button.dataset.authTab));
    });
    $$('[data-admin-tab]').forEach((button) => {
        button.addEventListener('click', () => {
            state.adminTab = button.dataset.adminTab;
            renderAdmin();
        });
    });

    document.addEventListener('click', (event) => {
        const closer = event.target.closest('[data-close]');
        if (closer) closeModal(closer.dataset.close);
    });
    document.addEventListener('keydown', (event) => {
        if (event.key !== 'Escape') return;
        if (!authModal.hidden) closeModal('auth-modal');
        else if (!detailModal.hidden) closeModal('detail-modal');
    });
}

async function init() {
    bindEvents();
    await restoreSession();
    renderNav();
    await loadCatalog();
    const params = new URLSearchParams(window.location.search);
    if (params.get('acceder') === '1') openAuth('account');
}

init();
