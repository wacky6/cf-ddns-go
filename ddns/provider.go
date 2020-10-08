package ddns

// Record type stands for a DNS zone record.
type Record struct {
	// DNS record type (e.g. A, AAA, TXT)
	Type string

	// DNS record content (e.g. 192.168.0.1)
	Content string
}

// Provider interface is responsible for get and set DNS zone records to a DDNS service provider.
type Provider interface {
	// Veirfy if the provided environment variables can be used to access DDNS service correctly.
	// For example, check if the login credentials are correct (can successfully update DNS record).
	VerifyConfig() error

	// Set DNS zone record.
	SetRecord(fqdn string, record Record) error
}
