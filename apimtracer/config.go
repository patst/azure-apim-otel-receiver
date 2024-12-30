package apimtracer

// Config represents the receiver config settings within the collector's config.yaml
type Config struct {
	FullyQualifiedNamespace string `mapstructure:"fully_qualified_namespace"`
	EventHubName            string `mapstructure:"event_hub_name"`
	ConsumerGroup           string `mapstructure:"consumer_group"`
	StorageContainerUrl     string `mapstructure:"storage_container_url"`
}

// Validate checks if the receiver configuration is valid
func (cfg *Config) Validate() error {
	return nil
}
