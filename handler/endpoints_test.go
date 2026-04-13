package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/SawitProRecruitment/UserService/repository"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	oapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
)

// Helper function to create JWT bearer token for testing
func createBearerToken(userID string) string {
	// Use a test secret for JWT generation
	token, err := GenerateToken(userID, "test-secret-key")
	if err != nil {
		panic("failed to generate test token: " + err.Error())
	}
	return token
}

func TestBearerTokenMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockRepositoryInterface(ctrl)
	server := &Server{Repository: mockRepo, JWTSecret: "test-secret-key"}
	e := echo.New()

	// Test handler that just returns OK
	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "ok"})
	}

	t.Run("Valid Bearer Token", func(t *testing.T) {
		userID := uuid.New().String()
		token, err := GenerateToken(userID, "test-secret-key")
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := server.BearerTokenMiddleware(testHandler)
		err = handler(c)

		if assert.NoError(t, err) {
			assert.Equal(t, userID, c.Get(ContextKeyUserID))
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("Missing Authorization Header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := server.BearerTokenMiddleware(testHandler)
		_ = handler(c)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		var resp generated.ErrorResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Contains(t, resp.Message, "Missing authorization header")
	})

	t.Run("Invalid Authorization Format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "InvalidToken")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := server.BearerTokenMiddleware(testHandler)
		_ = handler(c)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Empty Bearer Token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer ")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := server.BearerTokenMiddleware(testHandler)
		_ = handler(c)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Invalid Token Format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid_token_format")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := server.BearerTokenMiddleware(testHandler)
		_ = handler(c)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestPostEstate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockRepositoryInterface(ctrl)
	e := echo.New()
	server := &Server{Repository: mockRepo, JWTSecret: "test-secret-key"}
	userID := uuid.New().String()
	token, _ := GenerateToken(userID, "test-secret-key")

	t.Run("Success", func(t *testing.T) {
		reqBody := `{"length": 10, "width": 5}`
		req := httptest.NewRequest(http.MethodPost, "/estate", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(ContextKeyUserID, userID)

		estateId := uuid.New().String()
		mockRepo.EXPECT().
			CreateEstate(gomock.Any(), repository.CreateEstateInput{Length: 10, Width: 5}).
			Return(repository.CreateEstateOutput{Id: estateId}, nil)

		if assert.NoError(t, server.PostEstate(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("Unauthorized - No Token", func(t *testing.T) {
		reqBody := `{"length": 10, "width": 5}`
		req := httptest.NewRequest(http.MethodPost, "/estate", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, server.PostEstate(c)) {
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("Invalid Dimensions", func(t *testing.T) {
		reqBody := `{"length": 0, "width": 5}`
		req := httptest.NewRequest(http.MethodPost, "/estate", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(ContextKeyUserID, userID)

		if assert.NoError(t, server.PostEstate(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})
}

func TestPostEstateIdTree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockRepositoryInterface(ctrl)
	e := echo.New()
	server := &Server{Repository: mockRepo, JWTSecret: "test-secret-key"}
	estateIdRaw := uuid.New()
	estateId := oapi_types.UUID(estateIdRaw)
	userID := uuid.New().String()
	token, _ := GenerateToken(userID, "test-secret-key")

	t.Run("Success", func(t *testing.T) {
		reqBody := `{"x": 2, "y": 3, "height": 10}`
		req := httptest.NewRequest(http.MethodPost, "/estate/"+estateId.String()+"/tree", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(ContextKeyUserID, userID)

		mockRepo.EXPECT().GetEstateById(gomock.Any(), estateId.String()).
			Return(repository.Estate{Id: estateId.String(), Length: 10, Width: 10}, nil)

		treeId := uuid.New().String()
		mockRepo.EXPECT().CreateTree(gomock.Any(), repository.CreateTreeInput{
			EstateId: estateId.String(),
			X:        2,
			Y:        3,
			Height:   10,
		}).Return(repository.CreateTreeOutput{Id: treeId}, nil)

		if assert.NoError(t, server.PostEstateIdTree(c, estateId)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("Unauthorized - No Token", func(t *testing.T) {
		reqBody := `{"x": 2, "y": 3, "height": 10}`
		req := httptest.NewRequest(http.MethodPost, "/estate/"+estateId.String()+"/tree", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, server.PostEstateIdTree(c, estateId)) {
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("Out of Bounds", func(t *testing.T) {
		reqBody := `{"x": 11, "y": 3, "height": 10}`
		req := httptest.NewRequest(http.MethodPost, "/estate/"+estateId.String()+"/tree", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(ContextKeyUserID, userID)

		mockRepo.EXPECT().GetEstateById(gomock.Any(), estateId.String()).
			Return(repository.Estate{Id: estateId.String(), Length: 10, Width: 10}, nil)

		if assert.NoError(t, server.PostEstateIdTree(c, estateId)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("Estate Not Found", func(t *testing.T) {
		reqBody := `{"x": 2, "y": 3, "height": 10}`
		req := httptest.NewRequest(http.MethodPost, "/estate/"+estateId.String()+"/tree", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(ContextKeyUserID, userID)

		mockRepo.EXPECT().GetEstateById(gomock.Any(), estateId.String()).Return(repository.Estate{}, sql.ErrNoRows)

		if assert.NoError(t, server.PostEstateIdTree(c, estateId)) {
			assert.Equal(t, http.StatusNotFound, rec.Code)
		}
	})
}

func TestGetEstateIdStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockRepositoryInterface(ctrl)
	e := echo.New()
	server := &Server{Repository: mockRepo}
	estateIdRaw := uuid.New()
	estateId := oapi_types.UUID(estateIdRaw)
	userID := uuid.New().String()
	token := createBearerToken(userID)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/estate/"+estateId.String()+"/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(ContextKeyUserID, userID)

		mockRepo.EXPECT().GetEstateById(gomock.Any(), estateId.String()).Return(repository.Estate{}, nil)
		mockRepo.EXPECT().GetEstateStats(gomock.Any(), repository.GetEstateStatsInput{EstateId: estateId.String()}).
			Return(repository.GetEstateStatsOutput{Count: 10, Max: 20, Min: 5, Median: 12}, nil)

		if assert.NoError(t, server.GetEstateIdStats(c, estateId)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var resp generated.EstateStatsResponse
			assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, 10, resp.Count)
			assert.Equal(t, 12, resp.Median)
		}
	})

	t.Run("Unauthorized - No Token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/estate/"+estateId.String()+"/stats", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, server.GetEstateIdStats(c, estateId)) {
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		}
	})
}

func TestGetEstateIdDronePlan(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockRepositoryInterface(ctrl)
	e := echo.New()
	server := &Server{Repository: mockRepo}
	estateIdRaw := uuid.New()
	estateId := oapi_types.UUID(estateIdRaw)
	userID := uuid.New().String()
	token := createBearerToken(userID)

	t.Run("Success Calculation", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/estate/"+estateId.String()+"/drone-plan", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(ContextKeyUserID, userID)

		// 5x1 grid, one tree at (2,1) height 5
		mockRepo.EXPECT().GetEstateById(gomock.Any(), estateId.String()).
			Return(repository.Estate{Length: 5, Width: 1}, nil)

		trees := []repository.Tree{
			{X: 2, Y: 1, Height: 5},
		}
		mockRepo.EXPECT().GetTreesByEstateId(gomock.Any(), repository.GetTreesByEstateIdInput{EstateId: estateId.String()}).
			Return(repository.GetTreesByEstateIdOutput{Trees: trees}, nil)

		// Calculation:
		// Horizontal: (5*1 - 1) * 10 = 40
		// Vertical path:
		// (1,1): target 0, current 0. delta 0.
		// (2,1): target 6, current 0. delta 6.
		// (3,1): target 0, current 6. delta 6.
		// (4,1): target 0, current 0. delta 0.
		// (5,1): target 0, current 0. delta 0.
		// total vertical = 12
		// total distance = 40 + 12 = 52

		if assert.NoError(t, server.GetEstateIdDronePlan(c, estateId, generated.GetEstateIdDronePlanParams{})) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var resp generated.DronePlanResponse
			assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, 52, resp.Distance)
		}
	})

	t.Run("Unauthorized - No Token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/estate/"+estateId.String()+"/drone-plan", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, server.GetEstateIdDronePlan(c, estateId, generated.GetEstateIdDronePlanParams{})) {
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		}
	})
}

func TestPostLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockRepositoryInterface(ctrl)
	e := echo.New()
	server := &Server{Repository: mockRepo}

	t.Run("Success with username", func(t *testing.T) {
		reqBody := `{"username": "testuser", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		userId := uuid.New().String()
		// Use valid bcrypt hash for "password123"
		mockRepo.EXPECT().
			GetUserByUsernameOrEmail(gomock.Any(), repository.GetUserByUsernameOrEmailInput{Username: "testuser", Email: ""}).
			Return(repository.User{Id: userId, Username: "testuser", Email: "test@example.com", PasswordHash: "$2a$10$7cu3I0HGsd2ECtQ2ITgeb.AZKbsdDQT0JVAHyJJ6NU/IBAVKoh8EG"}, nil)

		if assert.NoError(t, server.PostLogin(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var resp generated.LoginResponse
			assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			// Validate JWT token can be parsed and contains correct user ID
			extractedUserID, err := ValidateToken(resp.Token, server.JWTSecret)
			assert.NoError(t, err)
			assert.Equal(t, userId, extractedUserID)
		}
	})

	t.Run("Success with email", func(t *testing.T) {
		reqBody := `{"email": "test@example.com", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		userId := uuid.New().String()
		// Use valid bcrypt hash for "password123"
		mockRepo.EXPECT().
			GetUserByUsernameOrEmail(gomock.Any(), repository.GetUserByUsernameOrEmailInput{Username: "", Email: "test@example.com"}).
			Return(repository.User{Id: userId, Username: "testuser", Email: "test@example.com", PasswordHash: "$2a$10$7cu3I0HGsd2ECtQ2ITgeb.AZKbsdDQT0JVAHyJJ6NU/IBAVKoh8EG"}, nil)

		if assert.NoError(t, server.PostLogin(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var resp generated.LoginResponse
			assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			// Validate JWT token can be parsed and contains correct user ID
			extractedUserID, err := ValidateToken(resp.Token, server.JWTSecret)
			assert.NoError(t, err)
			assert.Equal(t, userId, extractedUserID)
		}
	})

	t.Run("Invalid Credentials with username", func(t *testing.T) {
		reqBody := `{"username": "nonexistent", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepo.EXPECT().
			GetUserByUsernameOrEmail(gomock.Any(), repository.GetUserByUsernameOrEmailInput{Username: "nonexistent", Email: ""}).
			Return(repository.User{}, sql.ErrNoRows)

		if assert.NoError(t, server.PostLogin(c)) {
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("Invalid Credentials with email", func(t *testing.T) {
		reqBody := `{"email": "nonexistent@example.com", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepo.EXPECT().
			GetUserByUsernameOrEmail(gomock.Any(), repository.GetUserByUsernameOrEmailInput{Username: "", Email: "nonexistent@example.com"}).
			Return(repository.User{}, sql.ErrNoRows)

		if assert.NoError(t, server.PostLogin(c)) {
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("Missing Credentials", func(t *testing.T) {
		reqBody := `{"password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, server.PostLogin(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("Missing Password", func(t *testing.T) {
		reqBody := `{"username": "testuser"}`
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, server.PostLogin(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})
}

func TestPostUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockRepositoryInterface(ctrl)
	e := echo.New()
	server := &Server{Repository: mockRepo}

	t.Run("Success", func(t *testing.T) {
		reqBody := `{"username": "newuser", "email": "newuser@example.com", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepo.EXPECT().
			GetUserByUsername(gomock.Any(), "newuser").
			Return(repository.User{}, sql.ErrNoRows)

		mockRepo.EXPECT().
			GetUserByEmail(gomock.Any(), "newuser@example.com").
			Return(repository.User{}, sql.ErrNoRows)

		userId := uuid.New().String()
		// Use gomock.Any() for PasswordHash since bcrypt produces non-deterministic output
		mockRepo.EXPECT().
			CreateUser(gomock.Any(), gomock.All(
			// Verify other fields
			)).
			Do(func(ctx interface{}, input repository.CreateUserInput) {
				// Verify the input has expected values (except PasswordHash which is hashed)
				assert.Equal(t, "newuser", input.Username)
				assert.Equal(t, "newuser@example.com", input.Email)
				// PasswordHash should be a non-empty bcrypt hash
				assert.NotEmpty(t, input.PasswordHash)
				assert.NotEqual(t, "password123", input.PasswordHash)
			}).
			Return(repository.CreateUserOutput{Id: userId}, nil)

		if assert.NoError(t, server.PostUsers(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var resp generated.CreateUserResponse
			assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, userId, resp.Id.String())
		}
	})

	t.Run("Username Already Exists", func(t *testing.T) {
		reqBody := `{"username": "existinguser", "email": "new@example.com", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepo.EXPECT().
			GetUserByUsername(gomock.Any(), "existinguser").
			Return(repository.User{Id: uuid.New().String(), Username: "existinguser", Email: "existing@example.com"}, nil)

		if assert.NoError(t, server.PostUsers(c)) {
			assert.Equal(t, http.StatusConflict, rec.Code)
		}
	})

	t.Run("Email Already Exists", func(t *testing.T) {
		reqBody := `{"username": "newuser", "email": "existing@example.com", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepo.EXPECT().
			GetUserByUsername(gomock.Any(), "newuser").
			Return(repository.User{}, sql.ErrNoRows)

		mockRepo.EXPECT().
			GetUserByEmail(gomock.Any(), "existing@example.com").
			Return(repository.User{Id: uuid.New().String(), Username: "existinguser", Email: "existing@example.com"}, nil)

		if assert.NoError(t, server.PostUsers(c)) {
			assert.Equal(t, http.StatusConflict, rec.Code)
		}
	})

	t.Run("Missing Credentials", func(t *testing.T) {
		reqBody := `{"username": "", "email": "", "password": ""}`
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, server.PostUsers(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})
}

func TestGetUsersId(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockRepositoryInterface(ctrl)
	e := echo.New()
	server := &Server{Repository: mockRepo}
	userIdRaw := uuid.New()
	userId := oapi_types.UUID(userIdRaw)

	t.Run("Success", func(t *testing.T) {
		token := createBearerToken(userId.String())
		req := httptest.NewRequest(http.MethodGet, "/users/"+userId.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(ContextKeyUserID, userId.String())

		mockRepo.EXPECT().
			GetUserById(gomock.Any(), userId.String()).
			Return(repository.User{Id: userId.String(), Username: "testuser", Email: "test@example.com"}, nil)

		if assert.NoError(t, server.GetUsersId(c, userId)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var resp generated.UserResponse
			assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, "testuser", resp.Username)
			assert.Equal(t, "test@example.com", string(resp.Email))
		}
	})

	t.Run("Unauthorized - No Token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/"+userId.String(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, server.GetUsersId(c, userId)) {
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("Unauthorized - Different User", func(t *testing.T) {
		otherUserId := uuid.New().String()
		token := createBearerToken(otherUserId)
		req := httptest.NewRequest(http.MethodGet, "/users/"+userId.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(ContextKeyUserID, otherUserId)

		if assert.NoError(t, server.GetUsersId(c, userId)) {
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		token := createBearerToken(userId.String())
		req := httptest.NewRequest(http.MethodGet, "/users/"+userId.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(ContextKeyUserID, userId.String())

		mockRepo.EXPECT().
			GetUserById(gomock.Any(), userId.String()).
			Return(repository.User{}, sql.ErrNoRows)

		if assert.NoError(t, server.GetUsersId(c, userId)) {
			assert.Equal(t, http.StatusNotFound, rec.Code)
		}
	})
}

func TestPutUsersId(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockRepositoryInterface(ctrl)
	e := echo.New()
	server := &Server{Repository: mockRepo}
	userIdRaw := uuid.New()
	userId := oapi_types.UUID(userIdRaw)

	t.Run("Success", func(t *testing.T) {
		token := createBearerToken(userId.String())
		reqBody := `{"username": "updateduser", "email": "updated@example.com", "password": "newpassword"}`
		req := httptest.NewRequest(http.MethodPut, "/users/"+userId.String(), strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(ContextKeyUserID, userId.String())

		mockRepo.EXPECT().
			GetUserById(gomock.Any(), userId.String()).
			Return(repository.User{Id: userId.String(), Username: "olduser", Email: "old@example.com"}, nil)

		mockRepo.EXPECT().
			GetUserByUsername(gomock.Any(), "updateduser").
			Return(repository.User{}, sql.ErrNoRows)

		mockRepo.EXPECT().
			GetUserByEmail(gomock.Any(), "updated@example.com").
			Return(repository.User{}, sql.ErrNoRows)

		mockRepo.EXPECT().
			UpdateUser(gomock.Any(), gomock.Any()).
			Return(nil)

		if assert.NoError(t, server.PutUsersId(c, userId)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("Unauthorized - No Token", func(t *testing.T) {
		reqBody := `{"username": "updateduser", "email": "updated@example.com", "password": "newpassword"}`
		req := httptest.NewRequest(http.MethodPut, "/users/"+userId.String(), strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, server.PutUsersId(c, userId)) {
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("Unauthorized - Different User", func(t *testing.T) {
		otherUserId := uuid.New().String()
		token := createBearerToken(otherUserId)
		reqBody := `{"username": "updateduser", "email": "updated@example.com", "password": "newpassword"}`
		req := httptest.NewRequest(http.MethodPut, "/users/"+userId.String(), strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(ContextKeyUserID, otherUserId)

		if assert.NoError(t, server.PutUsersId(c, userId)) {
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		token := createBearerToken(userId.String())
		reqBody := `{"username": "updateduser", "email": "updated@example.com", "password": "newpassword"}`
		req := httptest.NewRequest(http.MethodPut, "/users/"+userId.String(), strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(ContextKeyUserID, userId.String())

		mockRepo.EXPECT().
			GetUserById(gomock.Any(), userId.String()).
			Return(repository.User{}, sql.ErrNoRows)

		if assert.NoError(t, server.PutUsersId(c, userId)) {
			assert.Equal(t, http.StatusNotFound, rec.Code)
		}
	})

	t.Run("Username Already Exists", func(t *testing.T) {
		otherUserId := uuid.New().String()
		token := createBearerToken(userId.String())
		reqBody := `{"username": "takenuser", "email": "updated@example.com", "password": "newpassword"}`
		req := httptest.NewRequest(http.MethodPut, "/users/"+userId.String(), strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(ContextKeyUserID, userId.String())

		mockRepo.EXPECT().
			GetUserById(gomock.Any(), userId.String()).
			Return(repository.User{Id: userId.String(), Username: "olduser", Email: "old@example.com"}, nil)

		mockRepo.EXPECT().
			GetUserByUsername(gomock.Any(), "takenuser").
			Return(repository.User{Id: otherUserId, Username: "takenuser", Email: "taken@example.com"}, nil)

		if assert.NoError(t, server.PutUsersId(c, userId)) {
			assert.Equal(t, http.StatusConflict, rec.Code)
		}
	})

	t.Run("Email Already Exists", func(t *testing.T) {
		otherUserId := uuid.New().String()
		token := createBearerToken(userId.String())
		reqBody := `{"username": "updateduser", "email": "taken@example.com", "password": "newpassword"}`
		req := httptest.NewRequest(http.MethodPut, "/users/"+userId.String(), strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(ContextKeyUserID, userId.String())

		mockRepo.EXPECT().
			GetUserById(gomock.Any(), userId.String()).
			Return(repository.User{Id: userId.String(), Username: "olduser", Email: "old@example.com"}, nil)

		mockRepo.EXPECT().
			GetUserByUsername(gomock.Any(), "updateduser").
			Return(repository.User{}, sql.ErrNoRows)

		mockRepo.EXPECT().
			GetUserByEmail(gomock.Any(), "taken@example.com").
			Return(repository.User{Id: otherUserId, Username: "otheruser", Email: "taken@example.com"}, nil)

		if assert.NoError(t, server.PutUsersId(c, userId)) {
			assert.Equal(t, http.StatusConflict, rec.Code)
		}
	})
}

func TestDeleteUsersId(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockRepositoryInterface(ctrl)
	e := echo.New()
	server := &Server{Repository: mockRepo}
	userIdRaw := uuid.New()
	userId := oapi_types.UUID(userIdRaw)

	t.Run("Success", func(t *testing.T) {
		token := createBearerToken(userId.String())
		req := httptest.NewRequest(http.MethodDelete, "/users/"+userId.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(ContextKeyUserID, userId.String())

		mockRepo.EXPECT().
			GetUserById(gomock.Any(), userId.String()).
			Return(repository.User{Id: userId.String(), Username: "testuser"}, nil)

		mockRepo.EXPECT().
			DeleteUser(gomock.Any(), userId.String()).
			Return(nil)

		if assert.NoError(t, server.DeleteUsersId(c, userId)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("Unauthorized - No Token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/users/"+userId.String(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, server.DeleteUsersId(c, userId)) {
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("Unauthorized - Different User", func(t *testing.T) {
		otherUserId := uuid.New().String()
		token := createBearerToken(otherUserId)
		req := httptest.NewRequest(http.MethodDelete, "/users/"+userId.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(ContextKeyUserID, otherUserId)

		if assert.NoError(t, server.DeleteUsersId(c, userId)) {
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		token := createBearerToken(userId.String())
		req := httptest.NewRequest(http.MethodDelete, "/users/"+userId.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(ContextKeyUserID, userId.String())

		mockRepo.EXPECT().
			GetUserById(gomock.Any(), userId.String()).
			Return(repository.User{}, sql.ErrNoRows)

		if assert.NoError(t, server.DeleteUsersId(c, userId)) {
			assert.Equal(t, http.StatusNotFound, rec.Code)
		}
	})
}
