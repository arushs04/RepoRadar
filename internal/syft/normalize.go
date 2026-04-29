package syft

import (
	"fmt"
	"strings"

	"supplygraph/internal/model"
)

// This file contains the function to normalize a Syft artifact into our internal model representation, decides whether it can be normalized.
func NormalizeArtifact(artifact Artifact) (model.NormalizedArtifact, bool, error) {
	// We require a PURL to be able to normalize, since that's how we determine the ecosystem and component identity.
	if artifact.PURL == "" {
		return model.NormalizedArtifact{}, false, nil
	}

	// We also require a version to be able to normalize, since our model has a separate ComponentVersion entity and we want to populate that if possible. If there's no version, or if the version is "UNKNOWN", then we consider it ineligible for normalization.
	if artifact.Version == "" || artifact.Version == "UNKNOWN" {
		return model.NormalizedArtifact{}, false, nil
	}

	ecosystem, componentPURL, err := parsePackageIdentity(artifact.PURL)
	if err != nil {
		return model.NormalizedArtifact{}, false, fmt.Errorf("parse package identity: %w", err)
	}

	return model.NormalizedArtifact{
		Component: model.Component{
			Name:      artifact.Name,
			Ecosystem: ecosystem,
			PURL:      componentPURL,
		},
		ComponentVersion: model.ComponentVersion{
			Version: artifact.Version,
		},
	}, true, nil
}

func parsePackageIdentity(purl string) (string, string, error) {
	// A valid PURL for our purposes looks like: pkg:<ecosystem>/<name>@<version>
	const prefix = "pkg:"
	if !strings.HasPrefix(purl, prefix) {
		return "", "", fmt.Errorf("invalid purl %q", purl)
	}

	// Strip the "pkg:" prefix, then split on the last "@" to separate the identity from the version. The identity part should contain the ecosystem and name, separated by a "/".
	withoutPrefix := strings.TrimPrefix(purl, prefix)
	atIndex := strings.LastIndex(withoutPrefix, "@")
	if atIndex == -1 {
		return "", "", fmt.Errorf("missing version separator in purl %q", purl)
	}

	// The part before the "@" is the identity, which includes the ecosystem and name. The part after the "@" is the version, which we don't need to parse here since it's already provided separately in the Syft artifact.
	identity := withoutPrefix[:atIndex]
	slashIndex := strings.Index(identity, "/")
	if slashIndex == -1 {
		return "", "", fmt.Errorf("missing ecosystem separator in purl %q", purl)
	}

	ecosystem := identity[:slashIndex]
	componentPURL := prefix + identity

	return ecosystem, componentPURL, nil
}
