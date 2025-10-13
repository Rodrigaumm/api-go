package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go-api/internal/db"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebhookHandler struct {
	queries *db.Queries
}

func NewWebhookHandler(dbpool *pgxpool.Pool) *WebhookHandler {
	return &WebhookHandler{
		queries: db.New(dbpool),
	}
}

type ProcessByPidRequest struct {
	PID int `json:"pid"`
}

type ProcessInfo struct {
	ProcessID             int             `json:"processId"`
	ParentProcessID       int             `json:"parentProcessId"`
	ProcessName           string          `json:"processName"`
	ThreadCount           int             `json:"threadCount"`
	HandleCount           int             `json:"handleCount"`
	BasePriority          int             `json:"basePriority"`
	CreateTime            string          `json:"createTime"`
	UserTime              string          `json:"userTime"`
	KernelTime            string          `json:"kernelTime"`
	WorkingSetSize        string          `json:"workingSetSize"`
	PeakWorkingSetSize    string          `json:"peakWorkingSetSize"`
	VirtualSize           string          `json:"virtualSize"`
	PeakVirtualSize       string          `json:"peakVirtualSize"`
	PagefileUsage         string          `json:"pagefileUsage"`
	PeakPagefileUsage     string          `json:"peakPagefileUsage"`
	PageFaultCount        int             `json:"pageFaultCount"`
	ReadOperationCount    int64           `json:"readOperationCount"`
	WriteOperationCount   int64           `json:"writeOperationCount"`
	OtherOperationCount   int64           `json:"otherOperationCount"`
	ReadTransferCount     int64           `json:"readTransferCount"`
	WriteTransferCount    int64           `json:"writeTransferCount"`
	OtherTransferCount    int64           `json:"otherTransferCount"`
	CurrentProcessAddress string          `json:"currentProcessAddress"`
	NextProcess           AdjacentProcess `json:"nextProcess"`
	PreviousProcess       AdjacentProcess `json:"previousProcess"`
}

type AdjacentProcess struct {
	EProcessAddress string `json:"eProcessAddress"`
	ProcessName     string `json:"processName"`
	ProcessID       int    `json:"processId"`
}

type IterateProcessesResponse struct {
	Success      bool          `json:"success"`
	ProcessCount int           `json:"processCount"`
	Processes    []ProcessInfo `json:"processes"`
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

// persistProcessInfo persists a single process info to the database
func (h *WebhookHandler) persistProcessInfo(ctx context.Context, userID *int64, processInfo ProcessInfo) (int64, error) {
	var userIDParam pgtype.Int8
	if userID != nil {
		userIDParam = pgtype.Int8{Int64: *userID, Valid: true}
	} else {
		userIDParam = pgtype.Int8{Valid: false}
	}

	// Helper function to convert string to pgtype.Text
	toText := func(s string) pgtype.Text {
		if s == "" {
			return pgtype.Text{Valid: false}
		}
		return pgtype.Text{String: s, Valid: true}
	}

	// Helper function to convert int to pgtype.Int4
	toInt4 := func(i int) pgtype.Int4 {
		if i == 0 {
			return pgtype.Int4{Valid: false}
		}
		return pgtype.Int4{Int32: int32(i), Valid: true}
	}

	params := db.CreateProcessInfoParams{
		UserID:                         userIDParam,
		ProcessID:                      int32(processInfo.ProcessID),
		ParentProcessID:                int32(processInfo.ParentProcessID),
		ProcessName:                    processInfo.ProcessName,
		ThreadCount:                    int32(processInfo.ThreadCount),
		HandleCount:                    int32(processInfo.HandleCount),
		BasePriority:                   int32(processInfo.BasePriority),
		CreateTime:                     processInfo.CreateTime,
		UserTime:                       processInfo.UserTime,
		KernelTime:                     processInfo.KernelTime,
		WorkingSetSize:                 processInfo.WorkingSetSize,
		PeakWorkingSetSize:             processInfo.PeakWorkingSetSize,
		VirtualSize:                    processInfo.VirtualSize,
		PeakVirtualSize:                processInfo.PeakVirtualSize,
		PagefileUsage:                  processInfo.PagefileUsage,
		PeakPagefileUsage:              processInfo.PeakPagefileUsage,
		PageFaultCount:                 int32(processInfo.PageFaultCount),
		ReadOperationCount:             processInfo.ReadOperationCount,
		WriteOperationCount:            processInfo.WriteOperationCount,
		OtherOperationCount:            processInfo.OtherOperationCount,
		ReadTransferCount:              processInfo.ReadTransferCount,
		WriteTransferCount:             processInfo.WriteTransferCount,
		OtherTransferCount:             processInfo.OtherTransferCount,
		CurrentProcessAddress:          processInfo.CurrentProcessAddress,
		NextProcessEprocessAddress:     toText(processInfo.NextProcess.EProcessAddress),
		NextProcessName:                toText(processInfo.NextProcess.ProcessName),
		NextProcessID:                  toInt4(processInfo.NextProcess.ProcessID),
		PreviousProcessEprocessAddress: toText(processInfo.PreviousProcess.EProcessAddress),
		PreviousProcessName:            toText(processInfo.PreviousProcess.ProcessName),
		PreviousProcessID:              toInt4(processInfo.PreviousProcess.ProcessID),
	}

	result, err := h.queries.CreateProcessInfo(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("erro ao persistir processo: %v", err)
	}

	return result.ID, nil
}

func (h *WebhookHandler) IterateProcesses(c *fiber.Ctx) error {
	webhookURL := c.Query("webhookurl")
	if webhookURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Parâmetro 'webhookurl' é obrigatório",
		})
	}

	fullURL := webhookURL + "/webhook/iterate-processes"

	// Make the HTTP request to the webhook
	responseBody, err := h.makeHTTPRequest(fullURL, "POST", nil)
	if err != nil {
		// If there's an error and user is authenticated, log the failed attempt
		userID, _, ok := GetUserFromContext(c)
		if ok {
			ctx := context.Background()
			var userIDParam pgtype.Int8
			userIDParam = pgtype.Int8{Int64: userID, Valid: true}

			_, _ = h.queries.CreateProcessIterationHistory(ctx, db.CreateProcessIterationHistoryParams{
				UserID:       userIDParam,
				WebhookUrl:   webhookURL,
				ProcessCount: 0,
				Success:      false,
				ErrorMessage: pgtype.Text{String: err.Error(), Valid: true},
			})
		}

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

	// Check if user is authenticated (has valid JWT)
	userID, _, ok := GetUserFromContext(c)
	if ok && response.Success {
		// User is authenticated, persist the data
		ctx := context.Background()

		var userIDParam pgtype.Int8
		userIDParam = pgtype.Int8{Int64: userID, Valid: true}

		// Create iteration history record
		iterationHistory, err := h.queries.CreateProcessIterationHistory(ctx, db.CreateProcessIterationHistoryParams{
			UserID:       userIDParam,
			WebhookUrl:   webhookURL,
			ProcessCount: int32(response.ProcessCount),
			Success:      true,
			ErrorMessage: pgtype.Text{Valid: false},
		})

		if err != nil {
			log.Errorf("Erro ao criar histórico de iteração: %v", err)
		} else {
			// Persist each process and link to iteration
			for _, processInfo := range response.Processes {
				processInfoID, err := h.persistProcessInfo(ctx, &userID, processInfo)
				if err != nil {
					log.Errorf("Erro ao persistir processo %d: %v", processInfo.ProcessID, err)
					continue
				}

				// Link process to iteration
				_, err = h.queries.CreateIterationProcess(ctx, db.CreateIterationProcessParams{
					IterationID:   iterationHistory.ID,
					ProcessInfoID: processInfoID,
				})
				if err != nil {
					log.Errorf("Erro ao vincular processo %d à iteração: %v", processInfo.ProcessID, err)
				}
			}

			log.Infof("Persistidos %d processos para o usuário %d", len(response.Processes), userID)
		}
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

	// Make the HTTP request to the webhook
	responseBody, err := h.makeHTTPRequest(fullURL, "POST", reqBody)
	if err != nil {
		// If there's an error and user is authenticated, log the failed attempt
		userID, _, ok := GetUserFromContext(c)
		if ok {
			ctx := context.Background()
			var userIDParam pgtype.Int8
			userIDParam = pgtype.Int8{Int64: userID, Valid: true}

			_, _ = h.queries.CreateProcessQueryHistory(ctx, db.CreateProcessQueryHistoryParams{
				UserID:        userIDParam,
				WebhookUrl:    webhookURL,
				RequestedPid:  int32(reqBody.PID),
				ProcessInfoID: pgtype.Int8{Valid: false},
				Success:       false,
				ErrorMessage:  pgtype.Text{String: err.Error(), Valid: true},
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Erro ao fazer requisição para %s: %v", fullURL, err),
		})
	}

	log.Info(responseBody)
	var response ProcessByPidResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return c.JSON(fiber.Map{
			"data": string(responseBody),
		})
	}

	// Check if user is authenticated (has valid JWT)
	userID, _, ok := GetUserFromContext(c)
	if ok && response.Success {
		// User is authenticated, persist the data
		ctx := context.Background()

		// Persist the process info
		processInfoID, err := h.persistProcessInfo(ctx, &userID, response.ProcessInfo)
		if err != nil {
			log.Errorf("Erro ao persistir processo %d: %v", response.ProcessInfo.ProcessID, err)
		} else {
			// Create query history record
			var userIDParam pgtype.Int8
			userIDParam = pgtype.Int8{Int64: userID, Valid: true}

			_, err = h.queries.CreateProcessQueryHistory(ctx, db.CreateProcessQueryHistoryParams{
				UserID:        userIDParam,
				WebhookUrl:    webhookURL,
				RequestedPid:  int32(reqBody.PID),
				ProcessInfoID: pgtype.Int8{Int64: processInfoID, Valid: true},
				Success:       true,
				ErrorMessage:  pgtype.Text{Valid: false},
			})

			if err != nil {
				log.Errorf("Erro ao criar histórico de query: %v", err)
			} else {
				log.Infof("Persistido processo %d para o usuário %d", response.ProcessInfo.ProcessID, userID)
			}
		}
	}

	return c.JSON(fiber.Map{
		"data": response,
	})
}
