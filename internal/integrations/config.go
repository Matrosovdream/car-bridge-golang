package integrations

type HTTPConfig struct {
	BaseURL        string
	APIKey         string
	TimeoutSeconds int
	Retries        int
}
