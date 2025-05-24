// Global app object
window.App = {
  config: {},
  ws: null,

  // Initialize the application
  init: function () {
    this.setupEventListeners();
    this.initializeRefreshButton();
    console.log("Export Trakt Web UI initialized");
  },

  // Setup global event listeners
  setupEventListeners: function () {
    // Handle refresh button
    const refreshBtn = document.getElementById("refresh-btn");
    if (refreshBtn) {
      refreshBtn.addEventListener("click", () => {
        window.location.reload();
      });
    }

    // Handle escape key for modals
    document.addEventListener("keydown", (e) => {
      if (e.key === "Escape") {
        this.closeAllModals();
      }
    });
  },

  // Initialize refresh button
  initializeRefreshButton: function () {
    const refreshBtn = document.getElementById("refresh-btn");
    if (refreshBtn) {
      refreshBtn.addEventListener("click", (e) => {
        e.preventDefault();
        const icon = refreshBtn.querySelector("i");
        if (icon) {
          icon.classList.add("fa-spin");
          setTimeout(() => {
            icon.classList.remove("fa-spin");
          }, 1000);
        }
      });
    }
  },

  // Close all modals
  closeAllModals: function () {
    const modals = document.querySelectorAll(".modal");
    modals.forEach((modal) => {
      modal.style.display = "none";
    });
  },
};

// Toast notification system
window.Toast = {
  container: null,

  init: function () {
    this.container = document.getElementById("toast-container");
    if (!this.container) {
      this.container = document.createElement("div");
      this.container.id = "toast-container";
      this.container.className = "toast-container";
      document.body.appendChild(this.container);
    }
  },

  show: function (message, type = "info", duration = 5000) {
    if (!this.container) this.init();

    const toast = document.createElement("div");
    toast.className = `toast ${type}`;

    const icon = this.getIcon(type);
    toast.innerHTML = `
            <div style="display: flex; align-items: center; gap: 0.5rem;">
                <i class="fas fa-${icon}"></i>
                <span>${message}</span>
            </div>
        `;

    this.container.appendChild(toast);

    // Auto remove after duration
    setTimeout(() => {
      if (toast.parentNode) {
        toast.style.animation = "slideOut 0.3s ease";
        setTimeout(() => {
          if (toast.parentNode) {
            this.container.removeChild(toast);
          }
        }, 300);
      }
    }, duration);

    // Click to dismiss
    toast.addEventListener("click", () => {
      if (toast.parentNode) {
        this.container.removeChild(toast);
      }
    });
  },

  getIcon: function (type) {
    switch (type) {
      case "success":
        return "check-circle";
      case "error":
        return "exclamation-triangle";
      case "warning":
        return "exclamation-circle";
      case "info":
      default:
        return "info-circle";
    }
  },
};

// Global toast function for convenience
window.showToast = function (message, type, duration) {
  Toast.show(message, type, duration);
};

// API helper functions
window.API = {
  baseURL: "/api/v1",

  // Generic fetch wrapper
  request: function (endpoint, options = {}) {
    const url = this.baseURL + endpoint;
    const defaultOptions = {
      headers: {
        "Content-Type": "application/json",
      },
    };

    const finalOptions = { ...defaultOptions, ...options };

    return fetch(url, finalOptions)
      .then((response) => {
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        return response.json();
      })
      .catch((error) => {
        console.error("API request failed:", error);
        showToast(`API Error: ${error.message}`, "error");
        throw error;
      });
  },

  // GET request
  get: function (endpoint) {
    return this.request(endpoint, { method: "GET" });
  },

  // POST request
  post: function (endpoint, data) {
    return this.request(endpoint, {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  // PUT request
  put: function (endpoint, data) {
    return this.request(endpoint, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  },

  // DELETE request
  delete: function (endpoint) {
    return this.request(endpoint, { method: "DELETE" });
  },
};

// WebSocket connection manager
window.WebSocket = {
  connection: null,
  reconnectAttempts: 0,
  maxReconnectAttempts: 5,
  reconnectDelay: 1000,

  connect: function () {
    if (this.connection && this.connection.readyState === WebSocket.OPEN) {
      return;
    }

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const url = `${protocol}//${window.location.host}/api/v1/ws`;

    try {
      this.connection = new WebSocket(url);

      this.connection.onopen = () => {
        console.log("WebSocket connected");
        this.reconnectAttempts = 0;
        showToast("Real-time connection established", "success", 3000);
      };

      this.connection.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          this.handleMessage(data);
        } catch (error) {
          console.error("Failed to parse WebSocket message:", error);
        }
      };

      this.connection.onclose = () => {
        console.log("WebSocket disconnected");
        this.scheduleReconnect();
      };

      this.connection.onerror = (error) => {
        console.error("WebSocket error:", error);
      };
    } catch (error) {
      console.error("Failed to create WebSocket connection:", error);
      this.scheduleReconnect();
    }
  },

  scheduleReconnect: function () {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay =
        this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

      console.log(
        `Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`
      );
      setTimeout(() => {
        this.connect();
      }, delay);
    } else {
      console.log("Max reconnection attempts reached");
      showToast("Real-time connection lost", "warning");
    }
  },

  handleMessage: function (data) {
    // Handle different message types
    switch (data.type) {
      case "export_started":
        showToast("Export started", "info");
        break;
      case "export_completed":
        showToast("Export completed successfully", "success");
        break;
      case "export_failed":
        showToast("Export failed", "error");
        break;
      case "health_update":
        this.updateHealthStatus(data.payload);
        break;
      default:
        console.log("Unknown WebSocket message type:", data.type);
    }
  },

  updateHealthStatus: function (healthData) {
    const statusElement = document.getElementById("system-status");
    if (statusElement) {
      statusElement.textContent = healthData.status;
      statusElement.className = `status ${healthData.status}`;
    }
  },

  send: function (data) {
    if (this.connection && this.connection.readyState === WebSocket.OPEN) {
      this.connection.send(JSON.stringify(data));
    } else {
      console.warn("WebSocket not connected, cannot send message");
    }
  },

  disconnect: function () {
    if (this.connection) {
      this.connection.close();
      this.connection = null;
    }
  },
};

// Utility functions
window.Utils = {
  // Format bytes to human readable format
  formatBytes: function (bytes, decimals = 2) {
    if (bytes === 0) return "0 Bytes";

    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ["Bytes", "KB", "MB", "GB", "TB"];

    const i = Math.floor(Math.log(bytes) / Math.log(k));

    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + " " + sizes[i];
  },

  // Format duration to human readable format
  formatDuration: function (seconds) {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;

    if (days > 0) {
      return `${days}d ${hours}h ${minutes}m`;
    } else if (hours > 0) {
      return `${hours}h ${minutes}m`;
    } else if (minutes > 0) {
      return `${minutes}m ${secs}s`;
    } else {
      return `${secs}s`;
    }
  },

  // Debounce function
  debounce: function (func, wait, immediate) {
    let timeout;
    return function executedFunction() {
      const context = this;
      const args = arguments;

      const later = function () {
        timeout = null;
        if (!immediate) func.apply(context, args);
      };

      const callNow = immediate && !timeout;
      clearTimeout(timeout);
      timeout = setTimeout(later, wait);

      if (callNow) func.apply(context, args);
    };
  },

  // Throttle function
  throttle: function (func, limit) {
    let inThrottle;
    return function () {
      const args = arguments;
      const context = this;
      if (!inThrottle) {
        func.apply(context, args);
        inThrottle = true;
        setTimeout(() => (inThrottle = false), limit);
      }
    };
  },
};

// Initialize when DOM is loaded
document.addEventListener("DOMContentLoaded", function () {
  App.init();
  Toast.init();

  // Connect WebSocket if not on configuration page
  if (!window.location.pathname.includes("/config")) {
    WebSocket.connect();
  }
});

// Cleanup on page unload
window.addEventListener("beforeunload", function () {
  WebSocket.disconnect();
});
