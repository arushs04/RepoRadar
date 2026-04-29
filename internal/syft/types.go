package syft

// This file defines the data structures for the Syft SBOM document and its components.
type Document struct {
	Artifacts  []Artifact `json:"artifacts"`  // List of artifacts (components) in the SBOM
	Source     Source     `json:"source"`     // Information about the source of the SBOM, what exactly was scanned (e.g., where it was generated from)
	Descriptor Descriptor `json:"descriptor"` // Information about the tool that generated the SBOM, metadata (e.g., Syft version)
}

type Artifact struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Type    string `json:"type"` // What Syft says the discovered thing is.
	PURL    string `json:"purl"`
}

type Source struct { // The source of the SBOM, what exactly was scanned (e.g., where it was generated from)
	ID       string         `json:"id"`
	Name     string         `json:"name"` // The name of the source, which could be a file path, a URL, or a description of the scanned entity.
	Type     string         `json:"type"` // The type of the source, which could indicate whether it's a file system path, a Git repository, a container image, etc.
	Metadata SourceMetadata `json:"metadata"`
}

type SourceMetadata struct {
	Path string `json:"path"` // The file system path that was scanned, if applicable. This field is relevant when the source type indicates a file system scan. It provides the specific location of the scanned directory or file on the local machine.
}

type Descriptor struct {
	Name    string `json:"name"`    // The name of the tool that generated the SBOM
	Version string `json:"version"` // The version of the tool that generated the SBOM
}
