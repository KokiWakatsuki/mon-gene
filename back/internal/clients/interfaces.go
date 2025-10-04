package clients

import (
	"context"
)

// ClaudeClient defines the interface for Claude API interactions
type ClaudeClient interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
}

// CoreClient defines the interface for Core API interactions
type CoreClient interface {
	AnalyzeProblem(ctx context.Context, problemText string, filters map[string]interface{}) (*CoreAnalysisResponse, error)
	GenerateGeometry(ctx context.Context, shapeType string, parameters map[string]interface{}) (string, error)
	GeneratePDF(ctx context.Context, problemText, imageBase64, solutionText string) (string, error)
	GenerateCustomGeometry(ctx context.Context, pythonCode, problemText string) (string, error)
}

// Core API response types
type CoreAnalysisResponse struct {
	Success             bool                              `json:"success"`
	NeedsGeometry       bool                              `json:"needs_geometry"`
	DetectedShapes      []string                          `json:"detected_shapes"`
	SuggestedParameters map[string]map[string]interface{} `json:"suggested_parameters"`
}

type CoreGeometryResponse struct {
	Success     bool   `json:"success"`
	ImageBase64 string `json:"image_base64"`
	ShapeType   string `json:"shape_type"`
}

type CorePDFResponse struct {
	Success   bool   `json:"success"`
	PDFBase64 string `json:"pdf_base64"`
}

type CoreCustomGeometryResponse struct {
	Success     bool   `json:"success"`
	ImageBase64 string `json:"image_base64"`
	ProblemText string `json:"problem_text"`
}
