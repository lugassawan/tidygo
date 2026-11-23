package allowed

// Handler is allowlisted, so it is silent even though it does not match ServeHTTP.
type Handler interface{ ServeHTTP() }

// Names that are not allowlisted are still checked.
type Server interface{ ServeHTTP() } // want `single-method interface Server should be named after its method ServeHTTP \(e\.g\. ServerHTTP\)`
