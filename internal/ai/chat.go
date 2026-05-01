package ai

import (
	"context"
	"fmt"
	"strings"

	"reporadar/internal/db"
	"reporadar/internal/model"
	"reporadar/internal/ollama"
)

type Service struct {
	repo   *db.Repository
	ollama *ollama.Client
}

type ChatRequest struct {
	AssetID  string           `json:"asset_id"`
	ScanID   string           `json:"scan_id"`
	Question string           `json:"question"`
	History  []ollama.Message `json:"history"`
}

type ChatResponse struct {
	Model  string `json:"model"`
	Answer string `json:"answer"`
}

func NewService(repo *db.Repository, client *ollama.Client) *Service {
	return &Service{repo: repo, ollama: client}
}

func (s *Service) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if s.ollama == nil {
		return nil, fmt.Errorf("ollama client is not configured")
	}
	if strings.TrimSpace(req.ScanID) == "" {
		return nil, fmt.Errorf("scan_id is required")
	}
	if strings.TrimSpace(req.Question) == "" {
		return nil, fmt.Errorf("question is required")
	}

	contextBlock, err := s.buildContext(ctx, req.AssetID, req.ScanID)
	if err != nil {
		return nil, err
	}

	messages := []ollama.Message{
		{
			Role: "system",
			Content: "You are RepoRadar Analyst. Answer only from the provided scan context. " +
				"If the user asks for something not present in the context, say that clearly. " +
				"Be concise, concrete, and prioritize remediation order when relevant.",
		},
		{
			Role:    "system",
			Content: contextBlock,
		},
	}

	for _, msg := range req.History {
		if msg.Role != "user" && msg.Role != "assistant" {
			continue
		}
		messages = append(messages, msg)
	}
	messages = append(messages, ollama.Message{
		Role:    "user",
		Content: req.Question,
	})

	answer, err := s.ollama.Chat(ctx, messages)
	if err != nil {
		return nil, err
	}

	return &ChatResponse{
		Model:  s.ollama.Model(),
		Answer: answer,
	}, nil
}

func (s *Service) buildContext(ctx context.Context, assetID, scanID string) (string, error) {
	scanSummary, err := s.repo.GetScanSummary(ctx, scanID)
	if err != nil {
		return "", fmt.Errorf("load scan summary: %w", err)
	}
	if scanSummary == nil {
		return "", fmt.Errorf("scan %q not found", scanID)
	}

	var assetSummary *model.AssetSummary
	if strings.TrimSpace(assetID) != "" {
		assetSummary, err = s.repo.GetAssetSummary(ctx, assetID)
		if err != nil {
			return "", fmt.Errorf("load asset summary: %w", err)
		}
	}

	findingsPage, err := s.repo.ListExpandedFindingsPageByScanID(ctx, scanID, model.FindingsFilter{
		Limit:  20,
		Offset: 0,
		SortBy: "severity",
		Order:  "desc",
	})
	if err != nil {
		return "", fmt.Errorf("load scan findings: %w", err)
	}

	var b strings.Builder
	b.WriteString("RepoRadar scan context\n")
	b.WriteString(fmt.Sprintf("scan_id: %s\n", scanID))
	b.WriteString(fmt.Sprintf("scan_total_findings: %d\n", scanSummary.TotalFindings))
	b.WriteString(fmt.Sprintf("scan_unique_vulnerabilities: %d\n", scanSummary.UniqueVulnerabilities))
	b.WriteString(fmt.Sprintf("scan_unique_packages_affected: %d\n", scanSummary.UniquePackagesAffected))
	b.WriteString(fmt.Sprintf("scan_severity_counts: %s\n", formatIntMap(scanSummary.SeverityCounts)))
	b.WriteString(fmt.Sprintf("scan_ecosystem_counts: %s\n", formatIntMap(scanSummary.EcosystemCounts)))

	if assetSummary != nil {
		b.WriteString(fmt.Sprintf("asset_id: %s\n", assetSummary.AssetID))
		b.WriteString(fmt.Sprintf("asset_total_scans: %d\n", assetSummary.TotalScans))
		b.WriteString(fmt.Sprintf("asset_total_findings: %d\n", assetSummary.TotalFindings))
		b.WriteString(fmt.Sprintf("asset_unique_vulnerabilities: %d\n", assetSummary.UniqueVulnerabilities))
		b.WriteString(fmt.Sprintf("asset_latest_scan_id: %s\n", assetSummary.LatestScanID))
		b.WriteString(fmt.Sprintf("asset_severity_counts: %s\n", formatIntMap(assetSummary.SeverityCounts)))
	}

	b.WriteString("top_findings:\n")
	for i, finding := range findingsPage.Items {
		score := "unscored"
		if finding.Vulnerability.SeverityScore != nil {
			score = fmt.Sprintf("%.1f", *finding.Vulnerability.SeverityScore)
		}

		b.WriteString(fmt.Sprintf(
			"%d. package=%s version=%s ecosystem=%s vulnerability=%s severity_label=%s severity_score=%s status=%s summary=%s\n",
			i+1,
			finding.Component.Name,
			finding.ComponentVersion.Version,
			finding.Component.Ecosystem,
			finding.Vulnerability.ExternalID,
			finding.Vulnerability.SeverityLabel,
			score,
			finding.Finding.Status,
			sanitizeLine(finding.Vulnerability.Summary),
		))
	}

	return b.String(), nil
}

func formatIntMap(values map[string]int) string {
	if len(values) == 0 {
		return "{}"
	}

	parts := make([]string, 0, len(values))
	for key, value := range values {
		parts = append(parts, fmt.Sprintf("%s=%d", key, value))
	}
	return strings.Join(parts, ", ")
}

func sanitizeLine(value string) string {
	return strings.ReplaceAll(strings.TrimSpace(value), "\n", " ")
}
