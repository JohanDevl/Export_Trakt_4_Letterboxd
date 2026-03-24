package handlers

import (
	"time"
)

// getExportsWithCache retourne les exports avec mise en cache intelligente
func (h *ExportsHandler) getExportsWithCache(page, limit int) []ExportItem {
	h.cache.mu.RLock()
	cacheValid := time.Since(h.cache.lastScan) < h.cache.cacheTTL && len(h.cache.exports) > 0
	h.cache.mu.RUnlock()

	if cacheValid {
		h.logger.Info("web.exports_cache_hit", map[string]interface{}{
			"cached_count": len(h.cache.exports),
		})
		return h.cache.exports
	}

	// Cache miss - scanner les exports avec lazy loading
	exports := h.scanExportFilesOptimized()

	// Mettre à jour le cache
	h.cache.mu.Lock()
	h.cache.exports = exports
	h.cache.lastScan = time.Now()
	h.cache.mu.Unlock()

	h.logger.Info("web.exports_cache_updated", map[string]interface{}{
		"total_exports": len(exports),
	})

	return exports
}
