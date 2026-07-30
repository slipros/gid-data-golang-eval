// Eval GID-212: settings.allow-results extends the contract for builders that
// do not produce SQL — a search-engine DSL builder has no args []any.
// The allowed list here: "(string, error)", "string", "[]string",
// "(*omd.SearchParams, error)".
package build

import (
	"omd"
)

// --- Negative class: signatures listed in allow-results pass ---

// A ready JSON DSL: (string, error) — allowed by the settings.
func DatasetFilterDSL(schemaFQN string) (string, error) {
	return `{"query":{}}`, nil
}

// A Lucene fragment: string — allowed by the settings.
func WildcardQuery(query string) string {
	return "*" + query + "*"
}

// Query parts: []string — allowed by the settings.
func DatasetsQueryParts(query string) []string {
	return []string{query}
}

// Search-client params: (*omd.SearchParams, error) — allowed by the settings;
// the package is spelled by its name (omd), not by its import path.
func SearchParamsForDatasets(schemaFQN string) (*omd.SearchParams, error) {
	return &omd.SearchParams{Query: schemaFQN}, nil
}

// The built-in contract keeps working alongside the settings.
func SelectJobs(status string) (string, []any, error) {
	return "SELECT 1", []any{status}, nil
}

// --- Positive class: a signature outside both the contract and allow-results ---

// (int, error) is in neither list — still a violation.
func CountDatasets() (int, error) { // want `GID-212: a build function must return \(sql string, args \[\]any, err error\) or \(\*batch\.Batch, error\)\. Fix: adjust the signature`
	return 0, nil
}

// A bare *omd.SearchParams without error: the allowed entry is the pair
// (*omd.SearchParams, error) — a partial match is not a match.
func SearchParamsBare() *omd.SearchParams { // want `GID-212: a build function must return \(sql string, args \[\]any, err error\) or \(\*batch\.Batch, error\)\. Fix: adjust the signature`
	return nil
}
