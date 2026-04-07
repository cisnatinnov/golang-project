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

func TestPostEstate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockRepositoryInterface(ctrl)
	e := echo.New()
	server := &Server{Repository: mockRepo}

	t.Run("Success", func(t *testing.T) {
		reqBody := `{"length": 10, "width": 5}`
		req := httptest.NewRequest(http.MethodPost, "/estate", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		estateId := uuid.New().String()
		mockRepo.EXPECT().
			CreateEstate(gomock.Any(), repository.CreateEstateInput{Length: 10, Width: 5}).
			Return(repository.CreateEstateOutput{Id: estateId}, nil)

		if assert.NoError(t, server.PostEstate(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("Invalid Dimensions", func(t *testing.T) {
		reqBody := `{"length": 0, "width": 5}`
		req := httptest.NewRequest(http.MethodPost, "/estate", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

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
	server := &Server{Repository: mockRepo}
	estateIdRaw := uuid.New()
	estateId := oapi_types.UUID(estateIdRaw)

	t.Run("Success", func(t *testing.T) {
		reqBody := `{"x": 2, "y": 3, "height": 10}`
		req := httptest.NewRequest(http.MethodPost, "/estate/"+estateId.String()+"/tree", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

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

	t.Run("Out of Bounds", func(t *testing.T) {
		reqBody := `{"x": 11, "y": 3, "height": 10}`
		req := httptest.NewRequest(http.MethodPost, "/estate/"+estateId.String()+"/tree", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

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
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

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

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/estate/"+estateId.String()+"/stats", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepo.EXPECT().GetEstateById(gomock.Any(), estateId.String()).Return(repository.Estate{}, nil)
		mockRepo.EXPECT().GetEstateStats(gomock.Any(), repository.GetEstateStatsInput{EstateId: estateId.String()}).
			Return(repository.GetEstateStatsOutput{Count: 10, Max: 20, Min: 5, Median: 12}, nil)

		if assert.NoError(t, server.GetEstateIdStats(c, estateId)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var resp generated.EstateStatsResponse
			json.Unmarshal(rec.Body.Bytes(), &resp)
			assert.Equal(t, 10, resp.Count)
			assert.Equal(t, 12, resp.Median)
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

	t.Run("Success Calculation", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/estate/"+estateId.String()+"/drone-plan", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

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
			json.Unmarshal(rec.Body.Bytes(), &resp)
			assert.Equal(t, 52, resp.Distance)
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
		mockRepo.EXPECT().
			GetUserByUsername(gomock.Any(), "testuser").
			Return(repository.User{Id: userId, Username: "testuser", Email: "test@example.com", PasswordHash: "hashed_password"}, nil)

		if assert.NoError(t, server.PostLogin(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var resp generated.LoginResponse
			json.Unmarshal(rec.Body.Bytes(), &resp)
			assert.Contains(t, resp.Token, "placeholder_token_")
		}
	})

	t.Run("Success with email", func(t *testing.T) {
		reqBody := `{"email": "test@example.com", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		userId := uuid.New().String()
		mockRepo.EXPECT().
			GetUserByEmail(gomock.Any(), "test@example.com").
			Return(repository.User{Id: userId, Username: "testuser", Email: "test@example.com", PasswordHash: "hashed_password"}, nil)

		if assert.NoError(t, server.PostLogin(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var resp generated.LoginResponse
			json.Unmarshal(rec.Body.Bytes(), &resp)
			assert.Contains(t, resp.Token, "placeholder_token_")
		}
	})

	t.Run("Invalid Credentials with username", func(t *testing.T) {
		reqBody := `{"username": "nonexistent", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepo.EXPECT().
			GetUserByUsername(gomock.Any(), "nonexistent").
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
			GetUserByEmail(gomock.Any(), "nonexistent@example.com").
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
		mockRepo.EXPECT().
			CreateUser(gomock.Any(), repository.CreateUserInput{
				Username:     "newuser",
				Email:        "newuser@example.com",
				PasswordHash: "password123",
			}).
			Return(repository.CreateUserOutput{Id: userId}, nil)

		if assert.NoError(t, server.PostUsers(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var resp generated.CreateUserResponse
			json.Unmarshal(rec.Body.Bytes(), &resp)
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
		req := httptest.NewRequest(http.MethodGet, "/users/"+userId.String(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepo.EXPECT().
			GetUserById(gomock.Any(), userId.String()).
			Return(repository.User{Id: userId.String(), Username: "testuser", Email: "test@example.com"}, nil)

		if assert.NoError(t, server.GetUsersId(c, userId)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var resp generated.UserResponse
			json.Unmarshal(rec.Body.Bytes(), &resp)
			assert.Equal(t, "testuser", resp.Username)
			assert.Equal(t, "test@example.com", string(resp.Email))
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/"+userId.String(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

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
		reqBody := `{"username": "updateduser", "email": "updated@example.com", "password": "newpassword"}`
		req := httptest.NewRequest(http.MethodPut, "/users/"+userId.String(), strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

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
			UpdateUser(gomock.Any(), repository.UpdateUserInput{
				Id:           userId.String(),
				Username:     "updateduser",
				Email:        "updated@example.com",
				PasswordHash: "newpassword",
			}).
			Return(nil)

		if assert.NoError(t, server.PutUsersId(c, userId)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		reqBody := `{"username": "updateduser", "email": "updated@example.com", "password": "newpassword"}`
		req := httptest.NewRequest(http.MethodPut, "/users/"+userId.String(), strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepo.EXPECT().
			GetUserById(gomock.Any(), userId.String()).
			Return(repository.User{}, sql.ErrNoRows)

		if assert.NoError(t, server.PutUsersId(c, userId)) {
			assert.Equal(t, http.StatusNotFound, rec.Code)
		}
	})

	t.Run("Username Already Exists", func(t *testing.T) {
		otherUserId := uuid.New().String()
		reqBody := `{"username": "takenuser", "email": "updated@example.com", "password": "newpassword"}`
		req := httptest.NewRequest(http.MethodPut, "/users/"+userId.String(), strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

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
		reqBody := `{"username": "updateduser", "email": "taken@example.com", "password": "newpassword"}`
		req := httptest.NewRequest(http.MethodPut, "/users/"+userId.String(), strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

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
		req := httptest.NewRequest(http.MethodDelete, "/users/"+userId.String(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

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

	t.Run("User Not Found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/users/"+userId.String(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepo.EXPECT().
			GetUserById(gomock.Any(), userId.String()).
			Return(repository.User{}, sql.ErrNoRows)

		if assert.NoError(t, server.DeleteUsersId(c, userId)) {
			assert.Equal(t, http.StatusNotFound, rec.Code)
		}
	})
}
