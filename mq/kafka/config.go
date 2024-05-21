package kafka

// Config kafka config
type Config struct {
	Brokers           []string `mapstructure:"brokers" json:"brokers" yaml:"brokers"`
	GroupID           string   `mapstructure:"group_id" json:"group_id" yaml:"group_id"`
	TopicName         string   `mapstructure:"topic_name" json:"topic_name" yaml:"topic_name"`
	Partition         int32    `mapstructure:"partition" json:"partition" yaml:"partition"`
	ReplicationFactor int      `mapstructure:"replication_factor" json:"replication_factor" yaml:"replication_factor"`
	MaxMessageSize    int64    `mapstructure:"max_message_size" json:"max_message_size" yaml:"max_message_size"`
	MinMessageSize    int64    `mapstructure:"min_message_size" json:"min_message_size" yaml:"min_message_size"`
}
