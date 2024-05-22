package rocketmq

type Config struct {
	Brokers   []string `mapstructure:"brokers" json:"brokers" yaml:"brokers"`
	GroupName string   `mapstructure:"group_name" json:"group_name" yaml:"group_name"`
	TopicName string   `mapstructure:"topic_name" json:"topic_name" yaml:"topic_name"`
}
