// Global JavaScript for Export Trakt 4 Letterboxd Web Interface

// HTML sanitization utility
function escapeHtml(unsafe) {
  if (typeof unsafe !== 'string') {
    return String(unsafe);
  }
  return unsafe
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

// Escape HTML attributes (more restrictive)
function escapeHtmlAttr(unsafe) {
  if (typeof unsafe !== 'string') {
    return String(unsafe);
  }
  return unsafe
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;')
    .replace(/\//g, '&#x2F;')
    .replace(/=/g, '&#x3D;')
    .replace(/`/g, '&#x60;');
}

// Get CSRF token from cookie
function getCSRFToken() {
  const name = 'csrf_token=';
  const decodedCookie = decodeURIComponent(document.cookie);
  const ca = decodedCookie.split(';');
  for(let i = 0; i < ca.length; i++) {
    let c = ca[i];
    while (c.charAt(0) === ' ') {
      c = c.substring(1);
    }
    if (c.indexOf(name) === 0) {
      return c.substring(name.length, c.length);
    }
  }
  return '';
}

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    initializeApp();
});

function initializeApp() {
    // Initialize real-time updates (SSE/WebSocket)
    initializeRealTimeUpdates();
    
    // Initialize tooltips and interactive elements
    initializeInteractiveElements();
    
    // Initialize keyboard shortcuts
    initializeKeyboardShortcuts();
    
    // Update last updated timestamp
    updateLastUpdatedTime();
}

// WebSocket connection for real-time updates
let websocket = null;
let reconnectInterval = null;

function initializeWebSocket() {
    if (!window.WebSocket) {
        console.log('WebSocket not supported');
        return;
    }
    
    connectWebSocket();
}

function connectWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws/status`;
    
    try {
        websocket = new WebSocket(wsUrl);
        
        websocket.onopen = function() {
            console.log('WebSocket connected');
            clearInterval(reconnectInterval);
            showConnectionStatus('connected');
        };
        
        websocket.onmessage = function(event) {
            const data = JSON.parse(event.data);
            handleWebSocketMessage(data);
        };
        
        websocket.onclose = function() {
            console.log('WebSocket disconnected');
            showConnectionStatus('disconnected');
            // Attempt to reconnect every 5 seconds
            reconnectInterval = setInterval(connectWebSocket, 5000);
        };
        
        websocket.onerror = function(error) {
            console.error('WebSocket error:', error);
            showConnectionStatus('error');
        };
        
    } catch (error) {
        console.error('Failed to create WebSocket connection:', error);
    }
}

function handleWebSocketMessage(data) {
    switch (data.type) {
        case 'status_update':
            updateStatusIndicators(data.payload);
            break;
        case 'export_progress':
            updateExportProgress(data.payload);
            break;
        case 'log_entry':
            addLogEntry(data.payload);
            break;
        case 'alert':
            showAlert(data.payload.type, data.payload.message);
            break;
        default:
            console.log('Unknown WebSocket message type:', data.type);
    }
}

function showConnectionStatus(status) {
    const indicator = document.querySelector('.connection-status');
    if (indicator) {
        indicator.className = `connection-status ${status}`;
        indicator.textContent = status.charAt(0).toUpperCase() + status.slice(1);
    }
}

// Status indicators update
function updateStatusIndicators(data) {
    // Update server status
    updateElement('.server-status', data.serverStatus);
    
    // Update token status
    updateElement('.token-status', data.tokenStatus?.isValid ? 'healthy' : 'error');
    
    // Update API status
    updateElement('.api-status', data.apiStatus);
    
    // Update last updated time
    updateLastUpdatedTime();
}

function updateElement(selector, value) {
    const element = document.querySelector(selector);
    if (element) {
        element.textContent = value;
        element.className = `status-indicator ${value}`;
    }
}

// Export progress updates
function updateExportProgress(data) {
    const progressContainer = document.getElementById('export-progress');
    if (!progressContainer) return;
    
    const progressFill = document.getElementById('progress-fill');
    const progressText = document.getElementById('progress-text');
    const progressPercent = document.getElementById('progress-percent');
    
    if (data.progress !== undefined && progressFill && progressPercent) {
        progressFill.style.width = data.progress + '%';
        progressPercent.textContent = data.progress + '%';
    }
    
    if (data.message && progressText) {
        progressText.textContent = data.message;
    }
    
    if (data.status === 'completed') {
        setTimeout(() => {
            progressContainer.style.display = 'none';
            showAlert('success', 'Export completed successfully!');
            // Refresh exports page if we're on it
            if (window.location.pathname === '/exports') {
                setTimeout(() => location.reload(), 1000);
            }
        }, 2000);
    }
    
    if (data.status === 'failed') {
        showAlert('error', data.error || 'Export failed');
        setTimeout(() => {
            progressContainer.style.display = 'none';
        }, 3000);
    }
}

// Log entry handling
function addLogEntry(logData) {
    const logViewer = document.getElementById('log-viewer');
    if (!logViewer) return;
    
    const logEntry = document.createElement('div');
    logEntry.className = `log-entry log-${logData.level}`;
    
    logEntry.innerHTML = `
        <span class="log-time">${escapeHtml(new Date(logData.time).toLocaleTimeString())}</span>
        <span class="log-level">${escapeHtml(logData.level.toUpperCase())}</span>
        <span class="log-message">${escapeHtml(logData.message)}</span>
        ${logData.context ? `<div class="log-context">${escapeHtml(logData.context)}</div>` : ''}
    `;
    
    logViewer.appendChild(logEntry);
    
    // Keep only the last 100 log entries
    const entries = logViewer.querySelectorAll('.log-entry');
    if (entries.length > 100) {
        entries[0].remove();
    }
    
    // Auto-scroll to bottom
    logViewer.scrollTop = logViewer.scrollHeight;
}

// Alert system
function showAlert(type, message, duration = 5000) {
    // Remove existing alerts
    const existingAlerts = document.querySelectorAll('.alert');
    existingAlerts.forEach(alert => alert.remove());
    
    // Create new alert
    const alert = document.createElement('div');
    alert.className = `alert alert-${type}`;
    
    const icon = getAlertIcon(type);
    alert.innerHTML = `
        <span class="alert-icon">${escapeHtml(icon)}</span>
        <span class="alert-message">${escapeHtml(message)}</span>
        <button class="alert-close" onclick="this.parentElement.remove()">&times;</button>
    `;
    
    // Insert at the top of the container
    const container = document.querySelector('.container');
    if (container) {
        container.insertBefore(alert, container.firstChild);
    }
    
    // Auto-remove after duration
    if (duration > 0) {
        setTimeout(() => {
            if (alert.parentElement) {
                alert.remove();
            }
        }, duration);
    }
}

function getAlertIcon(type) {
    const icons = {
        'success': '✅',
        'error': '❌',
        'warning': '⚠️',
        'info': 'ℹ️'
    };
    return icons[type] || 'ℹ️';
}

// Real-time updates using SSE
let eventSource = null;
let sseReconnectInterval = null;

function initializeRealTimeUpdates() {
    // Try WebSocket first, fallback to SSE
    if (window.WebSocket && false) { // Disable WebSocket for now, use SSE
        initializeWebSocket();
    } else {
        initializeSSE();
    }
}

function initializeSSE() {
    if (!window.EventSource) {
        console.log('SSE not supported, falling back to auto-refresh');
        initializeFallbackRefresh();
        return;
    }
    
    connectSSE();
}

function connectSSE() {
    // Determine appropriate SSE endpoint based on current page
    let sseUrl = '/sse/all';
    if (window.location.pathname === '/status' || window.location.pathname === '/') {
        sseUrl = '/sse/status';
    } else if (window.location.pathname === '/exports') {
        sseUrl = '/sse/export';
    }
    
    try {
        eventSource = new EventSource(sseUrl);
        
        eventSource.onopen = function() {
            console.log('SSE connected');
            clearInterval(sseReconnectInterval);
            showConnectionStatus('connected');
        };
        
        eventSource.onmessage = function(event) {
            try {
                const data = JSON.parse(event.data);
                handleRealtimeMessage({
                    type: 'message',
                    payload: data
                });
            } catch (e) {
                console.error('SSE message parsing error:', e);
            }
        };
        
        // Handle specific event types
        eventSource.addEventListener('status_update', function(event) {
            try {
                const data = JSON.parse(event.data);
                handleRealtimeMessage({
                    type: 'status_update',
                    payload: data
                });
            } catch (e) {
                console.error('SSE status_update parsing error:', e);
            }
        });
        
        eventSource.addEventListener('export_progress', function(event) {
            try {
                const data = JSON.parse(event.data);
                handleRealtimeMessage({
                    type: 'export_progress',
                    payload: data
                });
            } catch (e) {
                console.error('SSE export_progress parsing error:', e);
            }
        });
        
        eventSource.addEventListener('alert', function(event) {
            try {
                const data = JSON.parse(event.data);
                handleRealtimeMessage({
                    type: 'alert',
                    payload: data
                });
            } catch (e) {
                console.error('SSE alert parsing error:', e);
            }
        });
        
        eventSource.addEventListener('ping', function(event) {
            // Keep-alive ping, no action needed
            console.debug('SSE ping received');
        });
        
        eventSource.onerror = function(error) {
            console.error('SSE error:', error);
            showConnectionStatus('error');
            
            // Attempt to reconnect after 5 seconds
            eventSource.close();
            sseReconnectInterval = setInterval(() => {
                console.log('Attempting SSE reconnection...');
                connectSSE();
            }, 5000);
        };
        
    } catch (error) {
        console.error('Failed to create SSE connection:', error);
        initializeFallbackRefresh();
    }
}

function handleRealtimeMessage(data) {
    switch (data.type) {
        case 'status_update':
            updateStatusIndicators(data.payload);
            break;
        case 'export_progress':
            updateExportProgress(data.payload);
            break;
        case 'alert':
            showAlert(data.payload.type, data.payload.message);
            break;
        case 'server_health':
            console.log('Server health update:', data.payload);
            break;
        case 'token_update':
            console.log('Token status updated:', data.payload);
            updateStatusIndicators({ tokenStatus: data.payload });
            break;
        default:
            console.log('Unknown realtime message type:', data.type);
    }
}

// Fallback to old auto-refresh if real-time updates fail
function initializeFallbackRefresh() {
    console.log('Using fallback auto-refresh');
    
    // Auto-refresh status pages every 30 seconds
    if (window.location.pathname === '/status' || window.location.pathname === '/') {
        setInterval(() => {
            if (document.visibilityState === 'visible') {
                refreshStatusData();
            }
        }, 30000);
    }
}

function refreshStatusData() {
    fetch('/api/status')
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                updateStatusIndicators(data.data);
            }
        })
        .catch(error => {
            console.error('Failed to refresh status:', error);
        });
}

// Interactive elements
function initializeInteractiveElements() {
    // Add loading states to buttons (except export buttons which have their own logic)
    document.addEventListener('click', function(event) {
        if (event.target.classList.contains('btn-primary') && 
            !event.target.classList.contains('export-btn')) {
            const btn = event.target;
            const originalText = btn.textContent;
            
            btn.disabled = true;
            btn.textContent = '🔄 Processing...';
            
            // Reset after 30 seconds if not already reset
            setTimeout(() => {
                if (btn.disabled) {
                    btn.disabled = false;
                    btn.textContent = originalText;
                }
            }, 30000);
        }
    });
    
    // Add hover effects to cards
    const cards = document.querySelectorAll('.card, .action-card, .export-type-card');
    cards.forEach(card => {
        card.addEventListener('mouseenter', function() {
            this.style.transform = 'translateY(-2px)';
        });
        
        card.addEventListener('mouseleave', function() {
            this.style.transform = 'translateY(0)';
        });
    });
}

// Keyboard shortcuts
function initializeKeyboardShortcuts() {
    document.addEventListener('keydown', function(event) {
        // Ctrl/Cmd + R: Refresh page
        if ((event.ctrlKey || event.metaKey) && event.key === 'r') {
            // Let the default refresh happen
            return;
        }
        
        // Ctrl/Cmd + 1-4: Navigate to different pages
        if ((event.ctrlKey || event.metaKey) && event.key >= '1' && event.key <= '4') {
            event.preventDefault();
            const pages = ['/', '/exports', '/status', '/config'];
            const index = parseInt(event.key) - 1;
            if (pages[index]) {
                window.location.href = pages[index];
            }
        }
        
        // Escape: Close modals and alerts
        if (event.key === 'Escape') {
            const alerts = document.querySelectorAll('.alert');
            alerts.forEach(alert => alert.remove());
            
            const modals = document.querySelectorAll('.modal');
            modals.forEach(modal => modal.style.display = 'none');
        }
    });
}

// Utility functions
function updateLastUpdatedTime() {
    const element = document.getElementById('last-updated');
    if (element) {
        element.textContent = new Date().toLocaleTimeString();
    }
}

function formatBytes(bytes) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function formatDuration(seconds) {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const remainingSeconds = seconds % 60;
    
    if (hours > 0) {
        return `${hours}h ${minutes}m ${remainingSeconds}s`;
    } else if (minutes > 0) {
        return `${minutes}m ${remainingSeconds}s`;
    } else {
        return `${remainingSeconds}s`;
    }
}

function timeAgo(date) {
    const now = new Date();
    const diffInSeconds = Math.floor((now - date) / 1000);
    
    if (diffInSeconds < 60) {
        return 'just now';
    } else if (diffInSeconds < 3600) {
        const minutes = Math.floor(diffInSeconds / 60);
        return `${minutes} minute${minutes > 1 ? 's' : ''} ago`;
    } else if (diffInSeconds < 86400) {
        const hours = Math.floor(diffInSeconds / 3600);
        return `${hours} hour${hours > 1 ? 's' : ''} ago`;
    } else {
        const days = Math.floor(diffInSeconds / 86400);
        return `${days} day${days > 1 ? 's' : ''} ago`;
    }
}

// API helper functions
function apiRequest(url, options = {}) {
    const defaultOptions = {
        headers: {
            'Content-Type': 'application/json',
        },
    };
    
    const mergedOptions = { ...defaultOptions, ...options };
    
    return fetch(url, mergedOptions)
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return response.json();
        });
}

function downloadFile(url, filename) {
    const link = document.createElement('a');
    link.href = url;
    link.download = filename || '';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
}

// Template helper functions for Handlebars-like functionality
function filename(path) {
    return path.split('/').pop();
}

function mask(str) {
    if (!str || str.length < 8) return str;
    return str.substring(0, 4) + '***' + str.substring(str.length - 4);
}

function title(str) {
    return str.charAt(0).toUpperCase() + str.slice(1);
}

function upper(str) {
    return str.toUpperCase();
}

// Export functions for use in other scripts
window.ExportTraktApp = {
    showAlert,
    apiRequest,
    downloadFile,
    updateStatusIndicators,
    updateExportProgress,
    formatBytes,
    formatDuration,
    timeAgo,
    filename,
    mask,
    title,
    upper
};

// Console welcome message
console.log(`
🎬 Export Trakt 4 Letterboxd Web Interface
=========================================
Version: 1.0.0
Real-time Updates: ${window.EventSource ? 'SSE Enabled' : 'Fallback Mode'}
WebSocket: ${window.WebSocket ? 'Available' : 'Not Supported'}
Local Storage: ${window.localStorage ? 'Available' : 'Not Available'}

Keyboard Shortcuts:
- Ctrl/Cmd + 1: Dashboard
- Ctrl/Cmd + 2: Exports
- Ctrl/Cmd + 3: Status
- Ctrl/Cmd + 4: Config
- Escape: Close alerts/modals

Real-time Features:
- Live status updates
- Export progress tracking
- Instant notifications
`);

// Performance monitoring
if (window.performance && window.performance.mark) {
    window.performance.mark('app-initialized');
}