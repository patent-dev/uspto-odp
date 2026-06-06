package odp_test

import (
	"fmt"

	odp "github.com/patent-dev/uspto-odp"
)

func ExampleNewClient() {
	// The same API key covers the Patent, PTAB, Petition, Bulk Data, and
	// Office Action APIs; TSDR uses a separate key when set.
	config := odp.DefaultConfig()
	config.APIKey = "your-api-key"

	client, err := odp.NewClient(config)
	if err != nil {
		panic(err)
	}
	fmt.Println(client != nil)
	// Output: true
}

func ExampleNormalizePatentNumber() {
	// Patent numbers are normalized offline; resolving a grant or publication
	// number to its application number needs a client call.
	pn, err := odp.NormalizePatentNumber("US 11,646,472 B2")
	if err != nil {
		panic(err)
	}
	fmt.Println(pn.Normalized)
	fmt.Println(pn.FormatAsGrant())
	// Output:
	// 11646472
	// 11,646,472
}
