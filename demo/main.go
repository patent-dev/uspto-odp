// Package main provides a comprehensive demo of the USPTO ODP Go client.
//
// This demo showcases ALL 38 endpoints of the USPTO ODP API:
// - 13 Patent Application API endpoints
// - 3 Bulk Data API endpoints
// - 3 Petition API endpoints
// - 19 PTAB (Patent Trial and Appeal Board) API endpoints
//
// All request/response pairs are saved to demo/examples/ for reference.
//
// Usage:
//
//	export USPTO_API_KEY="your-api-key"
//
//	# Run all demos
//	./demo
//
//	# Run specific service demos
//	./demo -service=patent
//	./demo -service=ptab
//
//	# Skip saving examples
//	./demo -no-save
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	odp "github.com/patent-dev/uspto-odp"
)

var (
	apiKey      = flag.String("key", os.Getenv("USPTO_API_KEY"), "USPTO API key")
	patent      = flag.String("patent", TestPatentApp, "Patent number")
	service     = flag.String("service", "", "Service filter (patent|bulk|petition|ptab|xml)")
	endpoint    = flag.String("endpoint", "", "Specific endpoint to run")
	interactive = flag.Bool("interactive", false, "Run in interactive mode")
	examplesDir = flag.String("examples", "examples", "Directory to save examples")
	skipSave    = flag.Bool("no-save", false, "Skip saving request/response files")
)

// DemoContext holds shared context for all demos
type DemoContext struct {
	Client   *odp.Client
	Ctx      context.Context
	Saver    *ExampleSaver
	Patent   string
	SkipSave bool
}

func main() {
	flag.Parse()

	if *apiKey == "" {
		log.Fatal("API key required. Set USPTO_API_KEY environment variable or use -key flag")
	}

	config := odp.DefaultConfig()
	config.APIKey = *apiKey
	client, err := odp.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Create demo context
	dctx := &DemoContext{
		Client:   client,
		Ctx:      ctx,
		Patent:   *patent,
		SkipSave: *skipSave,
	}
	if !*skipSave {
		dctx.Saver = NewExampleSaver(*examplesDir)
	}

	if *interactive {
		runInteractive(ctx, client)
		return
	}

	switch *service {
	case "patent":
		demoPatentWithContext(dctx, *patent)
	case "bulk":
		if *interactive {
			reader := bufio.NewReader(os.Stdin)
			demoBulk(ctx, client, reader)
		} else {
			demoBulkWithContext(dctx)
		}
	case "petition":
		demoPetitionWithContext(dctx)
	case "ptab":
		demoPTABWithContext(dctx)
	case "xml":
		reader := bufio.NewReader(os.Stdin)
		demoXML(ctx, client, reader)
	case "":
		runAllWithContext(dctx)
	default:
		log.Fatalf("Unknown service: %s", *service)
	}
}

func runAll(ctx context.Context, client *odp.Client, patent string) {
	demoPatent(ctx, client, patent)
	demoPetition(ctx, client)
	demoPTAB(ctx, client)
}

func runAllWithContext(dctx *DemoContext) {
	demoPatentWithContext(dctx, dctx.Patent)
	demoPetitionWithContext(dctx)
	demoBulkWithContext(dctx)
	demoPTABWithContext(dctx)
}

func runInteractive(ctx context.Context, client *odp.Client) {
	reader := bufio.NewReader(os.Stdin)

	for {
		printHeader("USPTO ODP API Demo")
		fmt.Println("1. Patent API (13 endpoints)")
		fmt.Println("2. Petition API (3 endpoints)")
		fmt.Println("3. PTAB API (19 endpoints)")
		fmt.Println("4. Bulk Data Download")
		fmt.Println("5. Patent XML Full Text")
		fmt.Println("q. Quit")
		fmt.Print("\nSelect option: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			fmt.Print("\nEnter patent number (default: 17248024): ")
			patentInput, _ := reader.ReadString('\n')
			patentInput = strings.TrimSpace(patentInput)
			if patentInput == "" {
				patentInput = TestPatentApp
			}
			demoPatent(ctx, client, patentInput)
		case "2":
			demoPetition(ctx, client)
		case "3":
			demoPTAB(ctx, client)
		case "4":
			demoBulk(ctx, client, reader)
		case "5":
			demoXML(ctx, client, reader)
		case "q", "Q":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid option")
		}
	}
}
