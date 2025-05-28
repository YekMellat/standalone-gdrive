// Package pacer makes pacing and retrying API calls easy
package pacer

import (
	"sync"
	"time"
)

// State represents the public Pacer state that will be passed to the
// configured Calculator
type State struct {
	SleepTime          time.Duration // current time to sleep before adding the pacer token back
	ConsecutiveRetries int           // number of consecutive retries, will be 0 when the last invoker call returned false
	LastError          error         // the error returned by the last invoker call or nil
}

// Calculator is a generic calculation function for a Pacer.
type Calculator interface {
	// Calculate takes the current Pacer state and returns the sleep time after which
	// the next Pacer call will be done.
	Calculate(state State) time.Duration
}

// Pacer is the primary type of the pacer package. It allows to retry calls
// with a configurable delay in between.
type Pacer struct {
	pacerOptions
	mu         sync.Mutex    // Protecting read/writes
	pacer      chan struct{} // To pace the operations
	connTokens chan struct{} // Connection tokens
	state      State
}

type pacerOptions struct {
	maxConnections int         // Maximum number of concurrent connections
	retries        int         // Max number of retries
	calculator     Calculator  // switchable pacing algorithm - call with mu held
	invoker        InvokerFunc // wrapper function used to invoke the target function
}

// InvokerFunc is the signature of the wrapper function used to invoke the
// target function in Pacer.
type InvokerFunc func(try, tries int, f Paced) (bool, error)

// Option can be used in New to configure the Pacer.
type Option func(*pacerOptions)

// CalculatorOption sets a Calculator for the new Pacer.
func CalculatorOption(c Calculator) Option {
	return func(p *pacerOptions) { p.calculator = c }
}

// RetriesOption sets the number of retries for the new Pacer.
func RetriesOption(retries int) Option {
	return func(p *pacerOptions) { p.retries = retries }
}

// MaxConnectionsOption sets the number of concurrent connections for the new
// Pacer.
func MaxConnectionsOption(maxConnections int) Option {
	return func(p *pacerOptions) { p.maxConnections = maxConnections }
}

// InvokerOption sets the invoker func that will wrapper the inner function
// func for the pacer.
func InvokerOption(i InvokerFunc) Option {
	return func(p *pacerOptions) { p.invoker = i }
}

// Paced in the internal interface for the calls to the pacer.
type Paced func() (bool, error)

// New creates a Pacer with default values and executes the options.
func New(options ...Option) *Pacer {
	p := &Pacer{
		pacerOptions: pacerOptions{
			maxConnections: 8,
			retries:        10,
			calculator:     &DefaultCalculator{},
			invoker:        DefaultInvoker,
		},
	}

	// apply custom options
	for _, option := range options {
		option(&p.pacerOptions)
	}

	p.pacer = make(chan struct{}, 1)
	// Fill the channel with 1 token
	p.pacer <- struct{}{}
	p.connTokens = make(chan struct{}, p.maxConnections)
	// Fill the channel with maxConnections tokens
	for i := 0; i < p.maxConnections; i++ {
		p.connTokens <- struct{}{}
	}

	return p
}

// NewGoogleDrive creates a Google Drive specific pacer
func NewGoogleDrive(options ...Option) *Pacer {
	return New(options...)
}

// DefaultCalculator is a Calculator implementation that provide the default
// behaviour.
type DefaultCalculator struct {
	minSleep      time.Duration // minimum sleep time
	maxSleep      time.Duration // maximum sleep time
	decayConstant uint          // decay constant
	burst         int           // number of calls to allow without sleeping
}

// MinSleep sets the minimum sleep time
func MinSleep(t time.Duration) Option {
	return func(p *pacerOptions) {
		if c, ok := p.calculator.(*DefaultCalculator); ok {
			c.minSleep = t
		}
	}
}

// MaxSleep sets the maximum sleep time
func MaxSleep(t time.Duration) Option {
	return func(p *pacerOptions) {
		if c, ok := p.calculator.(*DefaultCalculator); ok {
			c.maxSleep = t
		}
	}
}

// DecayConstant sets the decay constant
func DecayConstant(t uint) Option {
	return func(p *pacerOptions) {
		if c, ok := p.calculator.(*DefaultCalculator); ok {
			c.decayConstant = t
		}
	}
}

// Burst sets the burst count
func Burst(t int) Option {
	return func(p *pacerOptions) {
		if c, ok := p.calculator.(*DefaultCalculator); ok {
			c.burst = t
		}
	}
}

// Calculate calculates the next sleep time based on the State
func (c *DefaultCalculator) Calculate(state State) time.Duration {
	if state.ConsecutiveRetries == 0 {
		return 0
	}
	if state.ConsecutiveRetries == 1 {
		return c.minSleep
	}
	sleepTime := state.SleepTime << c.decayConstant
	if sleepTime < c.minSleep {
		sleepTime = c.minSleep
	}
	if c.maxSleep > 0 && sleepTime > c.maxSleep {
		sleepTime = c.maxSleep
	}
	return sleepTime
}

// DefaultInvoker is the default InvokerFunc used by Pacer
func DefaultInvoker(try, tries int, paced Paced) (bool, error) {
	again, err := paced()
	if try >= tries {
		return false, err
	}
	return again, err
}

// Call runs f in a paced way
//
// It calls f and then waits the appropriate time before continuing.
// If the function f returns true then another call will be scheduled
// after the pacing sleep.
//
// This is useful for rate limiting and error handling.
//
// This implements calls with retries in a different way to the
// existing call, Call.
//
// If the function f returns an error then after a short delay Call
// will retry the operation. By default it will retry 10 times, but
// this can be changed with the Retries method.
//
// If the function f returns true then Call will sleep for the
// calculated time and then repeat the operation (not counting it as a
// retry).
//
// The error return from Call is the error (if any) returned from the
// last call of f.
func (p *Pacer) Call(f Paced) error {
	var (
		err                error
		again              bool
		consecutiveRetries int
	)

	for try := 0; try <= p.retries; try++ {
		p.mu.Lock()
		// Get a pacer token
		<-p.pacer
		// Get a connection token
		<-p.connTokens
		// Do the operation
		again, err = p.invoker(consecutiveRetries, p.retries, f)
		// Return the connection token
		p.connTokens <- struct{}{}
		// If ok or no more retries, restore the pacer and return
		if !again || try >= p.retries {
			p.pacer <- struct{}{}
			p.state.ConsecutiveRetries = 0
			p.state.LastError = nil
			p.mu.Unlock()
			return err
		}

		// Calculate delay
		p.state.ConsecutiveRetries++
		p.state.LastError = err
		sleepTime := p.calculator.Calculate(p.state)
		p.state.SleepTime = sleepTime
		p.mu.Unlock()

		// If the retry function returned true, indicate consecutive retries
		if again {
			consecutiveRetries++
		} else {
			consecutiveRetries = 0
		}

		// Sleep for the required time
		time.Sleep(sleepTime)
		p.mu.Lock()
		// Return a token to the pacer
		p.pacer <- struct{}{}
		p.mu.Unlock()

		if !again {
			return err
		}
	}

	// Unreachable
	return err
}
