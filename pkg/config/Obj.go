package config

type Config struct {
	Remark   string `json:"-" toml:"Remark"`
	InitTime int64  `json:"init_time" toml:"InitTime"`
	BasePath string `json:"base_path" toml:"BasePath"`
	DbType   string `json:"database_type" toml:"DbType"`
}
