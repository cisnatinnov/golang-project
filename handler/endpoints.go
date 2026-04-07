package handler

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"

	"github.com/google/uuid"
	oapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/SawitProRecruitment/UserService/repository"
	"github.com/labstack/echo/v4"
)

func (s *Server) PostEstate(ctx echo.Context) error {
	var req generated.CreateEstateRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: "Invalid request body"})
	}

	if req.Length < 1 || req.Width < 1 {
		return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: "Dimensions must be strictly positive"})
	}

	out, err := s.Repository.CreateEstate(ctx.Request().Context(), repository.CreateEstateInput{
		Length: req.Length,
		Width:  req.Width,
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: err.Error()})
	}

	// UUID conversion
	parsedId, _ := uuid.Parse(out.Id)
	return ctx.JSON(http.StatusOK, generated.CreateEstateResponse{
		Id: parsedId,
	})
}

func (s *Server) PostEstateIdTree(ctx echo.Context, id oapi_types.UUID) error {
	var req generated.AddTreeRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: "Invalid request body"})
	}

	estate, err := s.Repository.GetEstateById(ctx.Request().Context(), id.String())
	if err == sql.ErrNoRows {
		return ctx.JSON(http.StatusNotFound, generated.ErrorResponse{Message: "Estate not found"})
	} else if err != nil {
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: err.Error()})
	}

	// Based on test semantics: X goes up to Length, Y goes up to Width
	if req.X < 1 || req.X > estate.Length || req.Y < 1 || req.Y > estate.Width {
		return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: "Tree coordinate out of bounds"})
	}

	if req.Height < 1 {
		return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: "Tree height must be positive"})
	}

	out, err := s.Repository.CreateTree(ctx.Request().Context(), repository.CreateTreeInput{
		EstateId: id.String(),
		X:        req.X,
		Y:        req.Y,
		Height:   req.Height,
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: err.Error()})
	}

	parsedId, _ := uuid.Parse(out.Id)
	return ctx.JSON(http.StatusOK, generated.AddTreeResponse{
		Id: parsedId,
	})
}

func (s *Server) GetEstateIdStats(ctx echo.Context, id oapi_types.UUID) error {
	// check if estate exists
	_, err := s.Repository.GetEstateById(ctx.Request().Context(), id.String())
	if err == sql.ErrNoRows {
		return ctx.JSON(http.StatusNotFound, generated.ErrorResponse{Message: "Estate not found"})
	} else if err != nil {
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: err.Error()})
	}

	stats, err := s.Repository.GetEstateStats(ctx.Request().Context(), repository.GetEstateStatsInput{
		EstateId: id.String(),
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: err.Error()})
	}

	return ctx.JSON(http.StatusOK, generated.EstateStatsResponse{
		Count:  stats.Count,
		Max:    stats.Max,
		Min:    stats.Min,
		Median: stats.Median,
	})
}

func (s *Server) GetEstateIdDronePlan(ctx echo.Context, id oapi_types.UUID, params generated.GetEstateIdDronePlanParams) error {
	estate, err := s.Repository.GetEstateById(ctx.Request().Context(), id.String())
	if err == sql.ErrNoRows {
		return ctx.JSON(http.StatusNotFound, generated.ErrorResponse{Message: "Estate not found"})
	} else if err != nil {
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: err.Error()})
	}

	treesOut, err := s.Repository.GetTreesByEstateId(ctx.Request().Context(), repository.GetTreesByEstateIdInput{
		EstateId: id.String(),
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: err.Error()})
	}

	// Map tree coordinates to heights
	treeMap := make(map[string]int)
	for _, t := range treesOut.Trees {
		key := fmt.Sprintf("%d,%d", t.X, t.Y)
		treeMap[key] = t.Height
	}

	// Calculate Drone Distance
	totalCells := estate.Length * estate.Width
	horizontalDistance := (totalCells - 1) * 10

	verticalDistance := 0
	currentHeight := 0 // Altitude starts at 0

	for y := 1; y <= estate.Width; y++ {
		if y%2 == 1 {
			for x := 1; x <= estate.Length; x++ {
				key := fmt.Sprintf("%d,%d", x, y)
				targetAltitude := 0
				if h, exists := treeMap[key]; exists {
					targetAltitude = h + 1
				}
				verticalDistance += int(math.Abs(float64(targetAltitude - currentHeight)))
				currentHeight = targetAltitude
			}
		} else {
			for x := estate.Length; x >= 1; x-- {
				key := fmt.Sprintf("%d,%d", x, y)
				targetAltitude := 0
				if h, exists := treeMap[key]; exists {
					targetAltitude = h + 1
				}
				verticalDistance += int(math.Abs(float64(targetAltitude - currentHeight)))
				currentHeight = targetAltitude
			}
		}
	}

	totalDist := horizontalDistance + verticalDistance

	return ctx.JSON(http.StatusOK, generated.DronePlanResponse{
		Distance: totalDist,
	})
}

// GetHello implements the test endpoint
func (s *Server) GetHello(ctx echo.Context, params generated.GetHelloParams) error {
	var resp generated.HelloResponse
	resp.Message = fmt.Sprintf("Hello User %d", params.Id)
	return ctx.JSON(http.StatusOK, resp)
}

func (s *Server) PostLogin(ctx echo.Context) error {
	var req generated.LoginRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: "Invalid request body"})
	}

	if req.Password == "" {
		return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: "Password is required"})
	}

	// Require either username or email
	if (req.Username == nil || *req.Username == "") && (req.Email == nil || string(*req.Email) == "") {
		return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: "Username or email is required"})
	}

	var user repository.User
	var err error

	// Try to login with username if provided
	if req.Username != nil && *req.Username != "" {
		user, err = s.Repository.GetUserByUsername(ctx.Request().Context(), *req.Username)
	} else if req.Email != nil && string(*req.Email) != "" {
		// Otherwise try to login with email
		user, err = s.Repository.GetUserByEmail(ctx.Request().Context(), string(*req.Email))
	}

	if err == sql.ErrNoRows {
		return ctx.JSON(http.StatusUnauthorized, generated.ErrorResponse{Message: "Invalid credentials"})
	} else if err != nil {
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: err.Error()})
	}

	// Placeholder: In real implementation, verify password hash
	// For now, just return a token
	return ctx.JSON(http.StatusOK, generated.LoginResponse{
		Token: "placeholder_token_" + user.Id,
	})
}

func (s *Server) PostUsers(ctx echo.Context) error {
	var req generated.CreateUserRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: "Invalid request body"})
	}

	if req.Username == "" || string(req.Email) == "" || req.Password == "" {
		return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: "Username, email, and password are required"})
	}

	// Check if username already exists
	_, err := s.Repository.GetUserByUsername(ctx.Request().Context(), req.Username)
	if err == nil {
		return ctx.JSON(http.StatusConflict, generated.ErrorResponse{Message: "Username already exists"})
	} else if err != sql.ErrNoRows {
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: err.Error()})
	}

	// Check if email already exists
	_, err = s.Repository.GetUserByEmail(ctx.Request().Context(), string(req.Email))
	if err == nil {
		return ctx.JSON(http.StatusConflict, generated.ErrorResponse{Message: "Email already exists"})
	} else if err != sql.ErrNoRows {
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: err.Error()})
	}

	// Hash password (placeholder)
	hashedPassword := req.Password // In real implementation, hash this

	out, err := s.Repository.CreateUser(ctx.Request().Context(), repository.CreateUserInput{
		Username:     req.Username,
		Email:        string(req.Email),
		PasswordHash: hashedPassword,
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: err.Error()})
	}

	parsedId, _ := uuid.Parse(out.Id)
	return ctx.JSON(http.StatusOK, generated.CreateUserResponse{
		Id: parsedId,
	})
}

func (s *Server) GetUsersId(ctx echo.Context, id oapi_types.UUID) error {
	user, err := s.Repository.GetUserById(ctx.Request().Context(), id.String())
	if err == sql.ErrNoRows {
		return ctx.JSON(http.StatusNotFound, generated.ErrorResponse{Message: "User not found"})
	} else if err != nil {
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: err.Error()})
	}

	parsedId, _ := uuid.Parse(user.Id)
	return ctx.JSON(http.StatusOK, generated.UserResponse{
		Id:       parsedId,
		Username: user.Username,
		Email:    oapi_types.Email(user.Email),
	})
}

func (s *Server) PutUsersId(ctx echo.Context, id oapi_types.UUID) error {
	var req generated.UpdateUserRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: "Invalid request body"})
	}

	if req.Username == "" || string(req.Email) == "" || req.Password == "" {
		return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: "Username, email, and password are required"})
	}

	// Check if user exists
	currentUser, err := s.Repository.GetUserById(ctx.Request().Context(), id.String())
	if err == sql.ErrNoRows {
		return ctx.JSON(http.StatusNotFound, generated.ErrorResponse{Message: "User not found"})
	} else if err != nil {
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: err.Error()})
	}

	// Check if new username already exists (if changed)
	if req.Username != currentUser.Username {
		_, err := s.Repository.GetUserByUsername(ctx.Request().Context(), req.Username)
		if err == nil {
			return ctx.JSON(http.StatusConflict, generated.ErrorResponse{Message: "Username already exists"})
		} else if err != sql.ErrNoRows {
			return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: err.Error()})
		}
	}

	// Check if new email already exists (if changed)
	if string(req.Email) != currentUser.Email {
		_, err := s.Repository.GetUserByEmail(ctx.Request().Context(), string(req.Email))
		if err == nil {
			return ctx.JSON(http.StatusConflict, generated.ErrorResponse{Message: "Email already exists"})
		} else if err != sql.ErrNoRows {
			return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: err.Error()})
		}
	}

	// Hash password (placeholder)
	hashedPassword := req.Password // In real implementation, hash this

	err = s.Repository.UpdateUser(ctx.Request().Context(), repository.UpdateUserInput{
		Id:           id.String(),
		Username:     req.Username,
		Email:        string(req.Email),
		PasswordHash: hashedPassword,
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: err.Error()})
	}

	return ctx.JSON(http.StatusOK, generated.UpdateUserResponse{
		Message: "User updated successfully",
	})
}

func (s *Server) DeleteUsersId(ctx echo.Context, id oapi_types.UUID) error {
	// Check if user exists
	_, err := s.Repository.GetUserById(ctx.Request().Context(), id.String())
	if err == sql.ErrNoRows {
		return ctx.JSON(http.StatusNotFound, generated.ErrorResponse{Message: "User not found"})
	} else if err != nil {
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: err.Error()})
	}

	err = s.Repository.DeleteUser(ctx.Request().Context(), id.String())
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: err.Error()})
	}

	return ctx.JSON(http.StatusOK, generated.DeleteUserResponse{
		Message: "User deleted successfully",
	})
}
