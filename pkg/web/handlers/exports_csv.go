package handlers

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// countCSVRecordsOptimized compte les enregistrements de manière optimisée
// Utilise une estimation améliorée pour les gros fichiers
func (h *ExportsHandler) countCSVRecordsOptimized(filename string) int {
	info, err := os.Stat(filename)
	if err != nil {
		return 0
	}

	// Pour les fichiers moyens (< 10MB), compter précisément
	if info.Size() < 10*1024*1024 {
		return h.countCSVRecords(filename)
	}

	// Pour les très gros fichiers, utiliser une estimation améliorée
	file, err := os.Open(filename)
	if err != nil {
		return 0
	}
	defer file.Close()

	// Lire un échantillon plus large (500KB) pour plus de précision
	sampleSize := 500000
	if info.Size() < int64(sampleSize) {
		sampleSize = int(info.Size())
	}

	buf := make([]byte, sampleSize)
	n, err := file.Read(buf)
	if err != nil && n == 0 {
		return 0
	}

	// Compter les lignes dans l'échantillon
	lines := strings.Count(string(buf[:n]), "\n")
	if lines <= 1 {
		return 0 // Pas assez de lignes pour une estimation
	}

	// Lire aussi un échantillon du milieu du fichier pour améliorer la précision
	middleOffset := info.Size() / 2
	if middleOffset > int64(sampleSize) {
		_, err := file.Seek(middleOffset, 0)
		if err == nil {
			middleBuf := make([]byte, min(sampleSize/2, int(info.Size()-middleOffset)))
			middleN, err := file.Read(middleBuf)
			if err == nil && middleN > 0 {
				middleLines := strings.Count(string(middleBuf[:middleN]), "\n")
				if middleLines > 0 {
					// Moyenne pondérée des deux échantillons
					totalSampleSize := n + middleN
					totalSampleLines := lines + middleLines
					avgBytesPerLine := float64(totalSampleSize) / float64(totalSampleLines)
					estimatedTotalLines := int(float64(info.Size()) / avgBytesPerLine)

					// Soustraire 1 pour l'en-tête
					if estimatedTotalLines > 1 {
						return estimatedTotalLines - 1
					}
					return 0
				}
			}
		}
	}

	// Fallback vers l'estimation simple si l'échantillon du milieu échoue
	avgBytesPerLine := float64(n) / float64(lines)
	estimatedTotalLines := int(float64(info.Size()) / avgBytesPerLine)

	// Soustraire 1 pour l'en-tête
	if estimatedTotalLines > 1 {
		return estimatedTotalLines - 1
	}
	return 0
}

// min helper function for Go versions < 1.21
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// convertToConfigTimezone converts a time to the configured timezone
func (h *ExportsHandler) convertToConfigTimezone(t time.Time) time.Time {
	if h.config.Export.Timezone == "" || h.config.Export.Timezone == "UTC" {
		return t.UTC()
	}

	loc, err := time.LoadLocation(h.config.Export.Timezone)
	if err != nil {
		h.logger.Warn("web.timezone_load_failed", map[string]interface{}{
			"timezone": h.config.Export.Timezone,
			"error":    err.Error(),
		})
		return t.UTC() // Fallback to UTC
	}

	return t.In(loc)
}

// formatTimeInConfigTimezone formats a time in the configured timezone
func (h *ExportsHandler) formatTimeInConfigTimezone(t time.Time, layout string) string {
	convertedTime := h.convertToConfigTimezone(t)
	return convertedTime.Format(layout)
}

// Version originale conservée pour les petits fichiers
func (h *ExportsHandler) countCSVRecords(filename string) int {
	content, err := os.ReadFile(filename)
	if err != nil {
		return 0
	}

	lines := strings.Split(string(content), "\n")
	// Subtract 1 for header row, and filter out empty lines
	count := 0
	for i, line := range lines {
		if i == 0 { // Skip header
			continue
		}
		if strings.TrimSpace(line) != "" {
			count++
		}
	}

	return count
}

func (h *ExportsHandler) formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
