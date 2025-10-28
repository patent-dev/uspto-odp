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
	service     = flag.String("service", "", "Service filter (patent|bulk|petition|xml)")
	endpoint    = flag.String("endpoint", "", "Specific endpoint to run")
	interactive = flag.Bool("interactive", false, "Run in interactive mode")
)

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

	if *interactive {
		runInteractive(ctx, client)
		return
	}

	switch *service {
	case "patent":
		demoPatent(ctx, client, *patent)
	case "bulk":
		reader := bufio.NewReader(os.Stdin)
		demoBulk(ctx, client, reader)
	case "petition":
		demoPetition(ctx, client)
	case "xml":
		reader := bufio.NewReader(os.Stdin)
		demoXML(ctx, client, reader)
	case "":
		runAll(ctx, client, *patent)
	default:
		log.Fatalf("Unknown service: %s", *service)
	}
}

func runAll(ctx context.Context, client *odp.Client, patent string) {
	demoPatent(ctx, client, patent)
	demoPetition(ctx, client)
}

func runInteractive(ctx context.Context, client *odp.Client) {
	reader := bufio.NewReader(os.Stdin)

	for {
		printHeader("USPTO ODP API Demo")
		fmt.Println("1. Patent API (13 endpoints)")
		fmt.Println("2. Petition API (3 endpoints)")
		fmt.Println("3. Bulk Data Download")
		fmt.Println("4. Patent XML Full Text")
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
			demoBulk(ctx, client, reader)
		case "4":
			demoXML(ctx, client, reader)
		case "q", "Q":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid option")
		}
	}
}
