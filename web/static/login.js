const form = document.querySelector('#login-form');
const message = document.querySelector('#message');

form.addEventListener('submit', async (event) => {
    event.preventDefault();
    message.textContent = 'Comprobando acceso…';
    message.className = 'message';

    const data = Object.fromEntries(new FormData(form));
    try {
        const response = await fetch('/api/sessions', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data),
        });
        const result = await response.json();
        if (!response.ok) throw new Error(result.error || 'No se pudo iniciar sesión.');

        localStorage.setItem('automarket.session', JSON.stringify({
            token: result.token,
            user: result.user,
        }));
        message.textContent = `Sesión iniciada como ${result.user.name} (${result.user.role}).`;
        message.classList.add('success');
        window.setTimeout(() => window.location.assign('/'), 700);
    } catch (error) {
        message.textContent = error.message || 'No se pudo conectar con el servidor.';
        message.classList.add('error');
    }
});
