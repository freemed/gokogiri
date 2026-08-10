package mem

// LIBXML_VERSION is the libxml2 version this was originally built against.
// Pure Go implementation does not require a specific libxml2 version.
const LIBXML_VERSION = "2.0.0"

// LIBXML_NUMERIC_VERSION is a placeholder for the numeric version.
const LIBXML_NUMERIC_VERSION = 0

// AllocSize always returns 0 in the pure Go implementation.
func AllocSize() int {
	return 0
}
