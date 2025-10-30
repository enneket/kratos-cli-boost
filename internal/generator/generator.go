package generator

import (
	"os"
	"path/filepath"
	"strings"

	"kratos_cli_boost/internal/parser"
)

// createDirectoryIfNotExists creates a directory if it doesn't exist
func createDirectoryIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// getPackageName extracts package name from go package path
func getPackageName(goPackage string) string {
	parts := strings.Split(goPackage, "/")
	return parts[len(parts)-1]
}

// getOutputPath constructs the output path
func getOutputPath(targetDir, layer string, serviceInfo *parser.ServiceInfo) string {
	// Extract module name from go package (assuming go package is in format "module/path/package")
	parts := strings.Split(serviceInfo.GoPackage, "/")
	if len(parts) < 2 {
		return filepath.Join(targetDir, layer)
	}

	// Create directory structure: targetDir/module/.../layer
	modulePath := strings.Join(parts[:len(parts)-1], "/")
	return filepath.Join(targetDir, modulePath, layer)
}
