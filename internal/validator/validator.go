package validator

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationResult contains the result of code validation
type ValidationResult struct {
	IsValid  bool     `json:"isValid"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
	Score    int      `json:"score"` // 0-100, higher is safer
}

// Validator validates PyQGIS code for security and correctness
type Validator struct {
	dangerousPatterns []*regexp.Regexp
	allowedModules    map[string]bool
	allowedQGIS       map[string]bool
}

// NewValidator creates a new code validator
func NewValidator() *Validator {
	return &Validator{
		dangerousPatterns: compileDangerousPatterns(),
		allowedModules:    getAllowedModules(),
		allowedQGIS:       getAllowedQGISFunctions(),
	}
}

// ValidateCode validates Python code for safety
func (v *Validator) ValidateCode(code string) ValidationResult {
	result := ValidationResult{
		IsValid:  true,
		Errors:   []string{},
		Warnings: []string{},
		Score:    100,
	}

	// Check for dangerous patterns
	for _, pattern := range v.dangerousPatterns {
		if pattern.MatchString(code) {
			result.Errors = append(result.Errors,
				fmt.Sprintf("Найден опасный паттерн: %s", pattern.String()))
			result.IsValid = false
			result.Score -= 50
		}
	}

	// Check imports
	importErrors, importWarnings := v.validateImports(code)
	result.Errors = append(result.Errors, importErrors...)
	result.Warnings = append(result.Warnings, importWarnings...)
	if len(importErrors) > 0 {
		result.IsValid = false
		result.Score -= 30
	}

	// Check for file operations
	if v.hasFileOperations(code) {
		result.Warnings = append(result.Warnings,
			"Код содержит файловые операции - требуется дополнительная проверка")
		result.Score -= 10
	}

	// Check for network operations
	if v.hasNetworkOperations(code) {
		result.Errors = append(result.Errors,
			"Сетевые операции запрещены")
		result.IsValid = false
		result.Score -= 40
	}

	// Check for system calls
	if v.hasSystemCalls(code) {
		result.Errors = append(result.Errors,
			"Системные вызовы запрещены")
		result.IsValid = false
		result.Score -= 50
	}

	// Check for eval/exec abuse
	if v.hasCodeInjection(code) {
		result.Errors = append(result.Errors,
			"Обнаружена попытка инъекции кода (eval/exec)")
		result.IsValid = false
		result.Score -= 50
	}

	// Ensure score doesn't go below 0
	if result.Score < 0 {
		result.Score = 0
	}

	return result
}

// compileDangerousPatterns returns regex patterns for dangerous code
func compileDangerousPatterns() []*regexp.Regexp {
	patterns := []string{
		// File deletion
		`os\.remove\s*\(`,
		`os\.unlink\s*\(`,
		`shutil\.rmtree\s*\(`,
		`pathlib\.Path\s*\([^)]+\)\.unlink\s*\(`,

		// System commands
		`subprocess\.[a-zA-Z_]+\s*\(`,
		`os\.system\s*\(`,
		`os\.popen\s*\(`,
		`commands\.[a-zA-Z_]+\s*\(`,

		// Code execution
		`eval\s*\(`,
		`compile\s*\(`,
		`__import__\s*\(`,

		// File writing (except arcpy temp files)
		`open\s*\([^)]*['"]w['"]`,
		`open\s*\([^)]*['"]a['"]`,

		// Network
		`urllib\.[a-zA-Z_]+`,
		`requests\.[a-zA-Z_]+`,
		`http\.[a-zA-Z_]+`,
		`socket\.[a-zA-Z_]+`,

		// Dangerous builtins
		`globals\s*\(\s*\)`,
		`locals\s*\(\s*\)`,
		`vars\s*\(\s*\)`,
		`delattr\s*\(`,
		`setattr\s*\(`,
	}

	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		compiled = append(compiled, regexp.MustCompile(p))
	}
	return compiled
}

// getAllowedModules returns map of allowed Python modules
func getAllowedModules() map[string]bool {
	return map[string]bool{
		"qgis.core":       true,
		"qgis.processing": true,
		"qgis.gui":        true,
		"qgis.utils":      true,
		"qgis.PyQt5":      true,
		"qgis":            true,
		"processing":      true,
		"PyQt5":           true,
		"os.path":         true, // Read-only path operations
		"math":            true,
		"datetime":        true,
		"json":            true,
		"re":              true,
		"collections":     true,
	}
}

// getAllowedQGISFunctions returns commonly used safe QGIS/PyQGIS functions
func getAllowedQGISFunctions() map[string]bool {
	return map[string]bool{
		// Messaging
		"QgsMessageLog.logMessage": true,
		"QgsMessageLog.Info":       true,
		"QgsMessageLog.Warning":    true,
		"QgsMessageLog.Critical":   true,

		// Processing framework
		"processing.run":                   true,
		"processing.algorithmHelp":         true,
		"processing.createAlgorithmDialog": true,

		// Project
		"QgsProject.instance":       true,
		"QgsProject.mapLayers":      true,
		"QgsProject.addMapLayer":    true,
		"QgsProject.removeMapLayer": true,

		// Layers
		"QgsVectorLayer":               true,
		"QgsRasterLayer":               true,
		"QgsVectorFileWriter":          true,
		"QgsCoordinateReferenceSystem": true,

		// Geometry
		"QgsGeometry": true,
		"QgsPoint":    true,
		"QgsPointXY":  true,
		"QgsFeature":  true,
		"QgsField":    true,
		"QgsFields":   true,

		// Utils
		"iface.mapCanvas":      true,
		"iface.activeLayer":    true,
		"iface.addVectorLayer": true,
	}
}

// validateImports checks if imports are allowed
func (v *Validator) validateImports(code string) ([]string, []string) {
	errors := []string{}
	warnings := []string{}

	// Find all import statements
	importPattern := regexp.MustCompile(`(?m)^\s*(?:import|from)\s+([a-zA-Z0-9_.]+)`)
	matches := importPattern.FindAllStringSubmatch(code, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		module := match[1]

		// Check if module is allowed
		if !v.isModuleAllowed(module) {
			errors = append(errors,
				fmt.Sprintf("Модуль '%s' не разрешен", module))
		}
	}

	return errors, warnings
}

// isModuleAllowed checks if a module is in the allowed list
func (v *Validator) isModuleAllowed(module string) bool {
	// Check exact match
	if v.allowedModules[module] {
		return true
	}

	// Check if it's a submodule of qgis
	if strings.HasPrefix(module, "qgis") {
		return true
	}

	// Check if it's a submodule of PyQt5 (used by QGIS)
	if strings.HasPrefix(module, "PyQt5") {
		return true
	}

	// Check if it's a submodule of allowed module
	for allowed := range v.allowedModules {
		if strings.HasPrefix(module, allowed+".") {
			return true
		}
	}

	return false
}

// hasFileOperations checks for file operations
func (v *Validator) hasFileOperations(code string) bool {
	patterns := []string{
		`open\s*\(`,
		`\.write\s*\(`,
		`\.read\s*\(`,
	}

	for _, p := range patterns {
		if matched, _ := regexp.MatchString(p, code); matched {
			return true
		}
	}

	return false
}

// hasNetworkOperations checks for network operations
func (v *Validator) hasNetworkOperations(code string) bool {
	patterns := []string{
		`urllib`,
		`requests`,
		`http\.client`,
		`socket\s*\(`,
		`urlopen`,
	}

	for _, p := range patterns {
		if matched, _ := regexp.MatchString(p, code); matched {
			return true
		}
	}

	return false
}

// hasSystemCalls checks for system calls
func (v *Validator) hasSystemCalls(code string) bool {
	patterns := []string{
		`subprocess`,
		`os\.system`,
		`os\.popen`,
		`os\.spawn`,
		`commands\.`,
	}

	for _, p := range patterns {
		if matched, _ := regexp.MatchString(p, code); matched {
			return true
		}
	}

	return false
}

// hasCodeInjection checks for eval/exec abuse
func (v *Validator) hasCodeInjection(code string) bool {
	// Allow exec only if it's in arcpy context
	if strings.Contains(code, "exec(") {
		// This is dangerous - only our controlled exec should be used
		return true
	}

	if strings.Contains(code, "eval(") {
		return true
	}

	if strings.Contains(code, "compile(") {
		return true
	}

	return false
}
