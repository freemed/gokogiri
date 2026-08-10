package help

// LibxmlInitParser is a no-op in the pure Go implementation.
func LibxmlInitParser() {}

// LibxmlCleanUpParser is a no-op in the pure Go implementation.
func LibxmlCleanUpParser() {}

// LibxmlGetMemoryAllocation always returns 0 in the pure Go implementation.
func LibxmlGetMemoryAllocation() int { return 0 }

// LibxmlCheckMemoryLeak always returns true in the pure Go implementation.
func LibxmlCheckMemoryLeak() bool { return true }

// LibxmlReportMemoryLeak is a no-op in the pure Go implementation.
func LibxmlReportMemoryLeak() {}
