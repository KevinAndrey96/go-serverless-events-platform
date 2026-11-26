import { Html5QrcodeScanner } from 'html5-qrcode';

const form = document.getElementById('ticket-form');
const input = document.getElementById('ticket');
const result = document.getElementById('result');
const csrf = document.querySelector('meta[name="csrf-token"]').content;
const PAUSE_MS = 2500;

let busy = false;

function show(admitted, message) {
    result.textContent = message;
    result.className = `rounded-lg px-5 py-6 text-center text-lg font-semibold ${
        admitted ? 'bg-emerald-100 text-emerald-900' : 'bg-rose-100 text-rose-900'
    }`;
}

async function validate(ticket) {
    if (busy || ticket.trim() === '') {
        return;
    }
    busy = true;
    try {
        const response = await fetch(form.dataset.url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json', Accept: 'application/json', 'X-CSRF-TOKEN': csrf },
            body: JSON.stringify({ ticket: ticket.trim() }),
        });
        const body = await response.json().catch(() => ({}));
        show(body.admitted === true, body.message ?? 'No fue posible validar la entrada.');
    } catch {
        show(false, 'Sin conexión. Intenta de nuevo.');
    } finally {
        setTimeout(() => {
            busy = false;
        }, PAUSE_MS);
    }
}

form.addEventListener('submit', (event) => {
    event.preventDefault();
    validate(input.value);
    input.value = '';
});

new Html5QrcodeScanner('scanner', { fps: 10, qrbox: 250, rememberLastUsedCamera: true }, false).render(
    (text) => validate(text),
    () => {},
);
