package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type coreClient struct {
	baseURL string
	client  *http.Client
}

func NewCoreClient() CoreClient {
	baseURL := os.Getenv("CORE_API_URL")
	if baseURL == "" {
		baseURL = "http://core:1234" // デフォルトはDockerコンテナ名
	}
	
	return &coreClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *coreClient) AnalyzeProblem(ctx context.Context, problemText string, filters map[string]interface{}) (*CoreAnalysisResponse, error) {
	requestData := map[string]interface{}{
		"problem_text":     problemText,
		"unit_parameters":  filters,
		"subject":          "math",
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/analyze-problem", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response CoreAnalysisResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

func (c *coreClient) GenerateGeometry(ctx context.Context, shapeType string, parameters map[string]interface{}) (string, error) {
	requestData := map[string]interface{}{
		"shape_type": shapeType,
		"parameters": parameters,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/draw-geometry", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response CoreGeometryResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response.ImageBase64, nil
}

func (c *coreClient) GeneratePDF(ctx context.Context, problemText, imageBase64, solutionText string) (string, error) {
	requestData := map[string]interface{}{
		"problem_text":  problemText,
		"image_base64":  imageBase64,
		"solution_text": solutionText,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/generate-pdf", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response CorePDFResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response.PDFBase64, nil
}

func (c *coreClient) GenerateCustomGeometry(ctx context.Context, pythonCode, problemText string) (string, error) {
	fmt.Printf("🔍 GenerateCustomGeometry called with pythonCode length: %d\n", len(pythonCode))
	fmt.Printf("🔍 problemText: %s\n", problemText)
	
	requestData := map[string]interface{}{
		"python_code":  pythonCode,
		"problem_text": problemText,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	fmt.Printf("🔍 Sending request to: %s/draw-custom-geometry\n", c.baseURL)
	fmt.Printf("🔍 Request data: %s\n", string(jsonData))

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/draw-custom-geometry", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("🔍 Core API response status: %d\n", resp.StatusCode)
	fmt.Printf("🔍 Core API response body length: %d\n", len(body))
	fmt.Printf("🔍 Core API response body (first 200 chars): %s\n", string(body[:min(200, len(body))]))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// まず生のJSONをパースして内容を確認
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal raw response: %w", err)
	}
	
	fmt.Printf("🔍 Raw response keys: %v\n", getKeys(rawResponse))
	fmt.Printf("🔍 Raw response success: %v\n", rawResponse["success"])
	fmt.Printf("🔍 Raw response image_base64 exists: %v\n", rawResponse["image_base64"] != nil)
	if rawResponse["image_base64"] != nil {
		if imageStr, ok := rawResponse["image_base64"].(string); ok {
			fmt.Printf("🔍 Raw response image_base64 length: %d\n", len(imageStr))
		}
	}

	var response CoreCustomGeometryResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	fmt.Printf("🔍 Parsed response success: %v\n", response.Success)
	fmt.Printf("🔍 Parsed response ImageBase64 length: %d\n", len(response.ImageBase64))

	return response.ImageBase64, nil
}

func (c *coreClient) ExecutePython(ctx context.Context, pythonCode string) (string, error) {
	fmt.Printf("🔍 ExecutePython called with code length: %d\n", len(pythonCode))
	
	requestData := map[string]interface{}{
		"python_code": pythonCode,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	fmt.Printf("🔍 Sending Python execution request to: %s/execute-python\n", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/execute-python", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("🔍 Python execution response status: %d\n", resp.StatusCode)
	fmt.Printf("🔍 Python execution response length: %d\n", len(body))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Python execution failed with status %d: %s", resp.StatusCode, string(body))
	}

	// レスポンスの構造を確認
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	fmt.Printf("🔍 Python execution response keys: %v\n", getKeys(rawResponse))
	
	// 実行結果を取得
	if success, ok := rawResponse["success"].(bool); !ok || !success {
		errorMsg := "Unknown error"
		if errStr, exists := rawResponse["error"].(string); exists {
			errorMsg = errStr
		}
		return "", fmt.Errorf("Python execution failed: %s", errorMsg)
	}

	// 実行結果（stdout）を取得
	output := ""
	if outputStr, exists := rawResponse["output"].(string); exists {
		output = outputStr
	} else if resultStr, exists := rawResponse["result"].(string); exists {
		output = resultStr
	} else if stdoutStr, exists := rawResponse["stdout"].(string); exists {
		output = stdoutStr
	}

	fmt.Printf("🔍 Python execution output length: %d\n", len(output))
	
	return output, nil
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
