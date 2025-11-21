package search

import (
	"testing"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType interface{}
	}{
		{
			name:     "Single term",
			input:    "level:error",
			wantType: &query.QueryStringQuery{},
		},
		{
			name:     "AND query",
			input:    "level:error AND service:api",
			wantType: &query.ConjunctionQuery{},
		},
		{
			name:     "Multiple ANDs",
			input:    "a:1 AND b:2 AND c:3",
			wantType: &query.ConjunctionQuery{},
		},
		{
			name:     "Empty string",
			input:    "",
			wantType: &query.QueryStringQuery{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := parseQuery(tt.input)
			assert.IsType(t, tt.wantType, q)
		})
	}
}

func TestQueryStringQueryParentheses(t *testing.T) {
	// Now that we have stripOuterParentheses in parseQuery,
	// we should test parseQuery instead of NewQueryStringQuery directly if we want to verify the fix.
	// But NewQueryStringQuery itself is unchanged.
	// So we should test parseQuery.
	qStr := "(level:error OR level:info)"
	q := parseQuery(qStr) // Use parseQuery which has the fix
	// We can't easily inspect the internal structure of q without casting,
	// but we can check if it parses without error.
	// Better yet, let's verify it against an index in a small test.
	mapping := bleve.NewIndexMapping()
	index, err := bleve.NewMemOnly(mapping)
	require.NoError(t, err)
	defer index.Close()

	data := []struct {
		ID    string `json:"id"`
		Level string `json:"level"`
	}{
		{"1", "error"},
		{"2", "info"},
		{"3", "warn"},
	}

	for _, d := range data {
		index.Index(d.ID, d)
	}

	req := bleve.NewSearchRequest(q)
	res, err := index.Search(req)
	require.NoError(t, err)
	assert.Equal(t, 2, int(res.Total))
}
