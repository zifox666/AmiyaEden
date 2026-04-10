package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"errors"
	"math"
)

// HallOfFameService 名人堂业务逻辑层
type HallOfFameService struct {
	repo *repository.HallOfFameRepository
}

func NewHallOfFameService() *HallOfFameService {
	return &HallOfFameService{
		repo: repository.NewHallOfFameRepository(),
	}
}

// ─── Config ───

// GetConfig returns the singleton config, creating default if not exists.
func (s *HallOfFameService) GetConfig() (*model.HallOfFameConfig, error) {
	cfg, err := s.repo.GetConfig()
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		def := model.DefaultHallOfFameConfig()
		if err := s.repo.UpsertConfig(&def); err != nil {
			return nil, err
		}
		return &def, nil
	}
	return cfg, nil
}

// HofUpdateConfigRequest is the request body for updating temple config.
type HofUpdateConfigRequest struct {
	BackgroundImage *string `json:"background_image"` // nil = don't change
	CanvasWidth     *int    `json:"canvas_width"`
	CanvasHeight    *int    `json:"canvas_height"`
}

func (s *HallOfFameService) UpdateConfig(req *HofUpdateConfigRequest) (*model.HallOfFameConfig, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return nil, err
	}

	if req.BackgroundImage != nil {
		cfg.BackgroundImage = *req.BackgroundImage
	}
	if req.CanvasWidth != nil {
		if *req.CanvasWidth < 800 || *req.CanvasWidth > 7680 {
			return nil, errors.New("画布宽度必须在 800–7680 之间")
		}
		cfg.CanvasWidth = *req.CanvasWidth
	}
	if req.CanvasHeight != nil {
		if *req.CanvasHeight < 600 || *req.CanvasHeight > 4320 {
			return nil, errors.New("画布高度必须在 600–4320 之间")
		}
		cfg.CanvasHeight = *req.CanvasHeight
	}

	if err := s.repo.UpsertConfig(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// ─── Temple (public) ───

// TempleResponse is the combined response for the public temple endpoint.
type TempleResponse struct {
	Config model.HallOfFameConfig `json:"config"`
	Cards  []model.HallOfFameCard `json:"cards"`
}

func (s *HallOfFameService) GetTemple() (*TempleResponse, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return nil, err
	}
	cards, err := s.repo.ListCards(true) // visible only
	if err != nil {
		return nil, err
	}
	if cards == nil {
		cards = []model.HallOfFameCard{}
	}
	return &TempleResponse{Config: *cfg, Cards: cards}, nil
}

// ─── Cards ───

// ListAllCards returns all cards including hidden ones (admin).
func (s *HallOfFameService) ListAllCards() ([]model.HallOfFameCard, error) {
	cards, err := s.repo.ListCards(false) // include hidden
	if err != nil {
		return nil, err
	}
	if cards == nil {
		cards = []model.HallOfFameCard{}
	}
	return cards, nil
}

// CreateCardRequest is the request body for creating a new card.
type CreateCardRequest struct {
	Name              string  `json:"name" binding:"required"`
	Title             string  `json:"title"`
	Description       string  `json:"description"`
	Avatar            string  `json:"avatar"`
	PosX              float64 `json:"pos_x"`
	PosY              float64 `json:"pos_y"`
	Width             int     `json:"width"`
	Height            int     `json:"height"`
	StylePreset       string  `json:"style_preset"`
	CustomBgColor     string  `json:"custom_bg_color"`
	CustomTextColor   string  `json:"custom_text_color"`
	CustomBorderColor string  `json:"custom_border_color"`
	FontSize          int     `json:"font_size"`
	ZIndex            int     `json:"z_index"`
	Visible           *bool   `json:"visible"`
}

func (s *HallOfFameService) CreateCard(req *CreateCardRequest) (*model.HallOfFameCard, error) {
	if req.Name == "" {
		return nil, errors.New("名称不能为空")
	}
	preset := req.StylePreset
	if preset == "" {
		preset = "gold"
	}
	if !isValidStylePreset(preset) {
		return nil, errors.New("无效的样式预设")
	}
	width := req.Width
	if width <= 0 {
		width = 200
	}
	visible := true
	if req.Visible != nil {
		visible = *req.Visible
	}

	card := &model.HallOfFameCard{
		Name:              req.Name,
		Title:             req.Title,
		Description:       req.Description,
		Avatar:            req.Avatar,
		PosX:              clampPercent(req.PosX),
		PosY:              clampPercent(req.PosY),
		Width:             width,
		Height:            req.Height,
		StylePreset:       preset,
		CustomBgColor:     req.CustomBgColor,
		CustomTextColor:   req.CustomTextColor,
		CustomBorderColor: req.CustomBorderColor,
		FontSize:          req.FontSize,
		ZIndex:            req.ZIndex,
		Visible:           visible,
	}
	if err := s.repo.CreateCard(card); err != nil {
		return nil, err
	}
	return card, nil
}

// UpdateCardRequest is the request body for updating an existing card.
type UpdateCardRequest struct {
	Name              *string  `json:"name"`
	Title             *string  `json:"title"`
	Description       *string  `json:"description"`
	Avatar            *string  `json:"avatar"`
	PosX              *float64 `json:"pos_x"`
	PosY              *float64 `json:"pos_y"`
	Width             *int     `json:"width"`
	Height            *int     `json:"height"`
	StylePreset       *string  `json:"style_preset"`
	CustomBgColor     *string  `json:"custom_bg_color"`
	CustomTextColor   *string  `json:"custom_text_color"`
	CustomBorderColor *string  `json:"custom_border_color"`
	FontSize          *int     `json:"font_size"`
	ZIndex            *int     `json:"z_index"`
	Visible           *bool    `json:"visible"`
}

// CardLayoutUpdateRequest is the request body for batch layout saves.
type CardLayoutUpdateRequest struct {
	ID     uint    `json:"id"`
	PosX   float64 `json:"pos_x"`
	PosY   float64 `json:"pos_y"`
	Width  *int    `json:"width"`
	Height *int    `json:"height"`
	ZIndex int     `json:"z_index"`
}

func (s *HallOfFameService) UpdateCard(id uint, req *UpdateCardRequest) (*model.HallOfFameCard, error) {
	card, err := s.repo.GetCardByID(id)
	if err != nil {
		return nil, errors.New("卡片不存在")
	}

	updates, err := buildHallOfFameCardUpdateMap(req)
	if err != nil {
		return nil, err
	}
	if len(updates) == 0 {
		return card, nil
	}

	if err := s.repo.UpdateCardFields(id, updates); err != nil {
		return nil, err
	}

	return s.repo.GetCardByID(id)
}

// DeleteCard soft-deletes a card.
func (s *HallOfFameService) DeleteCard(id uint) error {
	if _, err := s.repo.GetCardByID(id); err != nil {
		return errors.New("卡片不存在")
	}
	return s.repo.DeleteCard(id)
}

// BatchUpdateLayout saves positions, size, and z-index for multiple cards.
func (s *HallOfFameService) BatchUpdateLayout(requests []CardLayoutUpdateRequest) error {
	updates, err := buildHallOfFameLayoutUpdates(requests)
	if err != nil {
		return err
	}
	if len(updates) == 0 {
		return nil
	}
	return s.repo.BatchUpdateLayout(updates)
}

// ─── Helpers ───

func clampPercent(v float64) float64 {
	return math.Max(0, math.Min(100, v))
}

var validStylePresets = map[string]bool{
	"gold": true, "silver": true, "bronze": true, "custom": true,
}

func isValidStylePreset(s string) bool {
	return validStylePresets[s]
}

func buildHallOfFameCardUpdateMap(req *UpdateCardRequest) (map[string]interface{}, error) {
	updates := map[string]interface{}{}

	if req.PosX != nil || req.PosY != nil || req.Width != nil || req.Height != nil || req.ZIndex != nil {
		return nil, errors.New("布局字段必须通过批量布局接口保存")
	}

	if req.Name != nil {
		if *req.Name == "" {
			return nil, errors.New("名称不能为空")
		}
		updates["name"] = *req.Name
	}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}
	if req.StylePreset != nil {
		if !isValidStylePreset(*req.StylePreset) {
			return nil, errors.New("无效的样式预设")
		}
		updates["style_preset"] = *req.StylePreset
	}
	if req.CustomBgColor != nil {
		updates["custom_bg_color"] = *req.CustomBgColor
	}
	if req.CustomTextColor != nil {
		updates["custom_text_color"] = *req.CustomTextColor
	}
	if req.CustomBorderColor != nil {
		updates["custom_border_color"] = *req.CustomBorderColor
	}
	if req.FontSize != nil {
		updates["font_size"] = *req.FontSize
	}
	if req.Visible != nil {
		updates["visible"] = *req.Visible
	}

	return updates, nil
}

func buildHallOfFameLayoutUpdates(requests []CardLayoutUpdateRequest) ([]model.CardLayoutUpdate, error) {
	if len(requests) == 0 {
		return nil, nil
	}

	updates := make([]model.CardLayoutUpdate, 0, len(requests))
	for _, req := range requests {
		if req.ID == 0 {
			return nil, errors.New("卡片 ID 不能为空")
		}
		if req.Width == nil || req.Height == nil {
			return nil, errors.New("卡片尺寸不能为空")
		}
		if *req.Width <= 0 {
			return nil, errors.New("卡片宽度必须大于 0")
		}
		if *req.Height < 0 {
			return nil, errors.New("卡片高度不能小于 0")
		}

		updates = append(updates, model.CardLayoutUpdate{
			ID:     req.ID,
			PosX:   clampPercent(req.PosX),
			PosY:   clampPercent(req.PosY),
			Width:  *req.Width,
			Height: *req.Height,
			ZIndex: req.ZIndex,
		})
	}

	return updates, nil
}
