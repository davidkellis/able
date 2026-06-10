package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"able/interpreter-go/internal/semanticabi/heapmodel"
)

func main() {
	outPath := flag.String("out", "", "report output path")
	check := flag.Bool("check", false, "fail if report is stale")
	flag.Parse()
	if *outPath == "" {
		fatalf("-out is required")
	}
	report, err := heapmodel.RunConformanceVectors()
	if err != nil {
		fatalf("%v", err)
	}
	generated, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fatalf("marshal report: %v", err)
	}
	generated = append(generated, '\n')
	if *check {
		current, err := os.ReadFile(*outPath)
		if err != nil {
			fatalf("read %s: %v", *outPath, err)
		}
		if !bytes.Equal(current, generated) {
			fatalf("%s is stale; regenerate the heap contract report", *outPath)
		}
		return
	}
	if err := os.WriteFile(*outPath, generated, 0o644); err != nil {
		fatalf("write %s: %v", *outPath, err)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "semanticabi-heapreport: "+format+"\n", args...)
	os.Exit(1)
}
