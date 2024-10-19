package config

type SSHConfig struct {
	ListenAddr string `env:"LISTEN_ADDR"`
	PublicURL  string `env:"PUBLIC_URL"`
}

type DBConfig struct {
	Driver     string `env:"DRIVER"`
	DataSource string `env:"DATA_SOURCE"`
}

type Config struct {
	SSH SSHConfig `envPrefix:"SSH_"`
	DB  DBConfig  `envPrefix:"DB_"`
}

func DefaultConfig() *Config {
	return &Config{
		SSH: SSHConfig{
			ListenAddr: "localhost:23234",
			PublicURL:  "ssh://localhost:23234",
		},
		DB: DBConfig{
			Driver:     "sqlite3",
			DataSource: "./tmp/terminal-pet.db",
		},
	}
}
