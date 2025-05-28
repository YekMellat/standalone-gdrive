// Package fs provides core functionality for filesystem-like operations
package fs

// Note: This file provides additional pacer functionality beyond what's in fs.go

// SetMaxConnections sets the maximum number of concurrent connections
func (p *Pacer) SetMaxConnections(n int) {
	if p.connTokens != nil {
		// Close old tokens
		close(p.connTokens)
	}

	// Create new tokens
	p.maxConnections = n
	p.connTokens = make(chan struct{}, n)

	// Fill the connection pool
	for i := 0; i < n; i++ {
		p.connTokens <- struct{}{}
	}
}

// GetToken gets a connection token, waiting if necessary
func (p *Pacer) GetToken() {
	<-p.connTokens
}

// PutToken returns a connection token
func (p *Pacer) PutToken() {
	p.connTokens <- struct{}{}
}

// CallWithoutContext calls a function with retries but without a context
// This is a simplified implementation
func (p *Pacer) CallWithoutContext(f func() error) error {
	// Get a token
	p.GetToken()
	defer p.PutToken()

	// Just call the function directly in this simplified implementation
	return f()
}
