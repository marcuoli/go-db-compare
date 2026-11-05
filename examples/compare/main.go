// Command compare demonstrates exact integer comparison without database access.
package main

import (
	"fmt"
	"github.com/marcuoli/go-db-compare/value"
)

func main() {
	result, err := value.Compare(value.Cell{State: value.Known, Kind: value.Integer, Exact: "9007199254740993"}, value.Cell{State: value.Known, Kind: value.Integer, Exact: "9007199254740994"})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s: target-source=%s\n", result.Status, result.Delta)
}
