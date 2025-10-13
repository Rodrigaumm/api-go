package handlers

import (
	"context"
	"strconv"
	"time"

	"go-api/internal/db"

	"github.com/gofiber/fiber/v2"
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

// ProcessInfoRequest represents the incoming process data
type ProcessInfoRequest struct {
	ProcessID             int    `json:"processId"`
	ParentProcessID       int    `json:"parentProcessId"`
	ProcessName           string `json:"processName"`
	ThreadCount           int    `json:"threadCount"`
	HandleCount           int    `json:"handleCount"`
	BasePriority          int    `json:"basePriority"`
	CreateTime            string `json:"createTime"`
	UserTime              string `json:"userTime"`
	KernelTime            string `json:"kernelTime"`
	WorkingSetSize        string `json:"workingSetSize"`
	PeakWorkingSetSize    string `json:"peakWorkingSetSize"`
	VirtualSize           string `json:"virtualSize"`
	PeakVirtualSize       string `json:"peakVirtualSize"`
	PagefileUsage         string `json:"pagefileUsage"`
	PeakPagefileUsage     string `json:"peakPagefileUsage"`
	PageFaultCount        int    `json:"pageFaultCount"`
	ReadOperationCount    int64  `json:"readOperationCount"`
	WriteOperationCount   int64  `json:"writeOperationCount"`
	OtherOperationCount   int64  `json:"otherOperationCount"`
	ReadTransferCount     int64  `json:"readTransferCount"`
	WriteTransferCount    int64  `json:"writeTransferCount"`
	OtherTransferCount    int64  `json:"otherTransferCount"`
	CurrentProcessAddress string `json:"currentProcessAddress"`
	// Adjacent processes
	NextProcessEProcessAddress     string `json:"nextProcessEProcessAddress,omitempty"`
	NextProcessName                string `json:"nextProcessName,omitempty"`
	NextProcessID                  int    `json:"nextProcessId,omitempty"`
	PreviousProcessEProcessAddress string `json:"previousProcessEProcessAddress,omitempty"`
	PreviousProcessName            string `json:"previousProcessName,omitempty"`
	PreviousProcessID              int    `json:"previousProcessId,omitempty"`
}

// AdjacentProcessResponse represents adjacent process information
type AdjacentProcessResponse struct {
	EProcessAddress string `json:"eProcessAddress,omitempty"`
	ProcessName     string `json:"processName,omitempty"`
	ProcessID       int    `json:"processId,omitempty"`
}

// ProcessInfoResponse represents the response format
type ProcessInfoResponse struct {
	ID                    int64                   `json:"id"`
	UserID                *int64                  `json:"userId,omitempty"`
	ProcessID             int                     `json:"processId"`
	ParentProcessID       int                     `json:"parentProcessId"`
	ProcessName           string                  `json:"processName"`
	ThreadCount           int                     `json:"threadCount"`
	HandleCount           int                     `json:"handleCount"`
	BasePriority          int                     `json:"basePriority"`
	CreateTime            string                  `json:"createTime"`
	UserTime              string                  `json:"userTime"`
	KernelTime            string                  `json:"kernelTime"`
	WorkingSetSize        string                  `json:"workingSetSize"`
	PeakWorkingSetSize    string                  `json:"peakWorkingSetSize"`
	VirtualSize           string                  `json:"virtualSize"`
	PeakVirtualSize       string                  `json:"peakVirtualSize"`
	PagefileUsage         string                  `json:"pagefileUsage"`
	PeakPagefileUsage     string                  `json:"peakPagefileUsage"`
	PageFaultCount        int                     `json:"pageFaultCount"`
	ReadOperationCount    int64                   `json:"readOperationCount"`
	WriteOperationCount   int64                   `json:"writeOperationCount"`
	OtherOperationCount   int64                   `json:"otherOperationCount"`
	ReadTransferCount     int64                   `json:"readTransferCount"`
	WriteTransferCount    int64                   `json:"writeTransferCount"`
	OtherTransferCount    int64                   `json:"otherTransferCount"`
	CurrentProcessAddress string                  `json:"currentProcessAddress"`
	NextProcess           AdjacentProcessResponse `json:"nextProcess"`
	PreviousProcess       AdjacentProcessResponse `json:"previousProcess"`
	CreatedAt             time.Time               `json:"createdAt"`
	UpdatedAt             time.Time               `json:"updatedAt"`
}

// IterationHistoryResponse represents iteration history
type IterationHistoryResponse struct {
	ID           int64     `json:"id"`
	UserID       *int64    `json:"userId,omitempty"`
	WebhookURL   string    `json:"webhookUrl"`
	ProcessCount int       `json:"processCount"`
	Success      bool      `json:"success"`
	ErrorMessage string    `json:"errorMessage,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

// QueryHistoryResponse represents query history
type QueryHistoryResponse struct {
	ID           int64                `json:"id"`
	UserID       *int64               `json:"userId,omitempty"`
	WebhookURL   string               `json:"webhookUrl"`
	RequestedPID int                  `json:"requestedPid"`
	ProcessInfo  *ProcessInfoResponse `json:"processInfo,omitempty"`
	Success      bool                 `json:"success"`
	ErrorMessage string               `json:"errorMessage,omitempty"`
	CreatedAt    time.Time            `json:"createdAt"`
}

func (h *ProcessHandler) CreateProcessInfo(c *fiber.Ctx) error {
	// Get user from context (set by JWT middleware)
	userID, _, ok := GetUserFromContext(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var req ProcessInfoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
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

	var userIDParam pgtype.Int8
	userIDParam = pgtype.Int8{Int64: userID, Valid: true}

	// Create process info
	processInfo, err := h.queries.CreateProcessInfo(context.Background(), db.CreateProcessInfoParams{
		UserID:                         userIDParam,
		ProcessID:                      int32(req.ProcessID),
		ParentProcessID:                int32(req.ParentProcessID),
		ProcessName:                    req.ProcessName,
		ThreadCount:                    int32(req.ThreadCount),
		HandleCount:                    int32(req.HandleCount),
		BasePriority:                   int32(req.BasePriority),
		CreateTime:                     req.CreateTime,
		UserTime:                       req.UserTime,
		KernelTime:                     req.KernelTime,
		WorkingSetSize:                 req.WorkingSetSize,
		PeakWorkingSetSize:             req.PeakWorkingSetSize,
		VirtualSize:                    req.VirtualSize,
		PeakVirtualSize:                req.PeakVirtualSize,
		PagefileUsage:                  req.PagefileUsage,
		PeakPagefileUsage:              req.PeakPagefileUsage,
		PageFaultCount:                 int32(req.PageFaultCount),
		ReadOperationCount:             req.ReadOperationCount,
		WriteOperationCount:            req.WriteOperationCount,
		OtherOperationCount:            req.OtherOperationCount,
		ReadTransferCount:              req.ReadTransferCount,
		WriteTransferCount:             req.WriteTransferCount,
		OtherTransferCount:             req.OtherTransferCount,
		CurrentProcessAddress:          req.CurrentProcessAddress,
		NextProcessEprocessAddress:     toText(req.NextProcessEProcessAddress),
		NextProcessName:                toText(req.NextProcessName),
		NextProcessID:                  toInt4(req.NextProcessID),
		PreviousProcessEprocessAddress: toText(req.PreviousProcessEProcessAddress),
		PreviousProcessName:            toText(req.PreviousProcessName),
		PreviousProcessID:              toInt4(req.PreviousProcessID),
	})

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create process info",
		})
	}

	return c.Status(201).JSON(toProcessInfoResponse(processInfo))
}

func (h *ProcessHandler) GetProcessInfos(c *fiber.Ctx) error {
	// Get user from context (set by JWT middleware)
	userID, _, ok := GetUserFromContext(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var userIDParam pgtype.Int8
	userIDParam = pgtype.Int8{Int64: userID, Valid: true}

	processInfos, err := h.queries.GetProcessInfosByUser(context.Background(), userIDParam)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve process infos",
		})
	}

	var response []ProcessInfoResponse
	for _, info := range processInfos {
		response = append(response, toProcessInfoResponse(info))
	}

	return c.JSON(response)
}

func (h *ProcessHandler) GetProcessInfo(c *fiber.Ctx) error {
	// Get user from context (set by JWT middleware)
	userID, _, ok := GetUserFromContext(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	var userIDParam pgtype.Int8
	userIDParam = pgtype.Int8{Int64: userID, Valid: true}

	processInfo, err := h.queries.GetProcessInfo(context.Background(), db.GetProcessInfoParams{
		ID:     id,
		UserID: userIDParam,
	})
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Process info not found",
		})
	}

	return c.JSON(toProcessInfoResponse(processInfo))
}

func (h *ProcessHandler) GetProcessInfosByProcessID(c *fiber.Ctx) error {
	// Get user from context (set by JWT middleware)
	userID, _, ok := GetUserFromContext(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	processID, err := strconv.ParseInt(c.Params("processId"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid process ID",
		})
	}

	var userIDParam pgtype.Int8
	userIDParam = pgtype.Int8{Int64: userID, Valid: true}

	processInfos, err := h.queries.GetProcessInfosByProcessID(context.Background(), db.GetProcessInfosByProcessIDParams{
		UserID:    userIDParam,
		ProcessID: int32(processID),
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve process infos",
		})
	}

	var response []ProcessInfoResponse
	for _, info := range processInfos {
		response = append(response, toProcessInfoResponse(info))
	}

	return c.JSON(response)
}

func (h *ProcessHandler) DeleteProcessInfo(c *fiber.Ctx) error {
	// Get user from context (set by JWT middleware)
	userID, _, ok := GetUserFromContext(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	var userIDParam pgtype.Int8
	userIDParam = pgtype.Int8{Int64: userID, Valid: true}

	// First check if the process info exists and belongs to the user
	_, err = h.queries.GetProcessInfo(context.Background(), db.GetProcessInfoParams{
		ID:     id,
		UserID: userIDParam,
	})
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Process info not found",
		})
	}

	// Delete the process info
	err = h.queries.DeleteProcessInfo(context.Background(), db.DeleteProcessInfoParams{
		ID:     id,
		UserID: userIDParam,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete process info",
		})
	}

	return c.Status(204).Send(nil)
}

// GetIterationHistory returns the iteration history for the authenticated user
func (h *ProcessHandler) GetIterationHistory(c *fiber.Ctx) error {
	userID, _, ok := GetUserFromContext(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var userIDParam pgtype.Int8
	userIDParam = pgtype.Int8{Int64: userID, Valid: true}

	history, err := h.queries.GetProcessIterationHistoryByUser(context.Background(), userIDParam)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve iteration history",
		})
	}

	var response []IterationHistoryResponse
	for _, h := range history {
		response = append(response, toIterationHistoryResponse(h))
	}

	return c.JSON(response)
}

// GetIterationProcesses returns all processes for a specific iteration
func (h *ProcessHandler) GetIterationProcesses(c *fiber.Ctx) error {
	userID, _, ok := GetUserFromContext(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	iterationID, err := strconv.ParseInt(c.Params("iterationId"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid iteration ID",
		})
	}

	// First verify the iteration belongs to the user
	iteration, err := h.queries.GetProcessIterationHistory(context.Background(), iterationID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Iteration not found",
		})
	}

	if iteration.UserID.Valid && iteration.UserID.Int64 != userID {
		return c.Status(403).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	// Get all processes for this iteration
	processes, err := h.queries.GetProcessesByIterationID(context.Background(), iterationID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve processes",
		})
	}

	var response []ProcessInfoResponse
	for _, info := range processes {
		response = append(response, toProcessInfoResponse(info))
	}

	return c.JSON(fiber.Map{
		"iteration": toIterationHistoryResponse(iteration),
		"processes": response,
	})
}

// GetQueryHistory returns the query history for the authenticated user
func (h *ProcessHandler) GetQueryHistory(c *fiber.Ctx) error {
	userID, _, ok := GetUserFromContext(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var userIDParam pgtype.Int8
	userIDParam = pgtype.Int8{Int64: userID, Valid: true}

	history, err := h.queries.GetProcessQueryHistoryByUser(context.Background(), userIDParam)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve query history",
		})
	}

	var response []QueryHistoryResponse
	for _, hist := range history {
		resp := toQueryHistoryResponse(hist)

		// If there's a process info ID, fetch the process info
		if hist.ProcessInfoID.Valid {
			processInfo, err := h.queries.GetProcessInfoByID(context.Background(), hist.ProcessInfoID.Int64)
			if err == nil {
				processResp := toProcessInfoResponse(processInfo)
				resp.ProcessInfo = &processResp
			}
		}

		response = append(response, resp)
	}

	return c.JSON(response)
}

// GetStatistics returns statistics for the authenticated user
func (h *ProcessHandler) GetStatistics(c *fiber.Ctx) error {
	userID, _, ok := GetUserFromContext(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var userIDParam pgtype.Int8
	userIDParam = pgtype.Int8{Int64: userID, Valid: true}

	processCount, _ := h.queries.GetUserProcessCount(context.Background(), userIDParam)
	iterationCount, _ := h.queries.GetUserIterationCount(context.Background(), userIDParam)
	queryCount, _ := h.queries.GetUserQueryCount(context.Background(), userIDParam)

	mostQueried, _ := h.queries.GetMostQueriedProcesses(context.Background(), db.GetMostQueriedProcessesParams{
		UserID: userIDParam,
		Limit:  10,
	})

	return c.JSON(fiber.Map{
		"totalProcesses":       processCount,
		"totalIterations":      iterationCount,
		"totalQueries":         queryCount,
		"mostQueriedProcesses": mostQueried,
	})
}

// Helper function to convert ProcessInfo to ProcessInfoResponse
func toProcessInfoResponse(info db.ProcessInfo) ProcessInfoResponse {
	var userID *int64
	if info.UserID.Valid {
		userID = &info.UserID.Int64
	}

	var nextProcess AdjacentProcessResponse
	if info.NextProcessEprocessAddress.Valid {
		nextProcess.EProcessAddress = info.NextProcessEprocessAddress.String
	}
	if info.NextProcessName.Valid {
		nextProcess.ProcessName = info.NextProcessName.String
	}
	if info.NextProcessID.Valid {
		nextProcess.ProcessID = int(info.NextProcessID.Int32)
	}

	var previousProcess AdjacentProcessResponse
	if info.PreviousProcessEprocessAddress.Valid {
		previousProcess.EProcessAddress = info.PreviousProcessEprocessAddress.String
	}
	if info.PreviousProcessName.Valid {
		previousProcess.ProcessName = info.PreviousProcessName.String
	}
	if info.PreviousProcessID.Valid {
		previousProcess.ProcessID = int(info.PreviousProcessID.Int32)
	}

	var createdAt, updatedAt time.Time
	if info.CreatedAt.Valid {
		createdAt = info.CreatedAt.Time
	}
	if info.UpdatedAt.Valid {
		updatedAt = info.UpdatedAt.Time
	}

	return ProcessInfoResponse{
		ID:                    info.ID,
		UserID:                userID,
		ProcessID:             int(info.ProcessID),
		ParentProcessID:       int(info.ParentProcessID),
		ProcessName:           info.ProcessName,
		ThreadCount:           int(info.ThreadCount),
		HandleCount:           int(info.HandleCount),
		BasePriority:          int(info.BasePriority),
		CreateTime:            info.CreateTime,
		UserTime:              info.UserTime,
		KernelTime:            info.KernelTime,
		WorkingSetSize:        info.WorkingSetSize,
		PeakWorkingSetSize:    info.PeakWorkingSetSize,
		VirtualSize:           info.VirtualSize,
		PeakVirtualSize:       info.PeakVirtualSize,
		PagefileUsage:         info.PagefileUsage,
		PeakPagefileUsage:     info.PeakPagefileUsage,
		PageFaultCount:        int(info.PageFaultCount),
		ReadOperationCount:    info.ReadOperationCount,
		WriteOperationCount:   info.WriteOperationCount,
		OtherOperationCount:   info.OtherOperationCount,
		ReadTransferCount:     info.ReadTransferCount,
		WriteTransferCount:    info.WriteTransferCount,
		OtherTransferCount:    info.OtherTransferCount,
		CurrentProcessAddress: info.CurrentProcessAddress,
		NextProcess:           nextProcess,
		PreviousProcess:       previousProcess,
		CreatedAt:             createdAt,
		UpdatedAt:             updatedAt,
	}
}

// Helper function to convert iteration history to response
func toIterationHistoryResponse(h db.ProcessIterationHistory) IterationHistoryResponse {
	var userID *int64
	if h.UserID.Valid {
		userID = &h.UserID.Int64
	}

	var errorMsg string
	if h.ErrorMessage.Valid {
		errorMsg = h.ErrorMessage.String
	}

	var createdAt time.Time
	if h.CreatedAt.Valid {
		createdAt = h.CreatedAt.Time
	}

	return IterationHistoryResponse{
		ID:           h.ID,
		UserID:       userID,
		WebhookURL:   h.WebhookUrl,
		ProcessCount: int(h.ProcessCount),
		Success:      h.Success,
		ErrorMessage: errorMsg,
		CreatedAt:    createdAt,
	}
}

// Helper function to convert query history to response
func toQueryHistoryResponse(h db.ProcessQueryHistory) QueryHistoryResponse {
	var userID *int64
	if h.UserID.Valid {
		userID = &h.UserID.Int64
	}

	var errorMsg string
	if h.ErrorMessage.Valid {
		errorMsg = h.ErrorMessage.String
	}

	var createdAt time.Time
	if h.CreatedAt.Valid {
		createdAt = h.CreatedAt.Time
	}

	return QueryHistoryResponse{
		ID:           h.ID,
		UserID:       userID,
		WebhookURL:   h.WebhookUrl,
		RequestedPID: int(h.RequestedPid),
		Success:      h.Success,
		ErrorMessage: errorMsg,
		CreatedAt:    createdAt,
	}
}
