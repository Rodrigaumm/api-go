package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

type WebhookHandler struct{}

func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{}
}

type ProcessByPidRequest struct {
	PID int `json:"pid"`
}

type ProcessInfo struct {
	ProcessID              int    `json:"processId"`
	ParentProcessID        int    `json:"parentProcessId"`
	ProcessName            string `json:"processName"`
	ThreadCount            int    `json:"threadCount"`
	HandleCount            int    `json:"handleCount"`
	BasePriority           int    `json:"basePriority"`
	CreateTime             string `json:"createTime"`
	UserTime               string `json:"userTime"`
	KernelTime             string `json:"kernelTime"`
	WorkingSetSize         string `json:"workingSetSize"`
	PeakWorkingSetSize     string `json:"peakWorkingSetSize"`
	VirtualSize            string `json:"virtualSize"`
	PeakVirtualSize        string `json:"peakVirtualSize"`
	PagefileUsage          string `json:"pagefileUsage"`
	PeakPagefileUsage      string `json:"peakPagefileUsage"`
	PageFaultCount         int    `json:"pageFaultCount"`
	ReadOperationCount     int64  `json:"readOperationCount"`
	WriteOperationCount    int64  `json:"writeOperationCount"`
	OtherOperationCount    int64  `json:"otherOperationCount"`
	ReadTransferCount      int64  `json:"readTransferCount"`
	WriteTransferCount     int64  `json:"writeTransferCount"`
	OtherTransferCount     int64  `json:"otherTransferCount"`
	CurrentProcessAddress  string `json:"currentProcessAddress"`
	NextProcessAddress     string `json:"nextProcessAddress"`
	PreviousProcessAddress string `json:"previousProcessAddress"`
}

type ProcessBasic struct {
	Index       int    `json:"index"`
	ProcessName string `json:"processName"`
	ProcessID   int    `json:"processId"`
}

type IterateProcessesResponse struct {
	Success      bool           `json:"success"`
	ProcessCount int            `json:"processCount"`
	Processes    []ProcessBasic `json:"processes"`
}

type ProcessByPidResponse struct {
	Success     bool        `json:"success"`
	ProcessInfo ProcessInfo `json:"processInfo"`
	Error       string      `json:"error,omitempty"`
}

func (h *WebhookHandler) makeHTTPRequest(url string, method string, body interface{}) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("erro ao serializar body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer requisição: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("requisição falhou com status %d: %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

func (h *WebhookHandler) IterateProcesses(c *fiber.Ctx) error {
	webhookURL := c.Query("webhookurl")
	if webhookURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Parâmetro 'webhookurl' é obrigatório",
		})
	}

	fullURL := webhookURL + "/webhook/iterate-processes"

	responseBody, err := h.makeHTTPRequest(fullURL, "POST", nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Erro ao fazer requisição para %s: %v", fullURL, err),
		})
	}

	var response IterateProcessesResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return c.JSON(fiber.Map{
			"data": string(responseBody),
		})
	}

	return c.JSON(fiber.Map{
		"data": response,
	})
}

func (h *WebhookHandler) ProcessByPid(c *fiber.Ctx) error {
	webhookURL := c.Query("webhookurl")
	if webhookURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Parâmetro 'webhookurl' é obrigatório",
		})
	}

	var reqBody ProcessByPidRequest
	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Erro ao fazer parse do body da requisição",
		})
	}

	if reqBody.PID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "PID deve ser um número positivo",
		})
	}

	fullURL := webhookURL + "/webhook/process-by-pid"

	responseBody, err := h.makeHTTPRequest(fullURL, "POST", reqBody)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Erro ao fazer requisição para %s: %v", fullURL, err),
		})
	}

	var response ProcessByPidResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return c.JSON(fiber.Map{
			"data": string(responseBody),
		})
	}

	return c.JSON(fiber.Map{
		"data": response,
	})
}
