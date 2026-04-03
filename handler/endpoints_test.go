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
