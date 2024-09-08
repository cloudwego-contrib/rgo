package utils

func GetCurrentPathWithUnderline() (string, error) {
	currentPath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Windows-specific path transformation
	currentPath = strings.ReplaceAll(currentPath, ":", "")
	currentPath = strings.ReplaceAll(currentPath, "\\", "_")

	return currentPath, nil
}

func GetDefaultUserPath() string {
	homeDir := os.Getenv("USERPROFILE")
	if homeDir == "" {
		log.Fatal("Cannot get user home directory")
	}
	return homeDir
}
