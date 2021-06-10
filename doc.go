// Package sips provides structures for implementing an IPFS pinning service.
//
// The primary purpose of this package is to allow a user to create an
// IPFS pinning service with minimal effort. The package is based
// around the PinHandler interface. An implementation of this
// interface can be passed to the Handler function in order to create
// an HTTP handler that serves a valid pinning service.
package sips
