package syft

import (
	"encoding/json"
	"fmt"
	"os"
)

// This file contains the function to load a Syft SBOM document from a JSON file.
func LoadDocument(path string) (Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Document{}, fmt.Errorf("read syft file: %w", err)
	}

	return LoadDocumentBytes(data)
}

func LoadDocumentBytes(data []byte) (Document, error) {
	var doc Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return Document{}, fmt.Errorf("unmarshal syft json: %w", err)
	}
	// Successfully loaded the Syft document, return it.
	return doc, nil
}
