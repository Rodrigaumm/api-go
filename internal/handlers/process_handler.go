package handlers

import (
	"strconv"

	"go-api/internal/db"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
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
	EProcessAddress *string `json:"eprocessAddress,omitempty"`
	ProcessName     *string `json:"processMame,omitempty"`
	ProcessID       *int64  `json:"processId,omitempty"`
	ID              *int64  `json:"id,omitempty"`
}

type ProcessInfoResponse struct {
	ID                    int64                    `json:"id"`
	SnapshotID            int64                    `json:"snapshotId"`
	UserID                *int64                   `json:"userId,omitempty"`
	ProcessID             int64                    `json:"processId"`
	ParentProcessID       int64                    `json:"parentProcessId"`
	ProcessName           string                   `json:"processName"`
	ThreadCount           int32                    `json:"threadCount"`
	HandleCount           int32                    `json:"handleCount"`
	BasePriority          int32                    `json:"basePriority"`
	CreateTime            string                   `json:"createTime"`
	UserTime              int32                    `json:"userTime"`
	KernelTime            int32                    `json:"kernelTime"`
	WorkingSetSize        int64                    `json:"workingSetSize"`
	PeakWorkingSetSize    int64                    `json:"peakWorkingSetSize"`
	VirtualSize           int64                    `json:"virtualSize"`
	PeakVirtualSize       int64                    `json:"peakVirtualSize"`
	ReadOperationCount    int64                    `json:"readOperationCount"`
	WriteOperationCount   int64                    `json:"writeOperationCount"`
	OtherOperationCount   int64                    `json:"otherOperationCount"`
	ReadTransferCount     int64                    `json:"readTransferCount"`
	WriteTransferCount    int64                    `json:"writeTransferCount"`
	OtherTransferCount    int64                    `json:"otherTransferCount"`
	CurrentProcessAddress string                   `json:"currentProcessAddress"`
	NextProcess           *AdjacentProcessResponse `json:"nextProcess,omitempty"`
	PreviousProcess       *AdjacentProcessResponse `json:"previousProcess,omitempty"`
	CreatedAt             string                   `json:"createdAt"`
	UpdatedAt             string                   `json:"updatedAt"`
}

type SnapshotResponse struct {
	ID           int64   `json:"id"`
	UserID       *int64  `json:"userId,omitempty"`
	WebhookURL   string  `json:"webhook_url"`
	SnapshotType string  `json:"snapshotType"`
	ProcessCount int32   `json:"processCount"`
	Success      bool    `json:"success"`
	ErrorMessage *string `json:"errorMessage,omitempty"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

type QueryHistoryResponse struct {
	ID            int64   `json:"id"`
	SnapshotID    int64   `json:"snapshotId"`
	UserID        *int64  `json:"userId,omitempty"`
	WebhookURL    string  `json:"webhookUrl"`
	RequestedPID  int32   `json:"requestedPid"`
	ProcessInfoID *int64  `json:"processInfoId,omitempty"`
	Success       bool    `json:"success"`
	ErrorMessage  *string `json:"errorMessage,omitempty"`
	CreatedAt     string  `json:"createdAt"`
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
	log.Debug("entrou no getProcessInfo")
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

	processResponse := toProcessInfoResponse(processInfo)
	log.Debug(processResponse.NextProcess.ID)
	log.Debug(processResponse.NextProcess.ProcessID)
	log.Debug(processResponse.NextProcess.ProcessName)
	return c.JSON(processResponse)
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
		ProcessID: int64(processID),
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
		log.Debug(err)
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
			nextProcess.ProcessID = &info.NextProcessID.Int64
		}
		if info.NextID.Valid {
			nextProcess.ID = &info.NextID.Int64
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
			previousProcess.ProcessID = &info.PreviousProcessID.Int64
		}
		if info.PreviousID.Valid {
			previousProcess.ID = &info.PreviousID.Int64
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
