package redis

type RedisConfig struct {
	Master        RedisConnection   `mapstructure:"master" json:"master" yaml:"master"`
	Slaves        []RedisConnection `mapstructure:"slaves" json:"slaves" yaml:"slaves"`
	PoolSize      int               `mapstructure:"pool_size" json:"pool_size" yaml:"pool_size"`
	EnableTracing bool              `mapstructure:"enable_tracing" json:"enable_tracing" yaml:"enable_tracing"`
	EnableMetrics bool              `mapstructure:"enable_metrics" json:"enable_metrics" yaml:"enable_metrics"`
}

type RedisConnection struct {
	Host     string `mapstructure:"host" json:"host" yaml:"host"`
	Port     int    `mapstructure:"port" json:"port" yaml:"port"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
	DB       int    `mapstructure:"db" json:"db" yaml:"db"`
}
