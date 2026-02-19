const statusEl = document.getElementById('status');
const statusText = statusEl.querySelector('.status-text');
const transcriptionList = document.getElementById('transcription-list');
const processingEl = document.getElementById('processing');
const clearBtn = document.getElementById('clear-btn');
const copyBtn = document.getElementById('copy-btn');

// Settings elements
const settingsEl = document.getElementById('settings');
const sourceSelect = document.getElementById('source-select');
const durationValue = document.getElementById('duration-value');
const durationDec = document.getElementById('duration-dec');
const durationInc = document.getElementById('duration-inc');
const startBtn = document.getElementById('start-btn');
const stopBtn = document.getElementById('stop-btn');

let transcriptions = [];
let chunkDuration = 5;
let isRecording = false;
let sseSource = null;
let seqCounter = 0;

function formatTime(date) {
    return date.toLocaleTimeString('en-US', {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: false
    });
}

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

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function showEmptyState() {
    transcriptionList.innerHTML = `
        <div class="empty-state">Waiting for transcriptions...</div>
    `;
}

function addTranscription(text, seq) {
    const emptyState = transcriptionList.querySelector('.empty-state');
    if (emptyState) {
        emptyState.remove();
    }

    const data = {
        seq: seq || ++seqCounter,
        text: text,
        time: formatTime(new Date())
    };

    transcriptions.push(data);

    const item = createTranscriptionItem(data);
    transcriptionList.appendChild(item);
    transcriptionList.scrollTop = transcriptionList.scrollHeight;
}

function setStatus(state, message) {
    statusEl.className = `status ${state}`;
    statusText.textContent = message;
}

function showProcessing() {
    processingEl.classList.add('active');
    transcriptionList.scrollTop = transcriptionList.scrollHeight;
}

function hideProcessing() {
    processingEl.classList.remove('active');
}

function connectSSE() {
    if (typeof EventSource === 'undefined') {
        setStatus('error', 'SSE not supported');
        return;
    }

    sseSource = new EventSource('/sse');

    sseSource.onopen = function() {
        setStatus('connected', 'Recording');
    };

    sseSource.onmessage = function(event) {
        try {
            const data = JSON.parse(event.data);

            // Handle message types
            if (data.type === 'connected') {
                console.log('SSE stream connected');
                return;
            }

            if (data.type === 'ended') {
                console.log('Session ended');
                disconnectSSE();
                onSessionEnded();
                return;
            }

            // Handle transcription
            if (data.text) {
                hideProcessing();
                addTranscription(data.text, data.seq);
                showProcessing(); // Show again for next chunk
            }
        } catch {
            if (event.data) {
                addTranscription(event.data);
            }
        }
    };

    sseSource.onerror = function() {
        disconnectSSE();
        setStatus('error', 'Connection lost');
        onSessionEnded();
    };
}

function disconnectSSE() {
    if (sseSource) {
        sseSource.close();
        sseSource = null;
    }
}

function onSessionEnded() {
    isRecording = false;
    settingsEl.classList.remove('recording');
    startBtn.disabled = false;
    stopBtn.disabled = true;
    hideProcessing();
    setStatus('', 'Ready');
}

function clearTranscriptions() {
    transcriptions = [];
    seqCounter = 0;
    showEmptyState();
}

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

clearBtn.addEventListener('click', clearTranscriptions);
copyBtn.addEventListener('click', copyAllTranscriptions);

// Duration controls
durationDec.addEventListener('click', () => {
    if (chunkDuration > 1) {
        chunkDuration--;
        durationValue.textContent = chunkDuration + 's';
    }
});

durationInc.addEventListener('click', () => {
    if (chunkDuration < 60) {
        chunkDuration++;
        durationValue.textContent = chunkDuration + 's';
    }
});

// Fetch audio sources
async function fetchSources() {
    try {
        const res = await fetch('/sources');
        const sources = await res.json();

        sourceSelect.innerHTML = '';
        if (sources && sources.length > 0) {
            sources.forEach((source) => {
                const option = document.createElement('option');
                option.value = source;
                option.textContent = source;
                sourceSelect.appendChild(option);
            });
        } else {
            sourceSelect.innerHTML = '<option value="">No sources found</option>';
        }
    } catch (err) {
        sourceSelect.innerHTML = '<option value="">Failed to load sources</option>';
    }
}

// Start recording
startBtn.addEventListener('click', async () => {
    const source = sourceSelect.value;
    if (!source) {
        alert('Please select an audio source');
        return;
    }

    try {
        startBtn.disabled = true;
        setStatus('', 'Starting...');

        const res = await fetch('/start', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                source: source,
                duration: chunkDuration
            })
        });

        if (res.ok) {
            isRecording = true;
            settingsEl.classList.add('recording');
            stopBtn.disabled = false;
            showProcessing();

            // Connect SSE after session starts
            connectSSE();
        }
            const err = await res.json();
            startBtn.disabled = false;
            setStatus('error', err.error || 'Failed to start');
    } catch (err) {
        startBtn.disabled = false;
        setStatus('error', 'Error: ' + err.message);
    }
});

// Stop recording
stopBtn.addEventListener('click', async () => {
    try {
        setStatus('', 'Stopping...');
        const res = await fetch('/stop', { method: 'POST' });
        if (res.ok) {
            disconnectSSE();
            onSessionEnded();
        }
    } catch (err) {
        console.error('Error stopping recording:', err);
    }
});

// Initialize
showEmptyState();
setStatus('', 'Ready');
fetchSources();
