package handlers

import (
	"strconv"

	"go-api/internal/db"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProcessHandler struct {
	queries *db.Queries
}

func NewProcessHandler(dbpool *pgxpool.Pool) *ProcessHandler {
	return &ProcessHandler{
		queries: db.New(dbpool),
	}
}

// Response structures
type AdjacentProcessResponse struct {
	EProcessAddress *string `json:"eprocess_address,omitempty"`
	ProcessName     *string `json:"process_name,omitempty"`
	ProcessID       *int32  `json:"process_id,omitempty"`
}

type ProcessInfoResponse struct {
	ID                    int64                    `json:"id"`
	SnapshotID            int64                    `json:"snapshot_id"`
	UserID                *int64                   `json:"user_id,omitempty"`
	ProcessID             int32                    `json:"process_id"`
	ParentProcessID       int32                    `json:"parent_process_id"`
	ProcessName           string                   `json:"process_name"`
	ThreadCount           int32                    `json:"thread_count"`
	HandleCount           int32                    `json:"handle_count"`
	BasePriority          int32                    `json:"base_priority"`
	CreateTime            string                   `json:"create_time"`
	UserTime              string                   `json:"user_time"`
	KernelTime            string                   `json:"kernel_time"`
	WorkingSetSize        string                   `json:"working_set_size"`
	PeakWorkingSetSize    string                   `json:"peak_working_set_size"`
	VirtualSize           string                   `json:"virtual_size"`
	PeakVirtualSize       string                   `json:"peak_virtual_size"`
	PagefileUsage         string                   `json:"pagefile_usage"`
	PeakPagefileUsage     string                   `json:"peak_pagefile_usage"`
	PageFaultCount        int32                    `json:"page_fault_count"`
	ReadOperationCount    int64                    `json:"read_operation_count"`
	WriteOperationCount   int64                    `json:"write_operation_count"`
	OtherOperationCount   int64                    `json:"other_operation_count"`
	ReadTransferCount     int64                    `json:"read_transfer_count"`
	WriteTransferCount    int64                    `json:"write_transfer_count"`
	OtherTransferCount    int64                    `json:"other_transfer_count"`
	CurrentProcessAddress string                   `json:"current_process_address"`
	NextProcess           *AdjacentProcessResponse `json:"next_process,omitempty"`
	PreviousProcess       *AdjacentProcessResponse `json:"previous_process,omitempty"`
	CreatedAt             string                   `json:"created_at"`
	UpdatedAt             string                   `json:"updated_at"`
}

type SnapshotResponse struct {
	ID           int64   `json:"id"`
	UserID       *int64  `json:"user_id,omitempty"`
	WebhookURL   string  `json:"webhook_url"`
	SnapshotType string  `json:"snapshot_type"`
	ProcessCount int32   `json:"process_count"`
	Success      bool    `json:"success"`
	ErrorMessage *string `json:"error_message,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type QueryHistoryResponse struct {
	ID            int64   `json:"id"`
	SnapshotID    int64   `json:"snapshot_id"`
	UserID        *int64  `json:"user_id,omitempty"`
	WebhookURL    string  `json:"webhook_url"`
	RequestedPID  int32   `json:"requested_pid"`
	ProcessInfoID *int64  `json:"process_info_id,omitempty"`
	Success       bool    `json:"success"`
	ErrorMessage  *string `json:"error_message,omitempty"`
	CreatedAt     string  `json:"created_at"`
}

// Get all snapshots for a user
func (h *ProcessHandler) GetSnapshots(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)

	snapshots, err := h.queries.GetProcessSnapshotsByUser(c.Context(), pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch snapshots",
		})
	}

	response := make([]SnapshotResponse, len(snapshots))
	for i, snapshot := range snapshots {
		response[i] = toSnapshotResponse(snapshot)
	}

	return c.JSON(response)
}

// Get a specific snapshot by ID
func (h *ProcessHandler) GetSnapshot(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid snapshot ID",
		})
	}

	snapshot, err := h.queries.GetProcessSnapshot(c.Context(), id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Snapshot not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch snapshot",
		})
	}

	// Check if user has access to this snapshot
	if snapshot.UserID.Valid && snapshot.UserID.Int64 != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	return c.JSON(toSnapshotResponse(snapshot))
}

// Get all processes in a snapshot
func (h *ProcessHandler) GetSnapshotProcesses(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)

	snapshotID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid snapshot ID",
		})
	}

	// Verify snapshot exists and user has access
	snapshot, err := h.queries.GetProcessSnapshot(c.Context(), snapshotID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Snapshot not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch snapshot",
		})
	}

	if snapshot.UserID.Valid && snapshot.UserID.Int64 != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	// Get all processes in this snapshot
	processes, err := h.queries.GetProcessInfosBySnapshot(c.Context(), snapshotID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch processes",
		})
	}

	response := make([]ProcessInfoResponse, len(processes))
	for i, process := range processes {
		response[i] = toProcessInfoResponse(process)
	}

	return c.JSON(fiber.Map{
		"snapshot":  toSnapshotResponse(snapshot),
		"processes": response,
	})
}

// Get snapshots by type (iteration or query)
func (h *ProcessHandler) GetSnapshotsByType(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	snapshotType := c.Params("type")

	if snapshotType != "iteration" && snapshotType != "query" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid snapshot type. Must be 'iteration' or 'query'",
		})
	}

	snapshots, err := h.queries.GetProcessSnapshotsByType(c.Context(), db.GetProcessSnapshotsByTypeParams{
		UserID:       pgtype.Int8{Int64: userID, Valid: true},
		SnapshotType: snapshotType,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch snapshots",
		})
	}

	response := make([]SnapshotResponse, len(snapshots))
	for i, snapshot := range snapshots {
		response[i] = toSnapshotResponse(snapshot)
	}

	return c.JSON(response)
}

// Get a specific process info by ID
func (h *ProcessHandler) GetProcessInfo(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid process info ID",
		})
	}

	processInfo, err := h.queries.GetProcessInfo(c.Context(), id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Process info not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch process info",
		})
	}

	// Check if user has access
	if processInfo.UserID.Valid && processInfo.UserID.Int64 != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	return c.JSON(toProcessInfoResponse(processInfo))
}

// Get all processes for a user
func (h *ProcessHandler) GetProcessInfos(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)

	processes, err := h.queries.GetProcessInfosByUser(c.Context(), pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch processes",
		})
	}

	response := make([]ProcessInfoResponse, len(processes))
	for i, process := range processes {
		response[i] = toProcessInfoResponse(process)
	}

	return c.JSON(response)
}

// Get all processes by process ID (across all snapshots)
func (h *ProcessHandler) GetProcessInfosByProcessID(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)

	processID, err := strconv.ParseInt(c.Params("pid"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid process ID",
		})
	}

	processes, err := h.queries.GetProcessInfosByProcessID(c.Context(), db.GetProcessInfosByProcessIDParams{
		UserID:    pgtype.Int8{Int64: userID, Valid: true},
		ProcessID: int32(processID),
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch processes",
		})
	}

	response := make([]ProcessInfoResponse, len(processes))
	for i, process := range processes {
		response[i] = toProcessInfoResponse(process)
	}

	return c.JSON(response)
}

// Delete a process info
func (h *ProcessHandler) DeleteProcessInfo(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid process info ID",
		})
	}

	// Check if process exists and user has access
	processInfo, err := h.queries.GetProcessInfo(c.Context(), id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Process info not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch process info",
		})
	}

	if processInfo.UserID.Valid && processInfo.UserID.Int64 != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	err = h.queries.DeleteProcessInfo(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete process info",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Process info deleted successfully",
	})
}

// Delete a snapshot (and all its processes)
func (h *ProcessHandler) DeleteSnapshot(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid snapshot ID",
		})
	}

	// Check if snapshot exists and user has access
	snapshot, err := h.queries.GetProcessSnapshot(c.Context(), id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Snapshot not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch snapshot",
		})
	}

	if snapshot.UserID.Valid && snapshot.UserID.Int64 != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	err = h.queries.DeleteProcessSnapshot(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete snapshot",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Snapshot deleted successfully",
	})
}

// Get query history for a user
func (h *ProcessHandler) GetQueryHistory(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)

	queries, err := h.queries.GetProcessQueriesByUser(c.Context(), pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch query history",
		})
	}

	response := make([]QueryHistoryResponse, len(queries))
	for i, query := range queries {
		response[i] = toQueryHistoryResponse(query)
	}

	return c.JSON(response)
}

// Get query history for a specific snapshot
func (h *ProcessHandler) GetSnapshotQueries(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)

	snapshotID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid snapshot ID",
		})
	}

	// Verify snapshot exists and user has access
	snapshot, err := h.queries.GetProcessSnapshot(c.Context(), snapshotID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Snapshot not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch snapshot",
		})
	}

	if snapshot.UserID.Valid && snapshot.UserID.Int64 != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	queries, err := h.queries.GetProcessQueriesBySnapshot(c.Context(), snapshotID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch queries",
		})
	}

	response := make([]QueryHistoryResponse, len(queries))
	for i, query := range queries {
		response[i] = toQueryHistoryResponse(query)
	}

	return c.JSON(response)
}

// Get statistics for a user
func (h *ProcessHandler) GetStatistics(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)

	processCount, err := h.queries.CountUserProcesses(c.Context(), pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch process count",
		})
	}

	snapshotCount, err := h.queries.CountUserSnapshots(c.Context(), pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch snapshot count",
		})
	}

	queryCount, err := h.queries.CountUserQueries(c.Context(), pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch query count",
		})
	}

	mostQueried, err := h.queries.GetMostQueriedProcesses(c.Context(), db.GetMostQueriedProcessesParams{
		UserID: pgtype.Int8{Int64: userID, Valid: true},
		Limit:  10,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch most queried processes",
		})
	}

	snapshotStats, err := h.queries.GetSnapshotStatistics(c.Context(), pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch snapshot statistics",
		})
	}

	return c.JSON(fiber.Map{
		"total_processes":        processCount,
		"total_snapshots":        snapshotCount,
		"total_queries":          queryCount,
		"most_queried_processes": mostQueried,
		"snapshot_statistics":    snapshotStats,
	})
}

// Helper functions
func toProcessInfoResponse(info db.ProcessInfo) ProcessInfoResponse {
	response := ProcessInfoResponse{
		ID:                    info.ID,
		SnapshotID:            info.SnapshotID,
		ProcessID:             info.ProcessID,
		ParentProcessID:       info.ParentProcessID,
		ProcessName:           info.ProcessName,
		ThreadCount:           info.ThreadCount,
		HandleCount:           info.HandleCount,
		BasePriority:          info.BasePriority,
		CreateTime:            info.CreateTime,
		UserTime:              info.UserTime,
		KernelTime:            info.KernelTime,
		WorkingSetSize:        info.WorkingSetSize,
		PeakWorkingSetSize:    info.PeakWorkingSetSize,
		VirtualSize:           info.VirtualSize,
		PeakVirtualSize:       info.PeakVirtualSize,
		PagefileUsage:         info.PagefileUsage,
		PeakPagefileUsage:     info.PeakPagefileUsage,
		PageFaultCount:        info.PageFaultCount,
		ReadOperationCount:    info.ReadOperationCount,
		WriteOperationCount:   info.WriteOperationCount,
		OtherOperationCount:   info.OtherOperationCount,
		ReadTransferCount:     info.ReadTransferCount,
		WriteTransferCount:    info.WriteTransferCount,
		OtherTransferCount:    info.OtherTransferCount,
		CurrentProcessAddress: info.CurrentProcessAddress,
		CreatedAt:             info.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:             info.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}

	if info.UserID.Valid {
		response.UserID = &info.UserID.Int64
	}

	if info.NextProcessEprocessAddress.Valid || info.NextProcessName.Valid || info.NextProcessID.Valid {
		nextProcess := &AdjacentProcessResponse{}
		if info.NextProcessEprocessAddress.Valid {
			nextProcess.EProcessAddress = &info.NextProcessEprocessAddress.String
		}
		if info.NextProcessName.Valid {
			nextProcess.ProcessName = &info.NextProcessName.String
		}
		if info.NextProcessID.Valid {
			nextProcess.ProcessID = &info.NextProcessID.Int32
		}
		response.NextProcess = nextProcess
	}

	if info.PreviousProcessEprocessAddress.Valid || info.PreviousProcessName.Valid || info.PreviousProcessID.Valid {
		previousProcess := &AdjacentProcessResponse{}
		if info.PreviousProcessEprocessAddress.Valid {
			previousProcess.EProcessAddress = &info.PreviousProcessEprocessAddress.String
		}
		if info.PreviousProcessName.Valid {
			previousProcess.ProcessName = &info.PreviousProcessName.String
		}
		if info.PreviousProcessID.Valid {
			previousProcess.ProcessID = &info.PreviousProcessID.Int32
		}
		response.PreviousProcess = previousProcess
	}

	return response
}

func toSnapshotResponse(snapshot db.ProcessSnapshot) SnapshotResponse {
	response := SnapshotResponse{
		ID:           snapshot.ID,
		WebhookURL:   snapshot.WebhookUrl,
		SnapshotType: snapshot.SnapshotType,
		ProcessCount: snapshot.ProcessCount,
		Success:      snapshot.Success,
		CreatedAt:    snapshot.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    snapshot.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}

	if snapshot.UserID.Valid {
		response.UserID = &snapshot.UserID.Int64
	}

	if snapshot.ErrorMessage.Valid {
		response.ErrorMessage = &snapshot.ErrorMessage.String
	}

	return response
}

func toQueryHistoryResponse(query db.ProcessQuery) QueryHistoryResponse {
	response := QueryHistoryResponse{
		ID:           query.ID,
		SnapshotID:   query.SnapshotID,
		WebhookURL:   query.WebhookUrl,
		RequestedPID: query.RequestedPid,
		Success:      query.Success,
		CreatedAt:    query.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}

	if query.UserID.Valid {
		response.UserID = &query.UserID.Int64
	}

	if query.ProcessInfoID.Valid {
		response.ProcessInfoID = &query.ProcessInfoID.Int64
	}

	if query.ErrorMessage.Valid {
		response.ErrorMessage = &query.ErrorMessage.String
	}

	return response
}
