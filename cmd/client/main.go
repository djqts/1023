package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"banner-fingerprint/pkg/api"
	"banner-fingerprint/pkg/client"
)

func main() {
	os.Exit(run())
}

func run() int {
	fs := flag.NewFlagSet("client", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	defaultServer := os.Getenv("SERVER_URL")
	if defaultServer == "" {
		defaultServer = "http://127.0.0.1:8080"
	}
	server := fs.String("server", defaultServer, "server base URL")
	input := fs.String("input", "data/input.json", "local JSON file with scan items")
	timeout := fs.Duration("timeout", 30*time.Second, "request timeout")
	jsonOut := fs.Bool("json", false, "print results as JSON instead of a table")
	batchSize := fs.Int("batch", 500, "items per HTTP request")
	retries := fs.Int("retries", 3, "retry count for 5xx/429/timeout")
	retryBackoff := fs.Duration("retry-backoff", 200*time.Millisecond, "initial retry backoff")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return 2
	}
	if fs.NArg() > 0 {
		fmt.Fprintln(os.Stderr, "unexpected arguments:", fs.Args())
		return 2
	}

	raw, err := os.ReadFile(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read input %s: %v\n", *input, err)
		return 2
	}
	items, err := api.DecodeBatch(raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse input %s: %v\n", *input, err)
		return 2
	}

	cli := client.NewHTTP(client.Config{
		BaseURL:      *server,
		Timeout:      *timeout,
		BatchSize:    *batchSize,
		MaxRetries:   *retries,
		RetryBackoff: *retryBackoff,
	})

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	resp, err := cli.Fingerprint(ctx, api.BatchRequest{Items: items})
	partial := false
	if err != nil {
		var partialErr *client.PartialError
		if errors.As(err, &partialErr) && len(resp.Results) > 0 {
			partial = true
			fmt.Fprintf(os.Stderr, "partial failure: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "fingerprint request failed: %v\n", err)
			return 1
		}
	}

	if code := writeResults(resp.Results, *jsonOut); code != 0 {
		return code
	}
	if partial {
		return 3
	}
	return 0
}

func writeResults(results []api.FingerprintResult, jsonOut bool) int {
	if jsonOut {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(results); err != nil {
			fmt.Fprintf(os.Stderr, "encode output: %v\n", err)
			return 1
		}
		return 0
	}
	printTable(results)
	return 0
}

func printTable(results []api.FingerprintResult) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "IP\tPORT\tPROTOCOL\tPRODUCT\tVERSION\tOS_HINT\tCONF")
	for _, result := range results {
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\t%s\t%.2f\n",
			display(result.IP),
			result.Port,
			display(result.Protocol),
			display(result.Product),
			display(result.Version),
			display(result.OSHint),
			result.Confidence,
		)
	}
	_ = w.Flush()

	unknown := 0
	for _, result := range results {
		if result.Protocol == api.ProtocolUnknown {
			unknown++
		}
	}
	fmt.Printf("\n%d items, %d unknown, %d recognized\n", len(results), unknown, len(results)-unknown)
}

func display(value string) string {
	if value == "" {
		return "-"
	}
	return value
}
