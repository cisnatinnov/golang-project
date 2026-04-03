package handler

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"

	oapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/google/uuid"

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
