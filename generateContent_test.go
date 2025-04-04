package go_connectwise

import (
	"encoding/json"
	"os"
	"testing"
)

type OpenApiTagData struct {
	Openapi string `json:"openapi"`
	Info    struct {
		Title   string `json:"title"`
		Version string `json:"version"`
	} `json:"info"`
	Servers []struct {
		Url string `json:"url"`
	} `json:"servers"`
	Paths map[string]struct {
		Methods map[string]struct {
			Tags []string `json:"tags"`
		} `json:"*"`
	} `json:"paths"`
}

func TestCreateBuildConfigs(t *testing.T) {
	// Because of the size of the cw-api.json, we split up the files for the editors' sake.
	// Load cw-api.json.
	cwApiJsonData, err := os.ReadFile("./cw-api.json")
	if err != nil {
		t.Fatalf("Failed to read cw-api.json: %v", err)
	}
	// Parse the json data into OpenApiTagData struct.
	var openApiTagData OpenApiTagData
	err = json.Unmarshal(cwApiJsonData, &openApiTagData)
	if err != nil {
		t.Fatalf("Failed to unmarshal cw-api.json: %v", err)
	}
	// Search all the paths, Methods, and Tags to find the unique tags for the API.
	if len(openApiTagData.Paths) == 0 {
		t.Fatal("No paths found in cw-api.json")
	}
	// Create a map to store unique tags.
	tagMap := make(map[string]bool)
	for _, pathValue := range openApiTagData.Paths {
		if len(pathValue.Methods) == 0 {
			t.Logf("Skipping path with no methods defined")
			continue // Skip if no methods are defined for the path.
		}
		for _, method := range pathValue.Methods {
			// Iterate over each method's tags.
			for _, tag := range method.Tags {
				tagMap[tag] = true // Add unique tag to the map.
			}
		}
	}
	// Now we can create the build configs for each tag.
	if len(tagMap) == 0 {
		t.Fatal("No tags found in cw-api.json")
	}
	// Loop through the tagMap and create build configs. output to console for verification.
	for tag, _ := range tagMap {
		// Create the build config string.
		buildConfig := `
# yaml-language-server: $schema=https://raw.githubusercontent.com/oapi-codegen/oapi-codegen/HEAD/configuration-schema.json
package: go_connectwise
generate:
  models: true
  client: true
output-options:
  include-tags:
    - ` + tag + `
output: connectwise-` + tag + `.gen.go`
		// Write the build config to a file.
		fileName := "connectwise-" + tag + ".gen.yaml"
		f, err2 := os.Create(fileName)
		if err2 != nil {
			t.Fatalf("Failed to create file %s: %v", fileName, err2)
		}
		defer f.Close() // Ensure the file is closed after writing.
		_, err2 = f.WriteString(buildConfig)
		if err2 != nil {
			t.Fatalf("Failed to write to file %s: %v", fileName, err2)
		}
	}
}
