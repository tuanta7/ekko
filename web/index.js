// DOM Elements
const statusEl = document.getElementById('status');
const statusText = statusEl.querySelector('.status-text');
const transcriptionList = document.getElementById('transcription-list');
const clearBtn = document.getElementById('clear-btn');
const copyBtn = document.getElementById('copy-btn');

// Transcription data store
let transcriptions = [];

// Format timestamp
function formatTime(date) {
    return date.toLocaleTimeString('en-US', {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: false
    });
}

// Create transcription item element
function createTranscriptionItem(data) {
    const item = document.createElement('div');
    item.className = 'transcription-item';

    item.innerHTML = `
        <div class="transcription-header">
            <span class="transcription-seq">#${data.seq}</span>
            <span class="transcription-time">${data.time}</span>
        </div>
        <div class="transcription-text">${escapeHtml(data.text)}</div>
    `;

    return item;
}

// Escape HTML to prevent XSS
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Show empty state
function showEmptyState() {
    transcriptionList.innerHTML = `
        <div class="empty-state">Waiting for transcriptions...</div>
    `;
}

// Add transcription to the list
function addTranscription(text, seq) {
    // Remove empty state if present
    const emptyState = transcriptionList.querySelector('.empty-state');
    if (emptyState) {
        emptyState.remove();
    }

    const data = {
        seq: seq || transcriptions.length + 1,
        text: text,
        time: formatTime(new Date())
    };

    transcriptions.push(data);

    const item = createTranscriptionItem(data);
    transcriptionList.appendChild(item);

    // Auto-scroll to bottom
    transcriptionList.scrollTop = transcriptionList.scrollHeight;
}

// Update connection status
function setStatus(state, message) {
    statusEl.className = `status ${state}`;
    statusText.textContent = message;
}

// Initialize SSE connection
function initSSE() {
    if (typeof EventSource === 'undefined') {
        setStatus('error', 'SSE not supported');
        return;
    }

    const source = new EventSource('/sse');

    source.onopen = function() {
        setStatus('connected', 'Connected');
    };

    source.onmessage = function(event) {
        try {
            // Try parsing as JSON first (for structured data with seq)
            const data = JSON.parse(event.data);
            addTranscription(data.text, data.seq);
        } catch {
            // Fallback to plain text
            addTranscription(event.data);
        }
    };

    source.onerror = function() {
        if (source.readyState === EventSource.CLOSED) {
            setStatus('error', 'Disconnected');
        } else {
            setStatus('error', 'Connection error');
        }

        // Attempt reconnection after 3 seconds
        setTimeout(() => {
            setStatus('', 'Reconnecting...');
            source.close();
            initSSE();
        }, 3000);
    };
}

// Clear all transcriptions
function clearTranscriptions() {
    transcriptions = [];
    showEmptyState();
}

// Copy all transcriptions to clipboard
function copyAllTranscriptions() {
    const text = transcriptions
        .map(t => `[${t.time}] ${t.text}`)
        .join('\n');

    navigator.clipboard.writeText(text).then(() => {
        const originalText = copyBtn.textContent;
        copyBtn.textContent = 'Copied!';
        setTimeout(() => {
            copyBtn.textContent = originalText;
        }, 2000);
    }).catch(err => {
        console.error('Failed to copy:', err);
    });
}

// Event listeners
clearBtn.addEventListener('click', clearTranscriptions);
copyBtn.addEventListener('click', copyAllTranscriptions);

// Initialize
showEmptyState();
initSSE();
