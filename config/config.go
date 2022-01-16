package config

// Configurations exported
type Configurations struct {
	Database DatabaseConfigurations
}

// DatabaseConfigurations exported
type DatabaseConfigurations struct {
	ConnectionString string
}
