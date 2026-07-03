package mongox

import (
	"testing"

	"github.com/stingwolf1080/dynamic-filter/pkg/filter"
)

func BenchmarkParseAndConvertFilter_Simple(b *testing.B) {
	qs := "filter[name][eq]=John&filter[age][gt]=20&page[number]=1&page[size]=10&sort=-created_at&search=test&group_by=status"
	for i := 0; i < b.N; i++ {
		opts := filter.FilterOptions{}
		opts.SetPagination()
		_ = filter.ParseBracketParams(qs, &opts)
		_ = convertFilter(false, &opts)
	}
}

func BenchmarkParseAndConvertFilter_Complex10Fields(b *testing.B) {
	qs := "filter[field1,field2][eq]=val1&filter[field3][in]=1,2,3&filter[field4][gt]=10&filter[field5][lt]=20&filter[field6][eq]=val6&filter[field7,field8][nlike]=test&filter[field9][neq]=null&filter[field10][neq]=true&page[number]=1&page[size]=100&sort=-created_at,updated_at&search=complex&group_by=status"
	for i := 0; i < b.N; i++ {
		opts := filter.FilterOptions{}
		opts.SetPagination()
		_ = filter.ParseBracketParams(qs, &opts)
		_ = convertFilter(false, &opts)
	}
}

func BenchmarkConvertFilterOnly_Complex10Fields(b *testing.B) {
	qs := "filter[field1,field2][eq]=val1&filter[field3][in]=1,2,3&filter[field4][gt]=10&filter[field5][lt]=20&filter[field6][eq]=val6&filter[field7,field8][nlike]=test&filter[field9][neq]=null&filter[field10][neq]=true&page[number]=1&page[size]=100&sort=-created_at,updated_at&search=complex&group_by=status"
	opts := filter.FilterOptions{}
	opts.SetPagination()
	_ = filter.ParseBracketParams(qs, &opts)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertFilter(false, &opts)
	}
}
