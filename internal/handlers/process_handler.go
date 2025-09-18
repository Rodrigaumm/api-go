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

// ProcessInfoRequest represents the incoming process data from Windows client
type ProcessInfoRequest struct {
	ProcessID              int64  `json:"process_id"`
	ParentProcessID        int64  `json:"parent_process_id"`
	ProcessName            string `json:"process_name"`
	ThreadCount            int64  `json:"thread_count"`
	HandleCount            int64  `json:"handle_count"`
	BasePriority           int32  `json:"base_priority"`
	CreateTime             string `json:"create_time"` // ISO 8601 format
	UserTime               int64  `json:"user_time"`
	KernelTime             int64  `json:"kernel_time"`
	WorkingSetSize         int64  `json:"working_set_size"`
	PeakWorkingSetSize     int64  `json:"peak_working_set_size"`
	VirtualSize            int64  `json:"virtual_size"`
	PeakVirtualSize        int64  `json:"peak_virtual_size"`
	PagefileUsage          int64  `json:"pagefile_usage"`
	PeakPagefileUsage      int64  `json:"peak_pagefile_usage"`
	PageFaultCount         int64  `json:"page_fault_count"`
	ReadOperationCount     int64  `json:"read_operation_count"`
	WriteOperationCount    int64  `json:"write_operation_count"`
	OtherOperationCount    int64  `json:"other_operation_count"`
	ReadTransferCount      int64  `json:"read_transfer_count"`
	WriteTransferCount     int64  `json:"write_transfer_count"`
	OtherTransferCount     int64  `json:"other_transfer_count"`
	CurrentProcessAddress  string `json:"current_process_address"`
	NextProcessAddress     string `json:"next_process_address"`
	PreviousProcessAddress string `json:"previous_process_address"`
}

// ProcessInfoResponse represents the response format
type ProcessInfoResponse struct {
	ID                     int64     `json:"id"`
	UserID                 int64     `json:"user_id"`
	ProcessID              int64     `json:"process_id"`
	ParentProcessID        int64     `json:"parent_process_id"`
	ProcessName            string    `json:"process_name"`
	ThreadCount            int64     `json:"thread_count"`
	HandleCount            int64     `json:"handle_count"`
	BasePriority           int32     `json:"base_priority"`
	CreateTime             time.Time `json:"create_time"`
	UserTime               int64     `json:"user_time"`
	KernelTime             int64     `json:"kernel_time"`
	WorkingSetSize         int64     `json:"working_set_size"`
	PeakWorkingSetSize     int64     `json:"peak_working_set_size"`
	VirtualSize            int64     `json:"virtual_size"`
	PeakVirtualSize        int64     `json:"peak_virtual_size"`
	PagefileUsage          int64     `json:"pagefile_usage"`
	PeakPagefileUsage      int64     `json:"peak_pagefile_usage"`
	PageFaultCount         int64     `json:"page_fault_count"`
	ReadOperationCount     int64     `json:"read_operation_count"`
	WriteOperationCount    int64     `json:"write_operation_count"`
	OtherOperationCount    int64     `json:"other_operation_count"`
	ReadTransferCount      int64     `json:"read_transfer_count"`
	WriteTransferCount     int64     `json:"write_transfer_count"`
	OtherTransferCount     int64     `json:"other_transfer_count"`
	CurrentProcessAddress  string    `json:"current_process_address"`
	NextProcessAddress     string    `json:"next_process_address"`
	PreviousProcessAddress string    `json:"previous_process_address"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
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

	// Parse create time
	createTime, err := time.Parse(time.RFC3339, req.CreateTime)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid create_time format. Use ISO 8601 format",
		})
	}

	// Convert createTime to pgtype.Timestamp
	var createTimePg pgtype.Timestamp
	if err := createTimePg.Scan(createTime); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid create_time format",
		})
	}

	// Convert address strings to pgtype.Text
	var currentAddr, nextAddr, prevAddr pgtype.Text
	if req.CurrentProcessAddress != "" {
		currentAddr = pgtype.Text{String: req.CurrentProcessAddress, Valid: true}
	}
	if req.NextProcessAddress != "" {
		nextAddr = pgtype.Text{String: req.NextProcessAddress, Valid: true}
	}
	if req.PreviousProcessAddress != "" {
		prevAddr = pgtype.Text{String: req.PreviousProcessAddress, Valid: true}
	}

	// Create process info
	processInfo, err := h.queries.CreateProcessInfo(context.Background(), db.CreateProcessInfoParams{
		UserID:                 userID,
		ProcessID:              req.ProcessID,
		ParentProcessID:        req.ParentProcessID,
		ProcessName:            req.ProcessName,
		ThreadCount:            req.ThreadCount,
		HandleCount:            req.HandleCount,
		BasePriority:           req.BasePriority,
		CreateTime:             createTimePg,
		UserTime:               req.UserTime,
		KernelTime:             req.KernelTime,
		WorkingSetSize:         req.WorkingSetSize,
		PeakWorkingSetSize:     req.PeakWorkingSetSize,
		VirtualSize:            req.VirtualSize,
		PeakVirtualSize:        req.PeakVirtualSize,
		PagefileUsage:          req.PagefileUsage,
		PeakPagefileUsage:      req.PeakPagefileUsage,
		PageFaultCount:         req.PageFaultCount,
		ReadOperationCount:     req.ReadOperationCount,
		WriteOperationCount:    req.WriteOperationCount,
		OtherOperationCount:    req.OtherOperationCount,
		ReadTransferCount:      req.ReadTransferCount,
		WriteTransferCount:     req.WriteTransferCount,
		OtherTransferCount:     req.OtherTransferCount,
		CurrentProcessAddress:  currentAddr,
		NextProcessAddress:     nextAddr,
		PreviousProcessAddress: prevAddr,
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

	processInfos, err := h.queries.GetProcessInfosByUser(context.Background(), userID)
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

	processInfo, err := h.queries.GetProcessInfo(context.Background(), db.GetProcessInfoParams{
		ID:     id,
		UserID: userID,
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

	processID, err := strconv.ParseInt(c.Params("processId"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid process ID",
		})
	}

	processInfos, err := h.queries.GetProcessInfosByProcessId(context.Background(), db.GetProcessInfosByProcessIdParams{
		UserID:    userID,
		ProcessID: processID,
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

	// First check if the process info exists and belongs to the user
	_, err = h.queries.GetProcessInfo(context.Background(), db.GetProcessInfoParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Process info not found",
		})
	}

	// Delete the process info
	err = h.queries.DeleteProcessInfo(context.Background(), db.DeleteProcessInfoParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete process info",
		})
	}

	return c.Status(204).Send(nil)
}

// Helper function to convert ProcessInfo to ProcessInfoResponse
func toProcessInfoResponse(info db.ProcessInfo) ProcessInfoResponse {
	var currentAddr, nextAddr, prevAddr string
	if info.CurrentProcessAddress.Valid {
		currentAddr = info.CurrentProcessAddress.String
	}
	if info.NextProcessAddress.Valid {
		nextAddr = info.NextProcessAddress.String
	}
	if info.PreviousProcessAddress.Valid {
		prevAddr = info.PreviousProcessAddress.String
	}

	// Convert pgtype.Timestamp to time.Time
	var createTime, createdAt, updatedAt time.Time
	if info.CreateTime.Valid {
		createTime = info.CreateTime.Time
	}
	if info.CreatedAt.Valid {
		createdAt = info.CreatedAt.Time
	}
	if info.UpdatedAt.Valid {
		updatedAt = info.UpdatedAt.Time
	}

	return ProcessInfoResponse{
		ID:                     info.ID,
		UserID:                 info.UserID,
		ProcessID:              info.ProcessID,
		ParentProcessID:        info.ParentProcessID,
		ProcessName:            info.ProcessName,
		ThreadCount:            info.ThreadCount,
		HandleCount:            info.HandleCount,
		BasePriority:           info.BasePriority,
		CreateTime:             createTime,
		UserTime:               info.UserTime,
		KernelTime:             info.KernelTime,
		WorkingSetSize:         info.WorkingSetSize,
		PeakWorkingSetSize:     info.PeakWorkingSetSize,
		VirtualSize:            info.VirtualSize,
		PeakVirtualSize:        info.PeakVirtualSize,
		PagefileUsage:          info.PagefileUsage,
		PeakPagefileUsage:      info.PeakPagefileUsage,
		PageFaultCount:         info.PageFaultCount,
		ReadOperationCount:     info.ReadOperationCount,
		WriteOperationCount:    info.WriteOperationCount,
		OtherOperationCount:    info.OtherOperationCount,
		ReadTransferCount:      info.ReadTransferCount,
		WriteTransferCount:     info.WriteTransferCount,
		OtherTransferCount:     info.OtherTransferCount,
		CurrentProcessAddress:  currentAddr,
		NextProcessAddress:     nextAddr,
		PreviousProcessAddress: prevAddr,
		CreatedAt:              createdAt,
		UpdatedAt:              updatedAt,
	}
}
