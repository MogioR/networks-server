package config

type Config struct {
	Api           Api      `json:"api"`
	Postgres      Postgres `json:"postgres"`
	LogLevel      string   `json:"log_level"`
	IsDevelopment bool     `env:"DEVELOPMENT"`
}

type Postgres struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Db       string `json:"db"`
	Sslmode  string `json:"sslmode"`
	MaxConns int    `json:"max_conns"`
	User     string `env:"PG_USER,notEmpty"`
	Pass     string `env:"PG_PASSWORD,notEmpty"`
}

type Api struct {
	TCPPort      string `json:"tcp_port"`
	HTTPPort     string `json:"http_port"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
	IdleTimeout  int    `json:"idle_timeout"`
}
