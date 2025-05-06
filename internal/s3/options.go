package s3

// Options опции для подключения к s3 хранилищу.
type Options struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Secure          bool
	RetriesCount    int
	RetryTimeout    int
}
