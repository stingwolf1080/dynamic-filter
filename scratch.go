package main
import (
	"fmt"
	"github.com/stingwolf1080/dynamic-filter/pkg/filter"
)
func main() {
	var opts filter.FilterOptions
	err := filter.ParseBracketParams(`filter[age]=[gte]:24&filter[name]=[eq]:Alice,[eq]:Dave`, &opts)
	if err != nil {
		fmt.Println("Err:", err)
		return
	}
	fmt.Printf("%+v\n", opts.Filter)
}
