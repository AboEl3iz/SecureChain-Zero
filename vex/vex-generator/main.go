package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

// OpenVEX document structure (OpenVEX v0.2.0)
type VEXDocument struct {
	Context    string         `json:"@context"`
	ID         string         `json:"@id"`
	Author     string         `json:"author"`
	Timestamp  string         `json:"timestamp"`
	Version    int            `json:"version"`
	Tooling    string         `json:"tooling"`
	Statements []VEXStatement `json:"statements"`
}

type VEXStatement struct {
	Vulnerability   Vulnerability `json:"vulnerability"`
	Products        []Product     `json:"products"`
	Status          string        `json:"status"`
	Justification   string        `json:"justification,omitempty"`
	ImpactStatement string        `json:"impact_statement,omitempty"`
	ActionStatement string        `json:"action_statement,omitempty"`
}

type Vulnerability struct {
	ID          string `json:"@id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type Product struct {
	ID            string         `json:"@id"`
	Subcomponents []Subcomponent `json:"subcomponents,omitempty"`
}

type Subcomponent struct {
	ID string `json:"@id"`
}

var validJustifications = map[string]bool{
	"component_not_present":                          true,
	"vulnerable_code_not_in_execute_path":            true,
	"vulnerable_code_cannot_be_controlled_by_adversary": true,
	"inline_mitigations_already_exist":               true,
}

var validStatuses = map[string]bool{
	"not_affected":        true,
	"affected":            true,
	"fixed":               true,
	"under_investigation": true,
}

func main() {
	productFlag := flag.String("product", "", "Product identifier (e.g., pkg:oci/secure-app@sha256:abc...)")
	vulnFlag := flag.String("vuln", "", "Vulnerability identifier (e.g., CVE-2026-111497)")
	statusFlag := flag.String("status", "not_affected", "VEX status (not_affected, affected, fixed, under_investigation)")
	justificationFlag := flag.String("justification", "vulnerable_code_not_in_execute_path", "CSAF/VEX justification")
	impactFlag := flag.String("impact", "", "Impact statement describing why the artifact is immune")
	authorFlag := flag.String("author", "security-bot@myorg.com", "Author email or identifier")
	outFileFlag := flag.String("out", "", "Output file path (default stdout)")

	flag.Parse()

	if *productFlag == "" || *vulnFlag == "" {
		fmt.Fprintf(os.Stderr, "Usage: vex-generator -product <product-purl> -vuln <cve-id> [options]\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if !validStatuses[*statusFlag] {
		fmt.Fprintf(os.Stderr, "Error: Invalid status %q. Must be one of: not_affected, affected, fixed, under_investigation\n", *statusFlag)
		os.Exit(1)
	}

	if *statusFlag == "not_affected" && !validJustifications[*justificationFlag] {
		fmt.Fprintf(os.Stderr, "Error: Invalid justification %q for not_affected status.\n", *justificationFlag)
		os.Exit(1)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	doc := VEXDocument{
		Context:   "https://openvex.dev/ns/v0.2.0",
		ID:        fmt.Sprintf("https://github.com/myorg/SecureChain-Zero/vex/%d-%s", time.Now().Year(), *vulnFlag),
		Author:    *authorFlag,
		Timestamp: now,
		Version:   1,
		Tooling:   "vex-generator 1.0.0 (SecureChain-Zero)",
		Statements: []VEXStatement{
			{
				Vulnerability: Vulnerability{
					ID:   fmt.Sprintf("https://osv.dev/%s", *vulnFlag),
					Name: *vulnFlag,
				},
				Products: []Product{
					{
						ID: *productFlag,
					},
				},
				Status:          *statusFlag,
				Justification:   *justificationFlag,
				ImpactStatement: *impactFlag,
			},
		},
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}

	if *outFileFlag != "" {
		if err := os.WriteFile(*outFileFlag, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Successfully wrote OpenVEX document to %s\n", *outFileFlag)
	} else {
		fmt.Println(string(data))
	}
}
