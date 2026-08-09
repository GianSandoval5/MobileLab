// Package plugin defines MobileLab's language-neutral, out-of-process plugin
// protocol and a small server helper for Go plugin authors.
//
// A plugin handles exactly one request per process. It reads the request from
// standard input, writes the correlated response to standard output, and keeps
// diagnostics on standard error. Plugin executables remain trusted local code;
// this package is a protocol boundary, not an operating-system sandbox.
package plugin
