package main

type ElasticConfig struct {
	URL        string
	Index      string
	APIKey     string
	VerifyCert bool
}

func LoadElasticConfig() ElasticConfig {
	return ElasticConfig{
		URL:        "xxxxx",
		Index:      "website-backend",
		APIKey:     "xxxxx==",
		VerifyCert: false,
	}
}
