package mcpserver

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"reporadar/internal/db"
	"reporadar/internal/model"
)

type Server struct {
	repo *db.Repository
}

func New(repo *db.Repository) *Server {
	return &Server{repo: repo}
}

func (s *Server) MCP() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "reporadar-mcp",
		Version: "v0.1.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_assets",
		Description: "List all known assets that have been ingested into RepoRadar.",
	}, s.listAssets)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_asset_summary",
		Description: "Return the aggregated summary for an asset across all scans.",
	}, s.getAssetSummary)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_scan_summary",
		Description: "Return the aggregated summary for a single scan.",
	}, s.getScanSummary)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_asset_findings",
		Description: "List findings for an asset with optional filtering, pagination, and sorting.",
	}, s.listAssetFindings)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_scan_findings",
		Description: "List findings for a scan with optional filtering, pagination, and sorting.",
	}, s.listScanFindings)

	return server
}

type AssetSummaryInput struct {
	AssetID string `json:"asset_id" jsonschema:"asset UUID to summarize"`
}

type ListAssetsInput struct{}

type ScanSummaryInput struct {
	ScanID string `json:"scan_id" jsonschema:"scan UUID to summarize"`
}

type FindingsInput struct {
	ID            string `json:"id" jsonschema:"asset UUID or scan UUID, depending on the tool"`
	Limit         int    `json:"limit,omitempty" jsonschema:"maximum findings to return; default 50, max 200"`
	Offset        int    `json:"offset,omitempty" jsonschema:"number of findings to skip before returning results"`
	Ecosystem     string `json:"ecosystem,omitempty" jsonschema:"exact ecosystem filter, for example npm"`
	Package       string `json:"package,omitempty" jsonschema:"exact package name filter"`
	Status        string `json:"status,omitempty" jsonschema:"exact finding status filter"`
	Vulnerability string `json:"vulnerability,omitempty" jsonschema:"exact vulnerability external ID filter"`
	SeverityLabel string `json:"severity_label,omitempty" jsonschema:"severity bucket filter: critical, high, medium, low, none, unknown"`
	SortBy        string `json:"sort_by,omitempty" jsonschema:"sort key: id, package, version, vulnerability, severity, scan_id"`
	Order         string `json:"order,omitempty" jsonschema:"sort order: asc or desc"`
}

type AssetSummaryOutput struct {
	AssetID                string         `json:"asset_id"`
	TotalScans             int            `json:"total_scans"`
	LatestScanID           string         `json:"latest_scan_id"`
	TotalFindings          int            `json:"total_findings"`
	UniqueVulnerabilities  int            `json:"unique_vulnerabilities"`
	UniquePackagesAffected int            `json:"unique_packages_affected"`
	EcosystemCounts        map[string]int `json:"ecosystem_counts"`
	SeverityCounts         map[string]int `json:"severity_counts"`
}

type AssetOutput struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AssetType string `json:"asset_type"`
	Source    string `json:"source"`
	CreatedAt string `json:"created_at"`
}

type ScanSummaryOutput struct {
	ScanID                 string         `json:"scan_id"`
	TotalFindings          int            `json:"total_findings"`
	UniqueVulnerabilities  int            `json:"unique_vulnerabilities"`
	UniquePackagesAffected int            `json:"unique_packages_affected"`
	EcosystemCounts        map[string]int `json:"ecosystem_counts"`
	SeverityCounts         map[string]int `json:"severity_counts"`
}

type FindingsPageOutput struct {
	Items  []FindingOutput `json:"items"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

type FindingOutput struct {
	ID               string                 `json:"id"`
	ScanID           string                 `json:"scan_id"`
	Status           string                 `json:"status"`
	FixedVersion     string                 `json:"fixed_version"`
	Vulnerability    VulnerabilityOutput    `json:"vulnerability"`
	ComponentVersion ComponentVersionOutput `json:"component_version"`
}

type VulnerabilityOutput struct {
	ID            string   `json:"id"`
	ExternalID    string   `json:"external_id"`
	Source        string   `json:"source"`
	Severity      string   `json:"severity"`
	SeverityScore *float64 `json:"severity_score"`
	SeverityLabel string   `json:"severity_label"`
	Summary       string   `json:"summary"`
}

type ComponentVersionOutput struct {
	ID        string          `json:"id"`
	Version   string          `json:"version"`
	Component ComponentOutput `json:"component"`
}

type ComponentOutput struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Ecosystem string `json:"ecosystem"`
	PURL      string `json:"purl"`
}

func (s *Server) listAssets(ctx context.Context, _ *mcp.CallToolRequest, _ ListAssetsInput) (*mcp.CallToolResult, []AssetOutput, error) {
	assets, err := s.repo.ListAssets(ctx)
	if err != nil {
		return nil, nil, err
	}

	response := make([]AssetOutput, 0, len(assets))
	for _, asset := range assets {
		response = append(response, newAssetOutput(asset))
	}

	return nil, response, nil
}

func (s *Server) getAssetSummary(ctx context.Context, _ *mcp.CallToolRequest, input AssetSummaryInput) (*mcp.CallToolResult, AssetSummaryOutput, error) {
	if strings.TrimSpace(input.AssetID) == "" {
		return nil, AssetSummaryOutput{}, fmt.Errorf("asset_id is required")
	}

	summary, err := s.repo.GetAssetSummary(ctx, input.AssetID)
	if err != nil {
		return nil, AssetSummaryOutput{}, err
	}

	return nil, newAssetSummaryOutput(summary), nil
}

func (s *Server) getScanSummary(ctx context.Context, _ *mcp.CallToolRequest, input ScanSummaryInput) (*mcp.CallToolResult, ScanSummaryOutput, error) {
	if strings.TrimSpace(input.ScanID) == "" {
		return nil, ScanSummaryOutput{}, fmt.Errorf("scan_id is required")
	}

	summary, err := s.repo.GetScanSummary(ctx, input.ScanID)
	if err != nil {
		return nil, ScanSummaryOutput{}, err
	}

	return nil, newScanSummaryOutput(summary), nil
}

func (s *Server) listAssetFindings(ctx context.Context, _ *mcp.CallToolRequest, input FindingsInput) (*mcp.CallToolResult, FindingsPageOutput, error) {
	if strings.TrimSpace(input.ID) == "" {
		return nil, FindingsPageOutput{}, fmt.Errorf("id is required")
	}

	filter, err := normalizeFindingsFilter(input)
	if err != nil {
		return nil, FindingsPageOutput{}, err
	}

	page, err := s.repo.ListExpandedFindingsPageByAssetID(ctx, input.ID, filter)
	if err != nil {
		return nil, FindingsPageOutput{}, err
	}

	return nil, newFindingsPageOutput(page), nil
}

func (s *Server) listScanFindings(ctx context.Context, _ *mcp.CallToolRequest, input FindingsInput) (*mcp.CallToolResult, FindingsPageOutput, error) {
	if strings.TrimSpace(input.ID) == "" {
		return nil, FindingsPageOutput{}, fmt.Errorf("id is required")
	}

	filter, err := normalizeFindingsFilter(input)
	if err != nil {
		return nil, FindingsPageOutput{}, err
	}

	page, err := s.repo.ListExpandedFindingsPageByScanID(ctx, input.ID, filter)
	if err != nil {
		return nil, FindingsPageOutput{}, err
	}

	return nil, newFindingsPageOutput(page), nil
}

func normalizeFindingsFilter(input FindingsInput) (model.FindingsFilter, error) {
	const (
		defaultLimit = 50
		maxLimit     = 200
	)

	filter := model.FindingsFilter{
		Limit:                   defaultLimit,
		Offset:                  0,
		Ecosystem:               strings.TrimSpace(input.Ecosystem),
		Package:                 strings.TrimSpace(input.Package),
		Status:                  strings.TrimSpace(input.Status),
		VulnerabilityExternalID: strings.TrimSpace(input.Vulnerability),
		SeverityLabel:           strings.TrimSpace(input.SeverityLabel),
		SortBy:                  strings.TrimSpace(input.SortBy),
		Order:                   strings.ToLower(strings.TrimSpace(input.Order)),
	}

	if input.Limit < 0 {
		return model.FindingsFilter{}, fmt.Errorf("limit must be non-negative")
	}
	if input.Offset < 0 {
		return model.FindingsFilter{}, fmt.Errorf("offset must be non-negative")
	}
	if input.Limit > 0 {
		filter.Limit = input.Limit
	}
	if filter.Limit > maxLimit {
		filter.Limit = maxLimit
	}
	filter.Offset = input.Offset

	if filter.Order != "" && filter.Order != "asc" && filter.Order != "desc" {
		return model.FindingsFilter{}, fmt.Errorf("order must be asc or desc")
	}

	return filter, nil
}

func newAssetOutput(asset *model.Asset) AssetOutput {
	return AssetOutput{
		ID:        asset.ID,
		Name:      asset.Name,
		AssetType: asset.AssetType,
		Source:    asset.Source,
		CreatedAt: asset.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func newAssetSummaryOutput(summary *model.AssetSummary) AssetSummaryOutput {
	return AssetSummaryOutput{
		AssetID:                summary.AssetID,
		TotalScans:             summary.TotalScans,
		LatestScanID:           summary.LatestScanID,
		TotalFindings:          summary.TotalFindings,
		UniqueVulnerabilities:  summary.UniqueVulnerabilities,
		UniquePackagesAffected: summary.UniquePackagesAffected,
		EcosystemCounts:        cloneIntMap(summary.EcosystemCounts),
		SeverityCounts:         defaultSeverityCounts(summary.SeverityCounts),
	}
}

func newScanSummaryOutput(summary *model.ScanSummary) ScanSummaryOutput {
	return ScanSummaryOutput{
		ScanID:                 summary.ScanID,
		TotalFindings:          summary.TotalFindings,
		UniqueVulnerabilities:  summary.UniqueVulnerabilities,
		UniquePackagesAffected: summary.UniquePackagesAffected,
		EcosystemCounts:        cloneIntMap(summary.EcosystemCounts),
		SeverityCounts:         defaultSeverityCounts(summary.SeverityCounts),
	}
}

func newFindingsPageOutput(page *model.FindingsPage) FindingsPageOutput {
	items := make([]FindingOutput, 0, len(page.Items))
	for _, finding := range page.Items {
		items = append(items, newFindingOutput(finding))
	}

	return FindingsPageOutput{
		Items:  items,
		Total:  page.Total,
		Limit:  page.Limit,
		Offset: page.Offset,
	}
}

func newFindingOutput(finding model.ExpandedFinding) FindingOutput {
	return FindingOutput{
		ID:           finding.Finding.ID,
		ScanID:       finding.Finding.ScanID,
		Status:       finding.Finding.Status,
		FixedVersion: finding.Finding.FixedVersion,
		Vulnerability: VulnerabilityOutput{
			ID:            finding.Vulnerability.ID,
			ExternalID:    finding.Vulnerability.ExternalID,
			Source:        finding.Vulnerability.Source,
			Severity:      finding.Vulnerability.Severity,
			SeverityScore: finding.Vulnerability.SeverityScore,
			SeverityLabel: finding.Vulnerability.SeverityLabel,
			Summary:       finding.Vulnerability.Summary,
		},
		ComponentVersion: ComponentVersionOutput{
			ID:      finding.ComponentVersion.ID,
			Version: finding.ComponentVersion.Version,
			Component: ComponentOutput{
				ID:        finding.Component.ID,
				Name:      finding.Component.Name,
				Ecosystem: finding.Component.Ecosystem,
				PURL:      finding.Component.PURL,
			},
		},
	}
}

func defaultSeverityCounts(counts map[string]int) map[string]int {
	out := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
		"none":     0,
		"unknown":  0,
	}
	for label, count := range counts {
		out[label] = count
	}
	return out
}

func cloneIntMap(in map[string]int) map[string]int {
	if in == nil {
		return map[string]int{}
	}
	out := make(map[string]int, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}
