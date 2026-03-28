// Command gen processes USPTO ODP swagger files and generates Go code.
//
// Usage:
//
//	go run ./cmd/gen
//
// This will:
//  1. Read original USPTO swagger/*.yaml files (kept untouched)
//  2. Bundle and apply fixes to produce swagger_fixed.yaml
//  3. Generate types to generated/types_gen.go
//  4. Generate client to generated/client_gen.go
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
	"gopkg.in/yaml.v3"
)

const (
	swaggerDir   = "swagger"
	mainFile     = "swagger/swagger.yaml"
	fixedFile    = "swagger_fixed.yaml"
	typesOutput  = "generated/types_gen.go"
	clientOutput = "generated/client_gen.go"
	packageName  = "generated"

	// Office Action DSAPI - separate package
	oaFixedFile    = "swagger_oa_fixed.yaml"
	oaTypesOutput  = "generated/oa/types_gen.go"
	oaClientOutput = "generated/oa/client_gen.go"
	oaPackageName  = "oa"

	// TSDR - separate package (different server + auth)
	tsdrSourceFile   = "swagger/tsdr-swagger.json"
	tsdrFixedFile    = "swagger_tsdr_fixed.json"
	tsdrTypesOutput  = "generated/tsdr/types_gen.go"
	tsdrClientOutput = "generated/tsdr/client_gen.go"
	tsdrPackageName  = "tsdr"
)

// Office Action spec files and their unique operationId prefixes
var oaSpecs = []struct {
	file   string
	prefix string // unique prefix for operationIds
}{
	{"oa-text-retrieval.yaml", "oa-actions"},
	{"oa-citations.yaml", "oa-citations"},
	{"oa-rejections.yaml", "oa-rejections"},
	{"oa-enriched-citations.yaml", "enriched-citations"},
}

func main() {
	log.Println("USPTO ODP Swagger Generator")
	log.Println("============================")

	// === ODP + PTAB APIs ===

	// Step 1: Bundle swagger files
	log.Println("Step 1: Bundling swagger files...")
	if err := bundleSwagger(); err != nil {
		log.Fatalf("Failed to bundle swagger: %v", err)
	}

	// Step 2: Apply fixes
	log.Println("Step 2: Applying fixes...")
	if err := applyFixes(); err != nil {
		log.Fatalf("Failed to apply fixes: %v", err)
	}

	// Step 3: Generate code
	log.Println("Step 3: Generating code...")
	if err := generateCode(); err != nil {
		log.Fatalf("Failed to generate code: %v", err)
	}

	// === Office Action DSAPI ===

	log.Println("")
	log.Println("Office Action DSAPI")
	log.Println("============================")

	// Step 4: Bundle OA specs
	log.Println("Step 4: Bundling Office Action specs...")
	if err := bundleOASwagger(); err != nil {
		log.Fatalf("Failed to bundle OA swagger: %v", err)
	}

	// Step 5: Apply OA fixes
	log.Println("Step 5: Applying OA fixes...")
	if err := applyOAFixes(); err != nil {
		log.Fatalf("Failed to apply OA fixes: %v", err)
	}

	// Step 6: Generate OA code
	log.Println("Step 6: Generating OA code...")
	if err := generateOACode(); err != nil {
		log.Fatalf("Failed to generate OA code: %v", err)
	}

	// === TSDR API ===

	log.Println("")
	log.Println("TSDR API")
	log.Println("============================")

	// Step 7: Apply TSDR fixes
	log.Println("Step 7: Applying TSDR fixes...")
	if err := applyTSDRFixes(); err != nil {
		log.Fatalf("Failed to apply TSDR fixes: %v", err)
	}

	// Step 8: Generate TSDR code
	log.Println("Step 8: Generating TSDR code...")
	if err := generateTSDRCode(); err != nil {
		log.Fatalf("Failed to generate TSDR code: %v", err)
	}

	log.Println("Done!")
}

// bundleSwagger reads all swagger YAML files and merges them into a single file.
// It resolves $ref to external files by inlining their content.
func bundleSwagger() error {
	// Load all YAML files in the swagger directory
	files := make(map[string]*yaml.Node)

	entries, err := os.ReadDir(swaggerDir)
	if err != nil {
		return fmt.Errorf("reading swagger dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			path := filepath.Join(swaggerDir, entry.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("reading %s: %w", path, err)
			}

			var doc yaml.Node
			if err := yaml.Unmarshal(data, &doc); err != nil {
				return fmt.Errorf("parsing %s: %w", path, err)
			}
			files[entry.Name()] = &doc
			log.Printf("  - Loaded %s", entry.Name())
		}
	}

	// Load main swagger file
	mainDoc, ok := files["swagger.yaml"]
	if !ok {
		return fmt.Errorf("swagger.yaml not found")
	}

	// Merge components from all other files into main doc
	for filename, doc := range files {
		if filename == "swagger.yaml" {
			continue
		}
		if err := mergeComponents(mainDoc, doc, filename); err != nil {
			return fmt.Errorf("merging %s: %w", filename, err)
		}
		log.Printf("  - Merged components from %s", filename)
	}

	// Resolve all external $refs by converting to internal refs
	if err := resolveExternalRefs(mainDoc, files); err != nil {
		return fmt.Errorf("resolving refs: %w", err)
	}

	// Write bundled file
	out, err := yaml.Marshal(mainDoc)
	if err != nil {
		return fmt.Errorf("marshaling bundled spec: %w", err)
	}

	if err := os.WriteFile(fixedFile, out, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", fixedFile, err)
	}

	log.Printf("  - Written %s", fixedFile)
	return nil
}

// resolveExternalRefs recursively resolves $ref to external files
func resolveExternalRefs(node *yaml.Node, files map[string]*yaml.Node) error {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			if err := resolveExternalRefs(child, files); err != nil {
				return err
			}
		}
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			if keyNode.Value == "$ref" && valueNode.Kind == yaml.ScalarNode {
				ref := valueNode.Value
				// Check if it's an external ref (starts with ./)
				if strings.HasPrefix(ref, "./") {
					// Parse: ./filename.yaml#/path/to/schema
					parts := strings.SplitN(ref, "#", 2)
					if len(parts) != 2 {
						continue
					}
					jsonPath := parts[1]

					// Convert to internal ref
					valueNode.Value = "#" + jsonPath
				}
			} else {
				if err := resolveExternalRefs(valueNode, files); err != nil {
					return err
				}
			}
		}
	case yaml.SequenceNode:
		for _, child := range node.Content {
			if err := resolveExternalRefs(child, files); err != nil {
				return err
			}
		}
	}

	return nil
}

// mergeComponents merges components and paths from source file into the main document
func mergeComponents(mainDoc *yaml.Node, sourceFile *yaml.Node, _ string) error {
	mainRoot := getDocRoot(mainDoc)
	sourceRoot := getDocRoot(sourceFile)

	// Merge components
	mainComponents := findChildNode(mainRoot, "components")
	sourceComponents := findChildNode(sourceRoot, "components")

	if sourceComponents != nil {
		if mainComponents == nil {
			mainRoot.Content = append(mainRoot.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "components"},
				&yaml.Node{Kind: yaml.MappingNode},
			)
			mainComponents = mainRoot.Content[len(mainRoot.Content)-1]
		}

		for i := 0; i < len(sourceComponents.Content); i += 2 {
			categoryKey := sourceComponents.Content[i].Value
			categoryValue := sourceComponents.Content[i+1]

			mainCategory := findChildNode(mainComponents, categoryKey)
			if mainCategory == nil {
				mainComponents.Content = append(mainComponents.Content,
					&yaml.Node{Kind: yaml.ScalarNode, Value: categoryKey},
					&yaml.Node{Kind: yaml.MappingNode},
				)
				mainCategory = mainComponents.Content[len(mainComponents.Content)-1]
			}

			for j := 0; j < len(categoryValue.Content); j += 2 {
				itemKey := categoryValue.Content[j]
				itemValue := categoryValue.Content[j+1]

				existingItem := findChildNodeIndex(mainCategory, itemKey.Value)
				if existingItem < 0 {
					// Add new item
					mainCategory.Content = append(mainCategory.Content, itemKey, itemValue)
				} else {
					// Replace if existing is just a $ref and source has actual content
					existingValue := mainCategory.Content[existingItem+1]
					if isJustRef(existingValue) && !isJustRef(itemValue) {
						mainCategory.Content[existingItem+1] = itemValue
					}
				}
			}
		}
	}

	// Merge paths from source files into main document
	mainPaths := findChildNode(mainRoot, "paths")
	sourcePaths := findChildNode(sourceRoot, "paths")

	if sourcePaths != nil && mainPaths != nil {
		// For each path in main that's just a $ref, replace it with actual content
		for i := 0; i < len(mainPaths.Content); i += 2 {
			pathValue := mainPaths.Content[i+1]

			if isJustRef(pathValue) {
				// Get the ref target path
				refNode := findChildNode(pathValue, "$ref")
				if refNode != nil && strings.HasPrefix(refNode.Value, "./") {
					// Parse: ./filename.yaml#/paths/~1encoded~1path
					parts := strings.SplitN(refNode.Value, "#", 2)
					if len(parts) == 2 {
						jsonPath := parts[1]
						// Path refs look like: /paths/~1api~1v1~1patent~1proceedings~1search
						if strings.HasPrefix(jsonPath, "/paths/") {
							encodedPath := strings.TrimPrefix(jsonPath, "/paths/")
							// URL decode: ~1 -> /
							decodedPath := strings.ReplaceAll(encodedPath, "~1", "/")
							decodedPath = strings.ReplaceAll(decodedPath, "~0", "~")

							// Find this path in source
							targetPath := findChildNode(sourcePaths, decodedPath)
							if targetPath != nil {
								// Replace the ref with actual content
								mainPaths.Content[i+1] = targetPath
							}
						}
					}
				}
			}
		}

	}

	return nil
}

func applyFixes() error {
	data, err := os.ReadFile(fixedFile)
	if err != nil {
		return fmt.Errorf("reading %s: %w", fixedFile, err)
	}

	content := string(data)
	fixCount := 0

	// Fix 1: frameNumber: type: string -> type: integer
	re1 := regexp.MustCompile(`(frameNumber:\s*\n\s+type:)\s*string`)
	if re1.MatchString(content) {
		content = re1.ReplaceAllString(content, "${1} integer")
		fixCount++
		log.Println("  - Fixed frameNumber: string -> integer")
	}

	// Fix 2: reelNumber: type: string -> type: integer
	re2 := regexp.MustCompile(`(reelNumber:\s*\n\s+type:)\s*string`)
	if re2.MatchString(content) {
		content = re2.ReplaceAllString(content, "${1} integer")
		fixCount++
		log.Println("  - Fixed reelNumber: string -> integer")
	}

	// Fix 3: petitionIssueConsideredTextBag items: type: object -> type: string
	re3 := regexp.MustCompile(`(petitionIssueConsideredTextBag:\s*\n\s+type:\s*array\s*\n\s+items:\s*\n\s+type:)\s*object`)
	if re3.MatchString(content) {
		content = re3.ReplaceAllString(content, "${1} string")
		fixCount++
		log.Println("  - Fixed petitionIssueConsideredTextBag items: object -> string")
	}

	// Fix 4: Remove the empty /api/v1/patent/applications/text-to-search endpoint
	re4 := regexp.MustCompile(`(?m)^\s*/api/v1/patent/applications/text-to-search:\s*\n\s+x-ip-domain:[^\n]*\n\s+x-service-type:[^\n]*\n\s+x-content-type:[^\n]*\n`)
	if re4.MatchString(content) {
		content = re4.ReplaceAllString(content, "")
		fixCount++
		log.Println("  - Removed empty text-to-search endpoint")
	}

	// Fix 5: Remove format: date from fields that return non-ISO datetime strings
	// The API returns "2025-09-23 00:57:53" format which isn't ISO date or RFC3339
	re5 := regexp.MustCompile(`(fileReleaseDate:\s*\n\s+type:\s*string)\s*\n\s+format:\s*date\b`)
	if re5.MatchString(content) {
		content = re5.ReplaceAllString(content, "${1}")
		fixCount++
		log.Println("  - Fixed fileReleaseDate: removed format (non-ISO datetime)")
	}

	re6 := regexp.MustCompile(`(fileLastModifiedDateTime:\s*\n\s+type:\s*string)\s*\n\s+format:\s*date\b`)
	if re6.MatchString(content) {
		content = re6.ReplaceAllString(content, "${1}")
		fixCount++
		log.Println("  - Fixed fileLastModifiedDateTime: removed format (non-ISO datetime)")
	}

	// Fix 7: Remove format: date-time from fields that return non-RFC3339 formats
	// The PTAB API returns "2025-11-26T23:58:00" without timezone which fails Go's time parsing
	datetimeFields := []string{
		"documentFilingDate",
		"documentFilingDateTime",
		"decisionIssueDate",
		"lastModifiedDateTime",
		"createdDateTime",
	}
	for _, field := range datetimeFields {
		re := regexp.MustCompile(fmt.Sprintf(`(%s:\s*\n\s+type:\s*string)\s*\n\s+format:\s*date-time\b`, field))
		if re.MatchString(content) {
			content = re.ReplaceAllString(content, "${1}")
			fixCount++
			log.Printf("  - Fixed %s: removed format: date-time (non-RFC3339)", field)
		}
	}

	// Fix 7b: Remove format: date from DateTime fields (they return datetime but spec says date)
	// e.g., appealLastModifiedDateTime has format: date but returns "2025-11-20T18:21:10"
	dateTimeFieldsWithDateFormat := []string{
		"appealLastModifiedDateTime",
	}
	for _, field := range dateTimeFieldsWithDateFormat {
		re := regexp.MustCompile(fmt.Sprintf(`(%s:\s*\n\s+type:\s*string)\s*\n\s+format:\s*date\b`, field))
		if re.MatchString(content) {
			content = re.ReplaceAllString(content, "${1}")
			fixCount++
			log.Printf("  - Fixed %s: removed format: date (returns datetime)", field)
		}
	}

	// Fix 7c: documentNumber - API returns number but swagger says string
	re7c := regexp.MustCompile(`(documentNumber:\s*\n\s+type:)\s*string`)
	if re7c.MatchString(content) {
		content = re7c.ReplaceAllString(content, "${1} integer")
		fixCount++
		log.Println("  - Fixed documentNumber: string -> integer (API returns number)")
	}

	// Fix 8: correspondenceAddress in Assignment - API returns object, swagger says array
	// Change from array to oneOf array/object to handle both cases
	re8 := regexp.MustCompile(`(correspondenceAddress:\s*\n\s+type:)\s*array`)
	if re8.MatchString(content) {
		// Just remove the type constraint entirely, let it be dynamic
		content = re8.ReplaceAllString(content, "${1} object")
		fixCount++
		log.Println("  - Fixed correspondenceAddress: array -> object (API returns object)")
	}

	// Fix 9: DecisionData fields - API returns arrays but swagger says string
	// statuteAndRuleBag and issueTypeBag are arrays of strings
	decisionDataStringToArrayFields := []string{
		"statuteAndRuleBag",
		"issueTypeBag",
	}
	for _, field := range decisionDataStringToArrayFields {
		// Match the field in DecisionData schema where it's defined as type: string
		re := regexp.MustCompile(fmt.Sprintf(`(%s:\s*\n\s+type:)\s*string(\s*\n\s+example:)`, field))
		if re.MatchString(content) {
			content = re.ReplaceAllString(content, "${1} array\n                    items:\n                        type: string${2}")
			fixCount++
			log.Printf("  - Fixed DecisionData.%s: string -> array of strings (API returns array)", field)
		}
	}

	// Fix 10: Error response code field - API returns string but swagger says integer
	// e.g., NotFound returns {"code": "404", ...} not {"code": 404, ...}
	re10 := regexp.MustCompile(`(code:\s*\n\s+type:)\s*integer(\s*\n\s+example:\s*\d+)`)
	if re10.MatchString(content) {
		content = re10.ReplaceAllString(content, "${1} string${2}")
		fixCount++
		log.Println("  - Fixed error response code: integer -> string (API returns string)")
	}

	// Fix 11: InterferenceDecisionRecord.decisionDocumentData -> documentData
	// The swagger says "decisionDocumentData" but the API returns "documentData"
	re11 := regexp.MustCompile(`(\s+)decisionDocumentData:(\s*\n\s+\$ref:\s*['"]#/components/schemas/InterferenceDecisionDocumentData['"])`)
	if re11.MatchString(content) {
		content = re11.ReplaceAllString(content, "${1}documentData:${2}")
		fixCount++
		log.Println("  - Fixed InterferenceDecisionRecord: decisionDocumentData -> documentData (API field name)")
	}

	// Fix 12: GetPatentAssignment inline response has assignmentBag as single object, but API returns array
	// Match assignmentBag directly followed by $ref to Assignment (not already an array type)
	re12 := regexp.MustCompile(`(assignmentBag:\s*\n\s+)(\$ref:\s*['"]#/components/schemas/Assignment['"])`)
	if re12.MatchString(content) {
		content = re12.ReplaceAllString(content, "${1}type: array\n                                                    items:\n                                                        ${2}")
		fixCount++
		log.Println("  - Fixed GetPatentAssignment: assignmentBag single object -> array (API returns array)")
	}

	log.Printf("  Applied %d regex fixes", fixCount)

	if err := os.WriteFile(fixedFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", fixedFile, err)
	}

	// Fix 5: Move response-like schemas to components/responses
	if err := fixResponseSchemas(); err != nil {
		return fmt.Errorf("fixing response schemas: %w", err)
	}

	return nil
}

// fixResponseSchemas handles response-like schema definitions.
// The USPTO swagger incorrectly defines error responses as schemas with response structure.
// These are used both as responses (in response context) and as schemas (inline).
// We split them: move to responses for response usage, create schema for schema usage.
func fixResponseSchemas() error {
	data, err := os.ReadFile(fixedFile)
	if err != nil {
		return err
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("parsing YAML: %w", err)
	}

	responseNames := []string{"BadRequest", "Forbidden", "NotFound", "Status413", "InternalError"}
	fixCount := 0

	components := findNode(&doc, "components")
	if components == nil {
		return fmt.Errorf("components not found")
	}

	schemas := findChildNode(components, "schemas")
	if schemas == nil {
		return fmt.Errorf("components.schemas not found")
	}

	// Find or create components.responses
	responses := findChildNode(components, "responses")
	if responses == nil {
		responses = &yaml.Node{Kind: yaml.MappingNode}
		components.Content = append(components.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "responses"},
			responses,
		)
	}

	// Process response-like schemas and create response wrappers for schemas used in response context
	newSchemasContent := make([]*yaml.Node, 0)
	for i := 0; i < len(schemas.Content); i += 2 {
		keyNode := schemas.Content[i]
		valueNode := schemas.Content[i+1]

		isResponseSchema := false
		for _, name := range responseNames {
			if keyNode.Value == name {
				isResponseSchema = true
				break
			}
		}

		if isResponseSchema {
			if findChildNode(valueNode, "content") != nil {
				// Already has response structure - move to responses, keep schema version
				responses.Content = append(responses.Content,
					&yaml.Node{Kind: yaml.ScalarNode, Value: keyNode.Value},
					valueNode,
				)

				// Extract inner schema for schema refs
				content := findChildNode(valueNode, "content")
				if content != nil {
					appJson := findChildNode(content, "application/json")
					if appJson != nil {
						innerSchema := findChildNode(appJson, "schema")
						if innerSchema != nil {
							newSchemasContent = append(newSchemasContent,
								&yaml.Node{Kind: yaml.ScalarNode, Value: keyNode.Value + "Schema"},
								innerSchema,
							)
						}
					}
				}

				fixCount++
				log.Printf("  - Split %s into response and schema", keyNode.Value)
			} else {
				// Is a schema - keep it, and create a response wrapper
				newSchemasContent = append(newSchemasContent, keyNode, valueNode)

				// Create response that wraps this schema
				responseNode := createResponseWrapper(keyNode.Value, valueNode)
				responses.Content = append(responses.Content,
					&yaml.Node{Kind: yaml.ScalarNode, Value: keyNode.Value},
					responseNode,
				)

				fixCount++
				log.Printf("  - Created response wrapper for schema %s", keyNode.Value)
			}
		} else {
			newSchemasContent = append(newSchemasContent, keyNode, valueNode)
		}
	}
	schemas.Content = newSchemasContent

	// Track which schemas were split (have XxxSchema version) vs wrapped (keep original name)
	splitNames := make([]string, 0)   // These have XxxSchema in schemas section
	wrappedNames := make([]string, 0) // These keep original name in schemas section

	// Check which response names have corresponding Schema versions
	for _, name := range responseNames {
		if findChildNode(schemas, name+"Schema") != nil {
			splitNames = append(splitNames, name)
		} else if findChildNode(schemas, name) != nil {
			wrappedNames = append(wrappedNames, name)
		}
	}

	// Update refs based on context:
	// - Refs to split schemas in schema context -> XxxSchema (only for actually split ones)
	// - All refs in response context -> responses/Xxx
	fixSchemaRefs(&doc, splitNames)

	// For response context, update both split and wrapped names
	allResponseNames := make([]string, 0, len(splitNames)+len(wrappedNames))
	allResponseNames = append(allResponseNames, splitNames...)
	allResponseNames = append(allResponseNames, wrappedNames...)
	fixResponseRefs(&doc, allResponseNames)

	if fixCount > 0 {
		out, err := yaml.Marshal(&doc)
		if err != nil {
			return fmt.Errorf("marshaling YAML: %w", err)
		}
		if err := os.WriteFile(fixedFile, out, 0644); err != nil {
			return fmt.Errorf("writing file: %w", err)
		}
		log.Printf("  Applied %d response/schema splits", fixCount)
	}

	return nil
}

// fixSchemaRefs updates refs to error schemas that appear in schema context
func fixSchemaRefs(node *yaml.Node, names []string) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			// If this is a "schema" key, check if its value contains a ref to one of our schemas
			if keyNode.Value == "schema" {
				updateSchemaContextRefs(valueNode, names)
			} else {
				fixSchemaRefs(valueNode, names)
			}
		}
	case yaml.SequenceNode, yaml.DocumentNode:
		for _, child := range node.Content {
			fixSchemaRefs(child, names)
		}
	}
}

// updateSchemaContextRefs updates refs within a schema context to use XxxSchema
func updateSchemaContextRefs(node *yaml.Node, names []string) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			if keyNode.Value == "$ref" && valueNode.Kind == yaml.ScalarNode {
				for _, name := range names {
					oldRef := "#/components/schemas/" + name
					newRef := "#/components/schemas/" + name + "Schema"
					if valueNode.Value == oldRef {
						valueNode.Value = newRef
					}
				}
			} else {
				updateSchemaContextRefs(valueNode, names)
			}
		}
	case yaml.SequenceNode:
		for _, child := range node.Content {
			updateSchemaContextRefs(child, names)
		}
	}
}

// fixResponseRefs updates refs in response context (like '400': {$ref: ...}) to point to responses
func fixResponseRefs(node *yaml.Node, names []string) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			// Check if this is a responses section or an HTTP status code
			if keyNode.Value == "responses" || isHTTPStatusCode(keyNode.Value) {
				updateResponseContextRefs(valueNode, names)
			} else {
				fixResponseRefs(valueNode, names)
			}
		}
	case yaml.SequenceNode, yaml.DocumentNode:
		for _, child := range node.Content {
			fixResponseRefs(child, names)
		}
	}
}

// updateResponseContextRefs updates refs in response context to use responses/Xxx
func updateResponseContextRefs(node *yaml.Node, names []string) {
	if node == nil {
		return
	}

	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			if keyNode.Value == "$ref" && valueNode.Kind == yaml.ScalarNode {
				for _, name := range names {
					oldRef := "#/components/schemas/" + name
					newRef := "#/components/responses/" + name
					if valueNode.Value == oldRef {
						valueNode.Value = newRef
					}
				}
			} else if isHTTPStatusCode(keyNode.Value) {
				// Recurse into status code entries
				updateResponseContextRefs(valueNode, names)
			}
		}
	}
}

func isHTTPStatusCode(s string) bool {
	if len(s) != 3 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// createResponseWrapper creates an OpenAPI response object that wraps a schema
func createResponseWrapper(name string, schema *yaml.Node) *yaml.Node {
	// Create: description: "Error", content: application/json: schema: $ref: #/components/schemas/Name
	return &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "description"},
			{Kind: yaml.ScalarNode, Value: name},
			{Kind: yaml.ScalarNode, Value: "content"},
			{Kind: yaml.MappingNode, Content: []*yaml.Node{
				{Kind: yaml.ScalarNode, Value: "application/json"},
				{Kind: yaml.MappingNode, Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "schema"},
					{Kind: yaml.MappingNode, Content: []*yaml.Node{
						{Kind: yaml.ScalarNode, Value: "$ref"},
						{Kind: yaml.ScalarNode, Value: "#/components/schemas/" + name},
					}},
				}},
			}},
		},
	}
}

func findNode(root *yaml.Node, keys ...string) *yaml.Node {
	current := root
	if current.Kind == yaml.DocumentNode && len(current.Content) > 0 {
		current = current.Content[0]
	}

	for _, key := range keys {
		current = findChildNode(current, key)
		if current == nil {
			return nil
		}
	}
	return current
}

func findChildNode(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func findChildNodeIndex(node *yaml.Node, key string) int {
	if node == nil || node.Kind != yaml.MappingNode {
		return -1
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return i
		}
	}
	return -1
}

// isJustRef returns true if the node is only a $ref (no other content)
func isJustRef(node *yaml.Node) bool {
	if node.Kind != yaml.MappingNode {
		return false
	}
	if len(node.Content) != 2 {
		return false
	}
	return node.Content[0].Value == "$ref"
}

func getDocRoot(doc *yaml.Node) *yaml.Node {
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		return doc.Content[0]
	}
	return doc
}

func generateCode() error {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	spec, err := loader.LoadFromFile(fixedFile)
	if err != nil {
		return fmt.Errorf("loading spec: %w", err)
	}

	// Generate types
	log.Println("  Generating types...")
	typesConfig := codegen.Configuration{
		PackageName: packageName,
		Generate: codegen.GenerateOptions{
			Models: true,
		},
	}

	typesCode, err := codegen.Generate(spec, typesConfig)
	if err != nil {
		return fmt.Errorf("generating types: %w", err)
	}

	if err := os.WriteFile(typesOutput, []byte(typesCode), 0644); err != nil {
		return fmt.Errorf("writing types: %w", err)
	}
	log.Printf("  - Written %s (%d bytes)", typesOutput, len(typesCode))

	// Generate client
	log.Println("  Generating client...")
	clientConfig := codegen.Configuration{
		PackageName: packageName,
		Generate: codegen.GenerateOptions{
			Client: true,
		},
	}

	clientCode, err := codegen.Generate(spec, clientConfig)
	if err != nil {
		return fmt.Errorf("generating client: %w", err)
	}

	if err := os.WriteFile(clientOutput, []byte(clientCode), 0644); err != nil {
		return fmt.Errorf("writing client: %w", err)
	}
	log.Printf("  - Written %s (%d bytes)", clientOutput, len(clientCode))

	return nil
}

// bundleOASwagger merges the 4 Office Action DSAPI specs into one OpenAPI file.
// The original specs are kept untouched in swagger/. This only combines their paths.
func bundleOASwagger() error {
	// Start with the first spec as the base
	baseData, err := os.ReadFile(filepath.Join(swaggerDir, oaSpecs[0].file))
	if err != nil {
		return fmt.Errorf("reading %s: %w", oaSpecs[0].file, err)
	}

	var baseDoc yaml.Node
	if err := yaml.Unmarshal(baseData, &baseDoc); err != nil {
		return fmt.Errorf("parsing %s: %w", oaSpecs[0].file, err)
	}
	log.Printf("  - Loaded %s (base)", oaSpecs[0].file)

	baseRoot := getDocRoot(&baseDoc)
	basePaths := findChildNode(baseRoot, "paths")

	// Update the title to reflect the combined spec
	info := findChildNode(baseRoot, "info")
	if info != nil {
		title := findChildNode(info, "title")
		if title != nil {
			title.Value = "USPTO Office Action APIs (DSAPI)"
		}
		desc := findChildNode(info, "description")
		if desc != nil {
			desc.Value = "Combined Office Action APIs: Text Retrieval, Citations, Rejections, and Enriched Citations. " +
				"These use the DSAPI pattern (form-encoded POST with Lucene/Solr queries)."
		}
	}

	// Update tags to include all APIs
	tags := findChildNode(baseRoot, "tags")
	if tags != nil {
		tags.Content = nil // clear
	}

	// Track existing paths to detect collisions
	existingPaths := make(map[string]string) // path -> source spec
	if basePaths != nil {
		for j := 0; j+1 < len(basePaths.Content); j += 2 {
			existingPaths[basePaths.Content[j].Value] = oaSpecs[0].file
		}
	}

	// Merge paths and components from remaining specs
	baseComponents := findChildNode(baseRoot, "components")
	baseSchemas := findChildNode(baseComponents, "schemas")

	for i := 1; i < len(oaSpecs); i++ {
		specData, err := os.ReadFile(filepath.Join(swaggerDir, oaSpecs[i].file))
		if err != nil {
			return fmt.Errorf("reading %s: %w", oaSpecs[i].file, err)
		}

		var specDoc yaml.Node
		if err := yaml.Unmarshal(specData, &specDoc); err != nil {
			return fmt.Errorf("parsing %s: %w", oaSpecs[i].file, err)
		}

		specRoot := getDocRoot(&specDoc)
		specPaths := findChildNode(specRoot, "paths")

		if specPaths != nil && basePaths != nil {
			for j := 0; j+1 < len(specPaths.Content); j += 2 {
				pathKey := specPaths.Content[j].Value
				if existingSpec, exists := existingPaths[pathKey]; exists {
					return fmt.Errorf("duplicate path %q: defined in both %s and %s", pathKey, existingSpec, oaSpecs[i].file)
				}
				existingPaths[pathKey] = oaSpecs[i].file
				basePaths.Content = append(basePaths.Content,
					specPaths.Content[j],
					specPaths.Content[j+1],
				)
			}
		}

		// Merge components/schemas if present
		specComponents := findChildNode(specRoot, "components")
		specSchemas := findChildNode(specComponents, "schemas")
		if specSchemas != nil && baseSchemas != nil {
			for j := 0; j+1 < len(specSchemas.Content); j += 2 {
				schemaName := specSchemas.Content[j].Value
				if findChildNode(baseSchemas, schemaName) == nil {
					baseSchemas.Content = append(baseSchemas.Content,
						specSchemas.Content[j],
						specSchemas.Content[j+1],
					)
				}
			}
		}

		log.Printf("  - Merged paths from %s", oaSpecs[i].file)
	}

	out, err := yaml.Marshal(&baseDoc)
	if err != nil {
		return fmt.Errorf("marshaling bundled OA spec: %w", err)
	}

	if err := os.WriteFile(oaFixedFile, out, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", oaFixedFile, err)
	}

	log.Printf("  - Written %s", oaFixedFile)
	return nil
}

// applyOAFixes applies bug fixes to the bundled Office Action spec via YAML tree manipulation.
// Each fix corrects a documented issue in the original USPTO swagger specs.
func applyOAFixes() error {
	data, err := os.ReadFile(oaFixedFile)
	if err != nil {
		return fmt.Errorf("reading %s: %w", oaFixedFile, err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("parsing YAML: %w", err)
	}

	root := getDocRoot(&doc)
	paths := findChildNode(root, "paths")
	if paths == nil {
		return fmt.Errorf("no paths found in OA spec")
	}

	// Build a map from path string to operationId prefix
	pathPrefixMap := make(map[string]string)
	for _, spec := range oaSpecs {
		specData, err := os.ReadFile(filepath.Join(swaggerDir, spec.file))
		if err != nil {
			return fmt.Errorf("reading %s for prefix map: %w", spec.file, err)
		}
		var specDoc yaml.Node
		if err := yaml.Unmarshal(specData, &specDoc); err != nil {
			return fmt.Errorf("parsing %s for prefix map: %w", spec.file, err)
		}
		specPaths := findChildNode(getDocRoot(&specDoc), "paths")
		if specPaths == nil {
			log.Printf("  Warning: no paths found in %s", spec.file)
			continue
		}
		for j := 0; j+1 < len(specPaths.Content); j += 2 {
			pathPrefixMap[specPaths.Content[j].Value] = spec.prefix
		}
	}

	fixCount := 0

	// Walk all paths and apply fixes
	for i := 0; i+1 < len(paths.Content); i += 2 {
		pathStr := paths.Content[i].Value
		pathNode := paths.Content[i+1]
		prefix := pathPrefixMap[pathStr]

		if pathNode.Kind != yaml.MappingNode {
			continue
		}

		// Walk each method (get, post, etc.) under this path
		for j := 0; j+1 < len(pathNode.Content); j += 2 {
			methodNode := pathNode.Content[j+1]
			if methodNode.Kind != yaml.MappingNode {
				continue
			}

			// Bug 1: Remove phantom path parameters (dataset, version declared as "in: path")
			params := findChildNode(methodNode, "parameters")
			if params != nil && params.Kind == yaml.SequenceNode {
				filtered := make([]*yaml.Node, 0)
				for _, param := range params.Content {
					nameNode := findChildNode(param, "name")
					inNode := findChildNode(param, "in")
					if nameNode != nil && inNode != nil && inNode.Value == "path" &&
						(nameNode.Value == "dataset" || nameNode.Value == "version") {
						fixCount++
						continue // skip phantom param
					}
					filtered = append(filtered, param)
				}
				params.Content = filtered
			}

			// Bug 2: Make operationIds unique (all specs use list-searchable-fields / perform-search)
			if prefix != "" {
				opIdIdx := findChildNodeIndex(methodNode, "operationId")
				if opIdIdx >= 0 {
					opIdNode := methodNode.Content[opIdIdx+1]
					switch opIdNode.Value {
					case "list-searchable-fields":
						opIdNode.Value = prefix + "-list-fields"
						fixCount++
					case "perform-search":
						opIdNode.Value = prefix + "-search"
						fixCount++
					}
				}
			}
		}
	}

	log.Printf("  - Bug fix: removed phantom path parameters + made operationIds unique")
	log.Printf("  Applied %d OA fixes", fixCount)

	out, err := yaml.Marshal(&doc)
	if err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}

	if err := os.WriteFile(oaFixedFile, out, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", oaFixedFile, err)
	}

	return nil
}

// generateOACode generates Go types and client from the bundled OA spec.
func generateOACode() error {
	// Ensure output directory exists
	if err := os.MkdirAll("generated/oa", 0755); err != nil {
		return fmt.Errorf("creating generated/oa: %w", err)
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	spec, err := loader.LoadFromFile(oaFixedFile)
	if err != nil {
		return fmt.Errorf("loading spec: %w", err)
	}

	// Generate types
	log.Println("  Generating OA types...")
	typesConfig := codegen.Configuration{
		PackageName: oaPackageName,
		Generate: codegen.GenerateOptions{
			Models: true,
		},
	}

	typesCode, err := codegen.Generate(spec, typesConfig)
	if err != nil {
		return fmt.Errorf("generating OA types: %w", err)
	}

	if err := os.WriteFile(oaTypesOutput, []byte(typesCode), 0644); err != nil {
		return fmt.Errorf("writing OA types: %w", err)
	}
	log.Printf("  - Written %s (%d bytes)", oaTypesOutput, len(typesCode))

	// Generate client
	log.Println("  Generating OA client...")
	clientConfig := codegen.Configuration{
		PackageName: oaPackageName,
		Generate: codegen.GenerateOptions{
			Client: true,
		},
	}

	clientCode, err := codegen.Generate(spec, clientConfig)
	if err != nil {
		return fmt.Errorf("generating OA client: %w", err)
	}

	if err := os.WriteFile(oaClientOutput, []byte(clientCode), 0644); err != nil {
		return fmt.Errorf("writing OA client: %w", err)
	}
	log.Printf("  - Written %s (%d bytes)", oaClientOutput, len(clientCode))

	return nil
}

// applyTSDRFixes copies the TSDR spec and applies bug fixes.
func applyTSDRFixes() error {
	data, err := os.ReadFile(tsdrSourceFile)
	if err != nil {
		return fmt.Errorf("reading %s: %w", tsdrSourceFile, err)
	}

	content := string(data)
	fixCount := 0

	// Bug 1: Server URL is protocol-relative "//tsdrapi.uspto.gov/"
	// oapi-codegen needs a full URL
	if strings.Contains(content, `"//tsdrapi.uspto.gov/"`) {
		content = strings.ReplaceAll(content, `"//tsdrapi.uspto.gov/"`, `"https://tsdrapi.uspto.gov"`)
		fixCount++
		log.Println("  - Bug fix: server URL protocol-relative -> https")
	}

	// Bug 2: date and date-time format fields cause time.Time parsing failures.
	// The API returns inconsistent formats: date-only ("2021-11-19") in fields
	// declared as date-time, and non-ISO formats elsewhere. Strip all date/date-time
	// formats so they generate as plain strings.
	for _, dateFmt := range []string{`"date-time"`, `"date"`} {
		old := `"format": ` + dateFmt
		if strings.Contains(content, old) {
			count := strings.Count(content, old)
			content = strings.ReplaceAll(content, old, `"x-original-format": `+dateFmt)
			fixCount++
			log.Printf("  - Bug fix: removed %d date format declarations (%s)", count, dateFmt)
		}
	}

	log.Printf("  Applied %d TSDR fixes", fixCount)

	if err := os.WriteFile(tsdrFixedFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", tsdrFixedFile, err)
	}
	log.Printf("  - Written %s", tsdrFixedFile)

	return nil
}

// generateTSDRCode generates Go types and client from the TSDR spec.
func generateTSDRCode() error {
	if err := os.MkdirAll("generated/tsdr", 0755); err != nil {
		return fmt.Errorf("creating generated/tsdr: %w", err)
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	spec, err := loader.LoadFromFile(tsdrFixedFile)
	if err != nil {
		return fmt.Errorf("loading TSDR spec: %w", err)
	}

	// Generate types
	log.Println("  Generating TSDR types...")
	typesConfig := codegen.Configuration{
		PackageName: tsdrPackageName,
		Generate: codegen.GenerateOptions{
			Models: true,
		},
	}

	typesCode, err := codegen.Generate(spec, typesConfig)
	if err != nil {
		return fmt.Errorf("generating TSDR types: %w", err)
	}

	if err := os.WriteFile(tsdrTypesOutput, []byte(typesCode), 0644); err != nil {
		return fmt.Errorf("writing TSDR types: %w", err)
	}
	log.Printf("  - Written %s (%d bytes)", tsdrTypesOutput, len(typesCode))

	// Generate client
	log.Println("  Generating TSDR client...")
	clientConfig := codegen.Configuration{
		PackageName: tsdrPackageName,
		Generate: codegen.GenerateOptions{
			Client: true,
		},
	}

	clientCode, err := codegen.Generate(spec, clientConfig)
	if err != nil {
		return fmt.Errorf("generating TSDR client: %w", err)
	}

	if err := os.WriteFile(tsdrClientOutput, []byte(clientCode), 0644); err != nil {
		return fmt.Errorf("writing TSDR client: %w", err)
	}
	log.Printf("  - Written %s (%d bytes)", tsdrClientOutput, len(clientCode))

	return nil
}
