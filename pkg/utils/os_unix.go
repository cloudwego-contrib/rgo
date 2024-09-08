package utils

import (
	"log"
	"os"
	"strings"
)

func GetCurrentPathWithUnderline() (string, error) {
	currentPath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	currentPath = strings.TrimSpace(currentPath)

	strings.TrimPrefix(currentPath, "/")
	currentPath = strings.ReplaceAll(currentPath, "/", "_")

	return currentPath, nil
}

func GetDefaultUserPath() string {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		log.Fatal("Cannot get user home directory")
	}
	return homeDir
}
