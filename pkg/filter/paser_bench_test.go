package filter

import (
	"testing"
)

func BenchmarkParseBracketParams_Simple(b *testing.B) {
	qs := "filter[name][eq]=John&filter[age][gt]=20&page[number]=1&page[size]=10&sort=-created_at&search=test&group_by=status"
	for i := 0; i < b.N; i++ {
		opts := &FilterOptions{}
		opts.SetPagination()
		_ = ParseBracketParams(qs, opts)
	}
}

func BenchmarkParseBracketParams_Complex10Fields(b *testing.B) {
	// A more complex query simulating multiple ANDs and ORs across 10 fields.
	// ANDs are represented by separate filter params, ORs by comma separated fields.
	qs := "filter[field1,field2][eq]=val1&filter[field3][in]=1,2,3&filter[field4][gt]=10&filter[field5][lt]=20&filter[field6][eq]=val6&filter[field7,field8][nlike]=test&filter[field9][neq]=null&filter[field10][neq]=true&page[number]=1&page[size]=100&sort=-created_at,updated_at&search=complex&group_by=status"
	for i := 0; i < b.N; i++ {
		opts := &FilterOptions{}
		opts.SetPagination()
		_ = ParseBracketParams(qs, opts)
	}
}

func BenchmarkParseBracketParams_SearchAndGroupBy(b *testing.B) {
	qs := "search=complex&group_by=status"
	for i := 0; i < b.N; i++ {
		opts := &FilterOptions{}
		opts.SetPagination()
		_ = ParseBracketParams(qs, opts)
	}
}

func BenchmarkCheckParams(b *testing.B) {
	qs := "filter[name][eq]=John&filter[age][gt]=20"
	opts := &FilterOptions{
		RestrictField: map[string]bool{"password": true},
	}
	for i := 0; i < b.N; i++ {
		_ = CheckParams(qs, opts)
	}
}

func BenchmarkCleanQuery(b *testing.B) {
	qs := "filter[name][eq]=John&filter[age][gt]=20&page[number]=1"
	for i := 0; i < b.N; i++ {
		_ = cleanQuery(qs)
	}
}
