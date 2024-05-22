package rocketmq

type Config struct {
	Endpoint  string `mapstructure:"end_point" json:"end_point" yaml:"end_point"`
	GroupName string `mapstructure:"group_name" json:"group_name" yaml:"group_name"`
	TopicName string `mapstructure:"topic_name" json:"topic_name" yaml:"topic_name"`
	NameSpace string `mapstructure:"name_space" json:"name_space" yaml:"name_space"`
	AccessKey string `mapstructure:"access_key" json:"access_key" yaml:"access_key"`
	SecretKey string `mapstructure:"secret_key" json:"secret_key" yaml:"secret_key"`
}
