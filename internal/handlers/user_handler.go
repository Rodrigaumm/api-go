package handlers

import (
	"crypto/sha512"
	"encoding/hex"
	"strconv"

	"go-api/internal/db"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserHandler struct {
	queries *db.Queries
}

func NewUserHandler(dbpool *pgxpool.Pool) *UserHandler {
	return &UserHandler{
		queries: db.New(dbpool),
	}
}

type CreateUserRequest struct {
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

type UpdateUserRequest struct {
	Name     *string `json:"name,omitempty"`
	Password *string `json:"password,omitempty"`
}

type UserResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// hashPassword cria a hash SHA-512 da senha
func hashPassword(password string) string {
	hash := sha512.Sum512([]byte(password))
	return hex.EncodeToString(hash[:])
}

func toUserResponse(user interface{}) UserResponse {
	switch u := user.(type) {
	case db.User:
		return UserResponse{
			ID:   u.ID,
			Name: u.Name,
		}
	default:
		return UserResponse{}
	}
}

// GetUsers - Listar todos os usuários
func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	users, err := h.queries.GetUsers(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Erro ao buscar usuários",
		})
	}

	response := make([]UserResponse, len(users))
	for i, user := range users {
		response[i] = toUserResponse(user)
	}

	return c.JSON(fiber.Map{
		"data":  response,
		"count": len(response),
	})
}

// GetUser - Buscar usuário por ID
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID inválido",
		})
	}

	user, err := h.queries.GetUser(c.Context(), id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{
				"error": "Usuário não encontrado",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Erro ao buscar usuário",
		})
	}

	return c.JSON(fiber.Map{
		"data": toUserResponse(user),
	})
}

// CreateUser - Criar novo usuário
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Dados inválidos",
		})
	}

	hashedPassword := hashPassword(req.Password)

	params := db.CreateUserParams{
		Name:     req.Name,
		Password: hashedPassword,
	}

	user, err := h.queries.CreateUser(c.Context(), params)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Erro ao criar usuário",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"data":    toUserResponse(user),
		"message": "Usuário criado com sucesso",
	})
}

// UpdateUser - Atualizar usuário
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID inválido",
		})
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Dados inválidos",
		})
	}

	currentUser, err := h.queries.GetUser(c.Context(), id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{
				"error": "Usuário não encontrado",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Erro ao buscar usuário",
		})
	}

	params := db.UpdateUserParams{
		ID: id,
	}

	if req.Name != nil {
		params.Name = *req.Name
	} else {
		params.Name = currentUser.Name
	}

	if req.Password != nil {
		params.Password = hashPassword(*req.Password)
	} else {
		params.Password = currentUser.Password
	}

	user, err := h.queries.UpdateUser(c.Context(), params)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Erro ao atualizar usuário",
		})
	}

	return c.JSON(fiber.Map{
		"data":    toUserResponse(user),
		"message": "Usuário atualizado com sucesso",
	})
}

// DeleteUser - Deletar usuário
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID inválido",
		})
	}

	err = h.queries.DeleteUser(c.Context(), id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{
				"error": "Usuário não encontrado",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Erro ao deletar usuário",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Usuário deletado com sucesso",
	})
}
