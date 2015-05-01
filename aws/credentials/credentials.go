// Package provides Credential retrieval and management
//
// The Credentials is the primary method of getting access to and managing
// credentials Values. Using dependency injection retrieval of the credential
// values is handled by a object which satisfies the Provider interface.
//
// By default the Credentials.Get() will cache the successful result of a
// Provider's Retrieve() until Provider.IsExpired() returns true. At which
// point Credentials will call Provider's Retrieve() to get new credential Value.
//
// The Provider is responsible for determining when credentials Value have expired.
// It is also important to note that Credentials will always call Retrieve the
// first time Credentials.Get() is called.
//
// Example of using the environment variable credentials.
//
//     creds := NewEnvCredentials()
//
//     // Retrieve the credentials value
//     credValue, err := creds.Get()
//     if err != nil {
//         // handle error
//     }
//
// Example of forcing credentials to expire and be refreshed on the next Get().
// This may be helpful to proactively expire credentials and refresh them sooner
// than they would naturally expire on their own.
//
//     creds := NewCredentials(&EC2RoleProvider{})
//     creds.Expire()
//     credsValue, err := creds.Get()
//     // New credentials will be retrieved instead of from cache.
//
//
// Custom Provider
//
// Each Provider built into this package also provides a helper method to generate
// a Credentials pointer setup with the provider. To use a custom Provider just
// create a type which satisfies the Provider interface and pass it to the
// NewCredentials method.
//
//     type MyProvider struct{}
//     func (m *MyProvider) Retrieve() (Value, error) {...}
//     func (m *MyProvider) IsExpired() bool {...}
//
//     creds := NewCredentials(&MyProvider{})
//     credValue, err := creds.Get()
//
package credentials

import (
	"sync"
	"time"
)

// A Value is the AWS credentials value for individual credential fields.
type Value struct {
	// AWS Access key ID
	AccessKeyID     string
	// AWS Secret Access Key
	SecretAccessKey string
	// AWS Session Token
	SessionToken    string
}

// A Provider is the interface for any component which will provide credentials
// Value. A provider is required to manage its own Expired state, and what to
// be expired means.
//
// The Provider should not need to implement its own mutexes, because
// that will be managed by Credentials.
type Provider interface {
	// Refresh returns nil if it successfully retrieved the value.
	// Error is returned if the value were not obtainable, or empty.
	Retrieve() (Value, error)

	// IsExpired returns if the credentials are no longer valid, and need
	// to be retrieved.
	IsExpired() bool
}

// A Credentials provides synchronous safe retrieval of AWS credentials Value.
// Credentials will cache the credentials value until they expire. Once the value
// expires the next Get will attempt to retrieve valid credentials.
//
// Credentials is safe to use across multiple goroutines and will manage the
// synchronous state so the Providers do not need to implement their own
// synchronization.
//
// The first Credentials.Get() will always call Provider.Retrieve() to get the
// first instance of the credentials Value. All calls to Get() after that
// will return the cached credentials Value until IsExpired() returns true.
type Credentials struct {
	creds        Value
	forceRefresh bool
	m            sync.Mutex

	provider Provider
}

// NewCredentials returns a pointer to a new Credentials with the provider set.
func NewCredentials(provider Provider) *Credentials {
	return &Credentials{
		provider: provider,
		forceRefresh: true,
	}
}

// Get returns the credentials value, or error if the credentials Value failed
// to be retrieved.
//
// Will return the cached credentials Value if it has not expired. If the
// credentials Value has expired the Provider's Retrieve() will be called
// to refresh the credentials.
//
// If Credentials.Expire() was called the credentials Value will be force
// expired, and the next call to Get() will cause them to be refreshed.
func (c *Credentials) Get() (Value, error) {
	c.m.Lock()
	defer c.m.Unlock()

	if c.isExpired() {
		creds, err := c.provider.Retrieve()
		if err != nil {
			return Value{}, err
		}
		c.creds = creds
		c.forceRefresh = false
	}

	return c.creds, nil
}

// Expire expires the credentials and forces them to be retrieved on the
// next call to Get().
//
// This will override the Provider's expired state, and force Credentials
// to call the Provider's Retrieve().
func (c *Credentials) Expire() {
	c.m.Lock()
	defer c.m.Unlock()

	c.forceRefresh = true
}

// IsExpired returns if the credentials are no longer valid, and need
// to be retrieved.
//
// If the Credentials were forced to be expired with Expire() this will
// reflect that override.
func (c *Credentials) IsExpired() bool {
	c.m.Lock()
	defer c.m.Unlock()

	return c.isExpired()
}

// isExpired helper method wrapping the definition of expired credentials.
func (c *Credentials) isExpired() bool {
	return c.forceRefresh || c.provider.IsExpired()
}

// Provide a stub-able time.Now for unit tests so expiry can be tested.
var currentTime = time.Now