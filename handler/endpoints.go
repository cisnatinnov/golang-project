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
	// Get authenticated user ID from context
	authUserID := GetUserIDFromContext(ctx)
	if authUserID == "" {
		return ctx.JSON(http.StatusUnauthorized, generated.ErrorResponse{Message: "Unauthorized"})
	}

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
	// Get authenticated user ID from context
	authUserID := GetUserIDFromContext(ctx)
	if authUserID == "" {
		return ctx.JSON(http.StatusUnauthorized, generated.ErrorResponse{Message: "Unauthorized"})
	}

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
	// Get authenticated user ID from context
	authUserID := GetUserIDFromContext(ctx)
	if authUserID == "" {
		return ctx.JSON(http.StatusUnauthorized, generated.ErrorResponse{Message: "Unauthorized"})
	}

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
	// Get authenticated user ID from context
	authUserID := GetUserIDFromContext(ctx)
	if authUserID == "" {
		return ctx.JSON(http.StatusUnauthorized, generated.ErrorResponse{Message: "Unauthorized"})
	}

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

	username := ""
	if req.Username != nil {
		username = *req.Username
	}

	email := ""
	if req.Email != nil {
		email = string(*req.Email)
	}

	user, err := s.Repository.GetUserByUsernameOrEmail(ctx.Request().Context(), repository.GetUserByUsernameOrEmailInput{
		Username: username,
		Email:    email,
	})

	if err == sql.ErrNoRows {
		return ctx.JSON(http.StatusUnauthorized, generated.ErrorResponse{Message: "Invalid credentials"})
	} else if err != nil {
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: err.Error()})
	}

	// Verify password
	err = repository.VerifyPassword(user.PasswordHash, req.Password)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, generated.ErrorResponse{Message: "Invalid credentials"})
	}

	// Generate JWT token
	token, err := GenerateToken(user.Id, s.JWTSecret)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{
			Message: "Failed to generate token: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, generated.LoginResponse{
		Token: token,
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

	// Hash password
	hashedPassword, err := repository.HashPassword(req.Password)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: "Failed to hash password"})
	}

	// Create user
	outUser, err := s.Repository.CreateUser(ctx.Request().Context(), repository.CreateUserInput{
		Username:     req.Username,
		PasswordHash: hashedPassword,
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: err.Error()})
	}

	// Create person profile
	personOut, err := s.Repository.CreatePerson(ctx.Request().Context(), repository.CreatePersonInput{
		UserId: outUser.Id,
	})
	if err != nil {
		// Rollback user creation if person creation fails
		_ = s.Repository.DeleteUser(ctx.Request().Context(), outUser.Id)
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: err.Error()})
	}

	// Add email to person profile
	_, err = s.Repository.CreatePersonEmail(ctx.Request().Context(), repository.CreatePersonEmailInput{
		UserId:    outUser.Id,
		Email:     string(req.Email),
		IsPrimary: true,
	})
	if err != nil {
		// Rollback if email creation fails
		_ = s.Repository.DeletePerson(ctx.Request().Context(), personOut.Id)
		_ = s.Repository.DeleteUser(ctx.Request().Context(), outUser.Id)
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: err.Error()})
	}

	parsedId, _ := uuid.Parse(outUser.Id)
	return ctx.JSON(http.StatusOK, generated.CreateUserResponse{
		Id: parsedId,
	})
}

func (s *Server) GetUsersId(ctx echo.Context, id oapi_types.UUID) error {
	// Get authenticated user ID from context
	authUserID := GetUserIDFromContext(ctx)
	if authUserID == "" {
		return ctx.JSON(http.StatusUnauthorized, generated.ErrorResponse{Message: "Unauthorized"})
	}

	// Verify user is accessing their own profile
	if authUserID != id.String() {
		return ctx.JSON(http.StatusUnauthorized, generated.ErrorResponse{Message: "Unauthorized to access this user"})
	}

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
	})
}

func (s *Server) PutUsersId(ctx echo.Context, id oapi_types.UUID) error {
	// Get authenticated user ID from context
	authUserID := GetUserIDFromContext(ctx)
	if authUserID == "" {
		return ctx.JSON(http.StatusUnauthorized, generated.ErrorResponse{Message: "Unauthorized"})
	}

	// Verify user is accessing their own profile
	if authUserID != id.String() {
		return ctx.JSON(http.StatusUnauthorized, generated.ErrorResponse{Message: "Unauthorized to access this user"})
	}

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

	// Hash password
	hashedPassword, err := repository.HashPassword(req.Password)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: "Failed to hash password"})
	}

	err = s.Repository.UpdateUser(ctx.Request().Context(), repository.UpdateUserInput{
		Id:           id.String(),
		Username:     req.Username,
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
	// Get authenticated user ID from context
	authUserID := GetUserIDFromContext(ctx)
	if authUserID == "" {
		return ctx.JSON(http.StatusUnauthorized, generated.ErrorResponse{Message: "Unauthorized"})
	}

	// Verify user is accessing their own profile
	if authUserID != id.String() {
		return ctx.JSON(http.StatusUnauthorized, generated.ErrorResponse{Message: "Unauthorized to access this user"})
	}

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
