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
	Pid int32 `json:"pid"`
}

type ProcessInfo struct {
	ProcessID                      int64   `json:"ProcessID"`
	ParentProcessID                int64   `json:"ParentProcessID"`
	ProcessName                    string  `json:"ProcessName"`
	ThreadCount                    int32   `json:"ThreadCount"`
	HandleCount                    int32   `json:"HandleCount"`
	BasePriority                   int32   `json:"BasePriority"`
	CreateTime                     string  `json:"CreateTime"`
	UserTime                       int32   `json:"UserTime"`
	KernelTime                     int32   `json:"KernelTime"`
	WorkingSetSize                 int64   `json:"WorkingSetSize"`
	PeakWorkingSetSize             int64   `json:"PeakWorkingSetSize"`
	VirtualSize                    int64   `json:"VirtualSize"`
	PeakVirtualSize                int64   `json:"PeakVirtualSize"`
	ReadOperationCount             int64   `json:"ReadOperationCount"`
	WriteOperationCount            int64   `json:"WriteOperationCount"`
	OtherOperationCount            int64   `json:"OtherOperationCount"`
	ReadTransferCount              int64   `json:"ReadTransferCount"`
	WriteTransferCount             int64   `json:"WriteTransferCount"`
	OtherTransferCount             int64   `json:"OtherTransferCount"`
	CurrentProcessAddress          string  `json:"CurrentProcessAddress"`
	NextProcessEProcessAddress     *string `json:"NextProcessEProcessAddress,omitempty"`
	NextProcessName                *string `json:"NextProcessName,omitempty"`
	NextProcessID                  *int64  `json:"NextProcessID,omitempty"`
	PreviousProcessEProcessAddress *string `json:"PreviousProcessEProcessAddress,omitempty"`
	PreviousProcessName            *string `json:"PreviousProcessName,omitempty"`
	PreviousProcessID              *int64  `json:"PreviousProcessID,omitempty"`
}

type AdjacentProcess struct {
	EProcessAddress string `json:"EProcessAddress"`
	ProcessName     string `json:"ProcessName"`
	ProcessID       int64  `json:"ProcessID"`
}

type IterateProcessesResponse struct {
	Processes []ProcessInfo `json:"processes"`
}

type ProcessByPidResponse struct {
	Process ProcessInfo `json:"process"`
}

func (h *WebhookHandler) makeHTTPRequest(url string, method string, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (h *WebhookHandler) persistProcessInfo(ctx context.Context, snapshotID int64, userID *int64, processInfo ProcessInfo) (int64, error) {
	var userIDParam pgtype.Int8
	if userID != nil {
		userIDParam = pgtype.Int8{Int64: *userID, Valid: true}
	}

	var nextProcessEProcessAddress, nextProcessName pgtype.Text
	var nextProcessID pgtype.Int8
	if processInfo.NextProcessEProcessAddress != nil {
		nextProcessEProcessAddress = pgtype.Text{String: *processInfo.NextProcessEProcessAddress, Valid: true}
	}
	if processInfo.NextProcessName != nil {
		nextProcessName = pgtype.Text{String: *processInfo.NextProcessName, Valid: true}
	}
	if processInfo.NextProcessID != nil {
		nextProcessID = pgtype.Int8{Int64: *processInfo.NextProcessID, Valid: true}
	}

	var previousProcessEProcessAddress, previousProcessName pgtype.Text
	var previousProcessID pgtype.Int8
	if processInfo.PreviousProcessEProcessAddress != nil {
		previousProcessEProcessAddress = pgtype.Text{String: *processInfo.PreviousProcessEProcessAddress, Valid: true}
	}
	if processInfo.PreviousProcessName != nil {
		previousProcessName = pgtype.Text{String: *processInfo.PreviousProcessName, Valid: true}
	}
	if processInfo.PreviousProcessID != nil {
		previousProcessID = pgtype.Int8{Int64: *processInfo.PreviousProcessID, Valid: true}
	}

	createdProcess, err := h.queries.CreateProcessInfo(ctx, db.CreateProcessInfoParams{
		SnapshotID:                     snapshotID,
		UserID:                         userIDParam,
		ProcessID:                      processInfo.ProcessID,
		ParentProcessID:                processInfo.ParentProcessID,
		ProcessName:                    processInfo.ProcessName,
		ThreadCount:                    processInfo.ThreadCount,
		HandleCount:                    processInfo.HandleCount,
		BasePriority:                   processInfo.BasePriority,
		CreateTime:                     processInfo.CreateTime,
		UserTime:                       processInfo.UserTime,
		KernelTime:                     processInfo.KernelTime,
		WorkingSetSize:                 processInfo.WorkingSetSize,
		PeakWorkingSetSize:             processInfo.PeakWorkingSetSize,
		VirtualSize:                    processInfo.VirtualSize,
		PeakVirtualSize:                processInfo.PeakVirtualSize,
		ReadOperationCount:             processInfo.ReadOperationCount,
		WriteOperationCount:            processInfo.WriteOperationCount,
		OtherOperationCount:            processInfo.OtherOperationCount,
		ReadTransferCount:              processInfo.ReadTransferCount,
		WriteTransferCount:             processInfo.WriteTransferCount,
		OtherTransferCount:             processInfo.OtherTransferCount,
		CurrentProcessAddress:          processInfo.CurrentProcessAddress,
		NextProcessEprocessAddress:     nextProcessEProcessAddress,
		NextProcessName:                nextProcessName,
		NextProcessID:                  nextProcessID,
		PreviousProcessEprocessAddress: previousProcessEProcessAddress,
		PreviousProcessName:            previousProcessName,
		PreviousProcessID:              previousProcessID,
	})

	if err != nil {
		return 0, fmt.Errorf("failed to create process info: %w", err)
	}

	return createdProcess.ID, nil
}

func (h *WebhookHandler) IterateProcesses(c *fiber.Ctx) error {
	var req struct {
		WebhookURL string `json:"webhook_url"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.WebhookURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "webhook_url is required",
		})
	}

	// Get user ID from JWT context (if authenticated)
	var userID *int64
	if userIDVal := c.Locals("userID"); userIDVal != nil {
		if uid, ok := userIDVal.(int64); ok {
			userID = &uid
		}
	}

	// Make request to webhook
	respBody, err := h.makeHTTPRequest(req.WebhookURL+"/webhook/iterate-processes", "POST", nil)
	if err != nil {
		// If authenticated, create failed snapshot
		if userID != nil {
			userIDParam := pgtype.Int8{Int64: *userID, Valid: true}

			_, _ = h.queries.CreateProcessSnapshot(c.Context(), db.CreateProcessSnapshotParams{
				UserID:       userIDParam,
				WebhookUrl:   req.WebhookURL,
				SnapshotType: "iteration",
				ProcessCount: 0,
				Success:      false,
				ErrorMessage: pgtype.Text{String: err.Error(), Valid: true},
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to call webhook: %v", err),
		})
	}

	// Parse response
	var webhookResp IterateProcessesResponse
	if err := json.Unmarshal(respBody, &webhookResp); err != nil {
		log.Debug(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse webhook response",
		})
	}

	// If not authenticated, return processes without persisting
	if userID == nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":       "Processes iterated successfully (not persisted - no authentication)",
			"process_count": len(webhookResp.Processes),
			"processes":     webhookResp.Processes,
		})
	}

	// Authenticated: Create snapshot and persist
	userIDParam := pgtype.Int8{Int64: *userID, Valid: true}

	snapshot, err := h.queries.CreateProcessSnapshot(c.Context(), db.CreateProcessSnapshotParams{
		UserID:       userIDParam,
		WebhookUrl:   req.WebhookURL,
		SnapshotType: "iteration",
		ProcessCount: int32(len(webhookResp.Processes)),
		Success:      true,
		ErrorMessage: pgtype.Text{Valid: false},
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create snapshot",
		})
	}

	// Persist all processes to this snapshot
	processIDs := make([]int64, 0, len(webhookResp.Processes))
	for _, processInfo := range webhookResp.Processes {
		processID, err := h.persistProcessInfo(c.Context(), snapshot.ID, userID, processInfo)
		if err != nil {
			// Log error but continue with other processes
			fmt.Printf("Failed to persist process %d: %v\n", processInfo.ProcessID, err)
			continue
		}
		processIDs = append(processIDs, processID)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":       "Processes iterated and persisted successfully",
		"snapshot_id":   snapshot.ID,
		"process_count": len(processIDs),
		"processes":     webhookResp.Processes,
	})
}

func (h *WebhookHandler) ProcessByPid(c *fiber.Ctx) error {
	var req struct {
		WebhookURL string `json:"webhook_url"`
		Pid        int32  `json:"pid"`
		SnapshotID *int64 `json:"snapshot_id,omitempty"` // Optional: add to existing snapshot
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.WebhookURL == "" || req.Pid == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "webhook_url and pid are required",
		})
	}

	// Get user ID from JWT context (if authenticated)
	var userID *int64
	if userIDVal := c.Locals("userID"); userIDVal != nil {
		if uid, ok := userIDVal.(int64); ok {
			userID = &uid
		}
	}

	// Make request to webhook
	webhookReq := ProcessByPidRequest{Pid: req.Pid}
	respBody, err := h.makeHTTPRequest(req.WebhookURL, "POST", webhookReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to call webhook: %v", err),
		})
	}

	// Parse response
	var webhookResp ProcessByPidResponse
	if err := json.Unmarshal(respBody, &webhookResp); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse webhook response",
		})
	}

	// If not authenticated, return process without persisting
	if userID == nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Process queried successfully (not persisted - no authentication)",
			"process": webhookResp.Process,
		})
	}

	// Authenticated: Persist to snapshot
	userIDParam := pgtype.Int8{Int64: *userID, Valid: true}

	// Determine which snapshot to use
	var snapshotID int64
	if req.SnapshotID != nil {
		// Add to existing snapshot
		snapshotID = *req.SnapshotID

		// Verify snapshot exists and belongs to user
		snapshot, err := h.queries.GetProcessSnapshot(c.Context(), snapshotID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Snapshot not found",
			})
		}

		// Check if user owns this snapshot
		if snapshot.UserID.Valid && snapshot.UserID.Int64 != *userID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "You don't have permission to add to this snapshot",
			})
		}

		// Update snapshot process count
		err = h.queries.UpdateProcessSnapshotCount(c.Context(), db.UpdateProcessSnapshotCountParams{
			ID:           snapshotID,
			ProcessCount: snapshot.ProcessCount + 1,
		})
		if err != nil {
			fmt.Printf("Failed to update snapshot count: %v\n", err)
		}
	} else {
		// Create new snapshot for this query
		snapshot, err := h.queries.CreateProcessSnapshot(c.Context(), db.CreateProcessSnapshotParams{
			UserID:       userIDParam,
			WebhookUrl:   req.WebhookURL,
			SnapshotType: "query",
			ProcessCount: 1,
			Success:      true,
			ErrorMessage: pgtype.Text{Valid: false},
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create snapshot",
			})
		}
		snapshotID = snapshot.ID
	}

	// Persist process info
	processInfoID, err := h.persistProcessInfo(c.Context(), snapshotID, userID, webhookResp.Process)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to persist process info",
		})
	}

	// Create query history record
	_, err = h.queries.CreateProcessQuery(c.Context(), db.CreateProcessQueryParams{
		SnapshotID:    snapshotID,
		UserID:        userIDParam,
		WebhookUrl:    req.WebhookURL,
		RequestedPid:  req.Pid,
		ProcessInfoID: pgtype.Int8{Int64: processInfoID, Valid: true},
		Success:       true,
		ErrorMessage:  pgtype.Text{Valid: false},
	})

	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to create query history: %v\n", err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":         "Process queried and persisted successfully",
		"snapshot_id":     snapshotID,
		"process_info_id": processInfoID,
		"process":         webhookResp.Process,
	})
}
