package config

// Config 全局配置根结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Log      LogConfig      `mapstructure:"log"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Redis    RedisConfig    `mapstructure:"redis"`
	EveSSO   EveSSOConfig   `mapstructure:"eve_sso"`
	SDE      SDEConfig      `mapstructure:"sde"`
}

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug | release | test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	Charset      string `mapstructure:"charset"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`       // debug | info | warn | error
	Filename   string `mapstructure:"filename"`    // 日志文件路径
	MaxSize    int    `mapstructure:"max_size"`    // 单文件最大 MB
	MaxBackups int    `mapstructure:"max_backups"` // 最多保留文件数
	MaxAge     int    `mapstructure:"max_age"`     // 最多保留天数
	Compress   bool   `mapstructure:"compress"`    // 是否压缩归档
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret    string `mapstructure:"secret"`
	ExpireDay int    `mapstructure:"expire_day"`
}

// RedisConfig Redis 缓存配置
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`      // 地址，如 127.0.0.1:6379
	Password string `mapstructure:"password"`  // 无密码留空
	DB       int    `mapstructure:"db"`        // 默认 DB 0
	PoolSize int    `mapstructure:"pool_size"` // 连接池大小
}

// EveSSOConfig EVE Online SSO 配置
type EveSSOConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	CallbackURL  string `mapstructure:"callback_url"`
}

// SDEConfig SDE 模块配置
type SDEConfig struct {
	APIKey string `mapstructure:"api_key"` // 用于保护数据查询接口的 API Key
	Proxy  string `mapstructure:"proxy"`   // 下载 SDE 时使用的 HTTP/SOCKS5 代理，留空则不使用
}
