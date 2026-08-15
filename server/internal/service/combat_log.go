package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const MaxCombatLogSize = 20 << 20 // 20 MiB

var (
	ErrCombatLogForbidden = errors.New("无权查看其他成员的战斗日志")
	ErrCombatLogNotFound  = errors.New("战斗日志分析不存在")
)

type CombatLogService struct {
	repo    *repository.CombatLogRepository
	sdeRepo *repository.SdeRepository
}

func NewCombatLogService() *CombatLogService {
	return &CombatLogService{repo: repository.NewCombatLogRepository(), sdeRepo: repository.NewSdeRepository()}
}

type CombatLogListResult struct {
	List     []repository.CombatLogReportItem `json:"list"`
	Total    int64                            `json:"total"`
	Page     int                              `json:"page"`
	PageSize int                              `json:"page_size"`
}

func (s *CombatLogService) Upload(userID uint, filename string, data []byte) (*repository.CombatLogDetail, error) {
	if len(data) == 0 {
		return nil, errors.New("上传的日志为空")
	}
	if len(data) > MaxCombatLogSize {
		return nil, fmt.Errorf("日志大小不能超过 %d MB", MaxCombatLogSize>>20)
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".txt" && ext != ".log" {
		return nil, errors.New("仅支持 .txt 或 .log 文件")
	}
	hashBytes := sha256.Sum256(data)
	hash := hex.EncodeToString(hashBytes[:])
	existing, findErr := s.repo.FindByUserAndHash(userID, hash)
	if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("检查重复日志失败: %w", findErr)
	}

	parsed, err := parseCombatLog(data)
	if err != nil {
		return nil, err
	}
	if findErr == nil {
		_, _, hitQualities := buildCombatAnalysis(parsed, &model.CombatLogReport{})
		if err := s.repo.ReplaceHitQualities(existing.ID, hitQualities); err != nil {
			return nil, fmt.Errorf("更新命中质量分析失败: %w", err)
		}
		return s.Get(userID, existing.ID, false)
	}
	duration := int64(parsed.EndedAt.Sub(parsed.StartedAt).Seconds())
	if duration < 1 {
		duration = 1
	}
	report := &model.CombatLogReport{
		UserID: userID, OriginalFilename: filepath.Base(filename), FileHash: hash,
		CharacterName: parsed.CharacterName, Language: parsed.Language,
		StartedAt: parsed.StartedAt, EndedAt: parsed.EndedAt, DurationSeconds: duration,
		SourceBytes: int64(len(data)),
	}
	breakdowns, timeline, hitQualities := buildCombatAnalysis(parsed, report)
	s.enrichMiningVolume(parsed.Language, report, breakdowns)
	if err := s.repo.Create(report, breakdowns, timeline, hitQualities); err != nil {
		return nil, fmt.Errorf("保存日志分析失败: %w", err)
	}
	return s.Get(userID, report.ID, false)
}

func (s *CombatLogService) enrichMiningVolume(language string, report *model.CombatLogReport, breakdowns []model.CombatLogBreakdown) {
	names := make([]string, 0)
	seen := make(map[string]struct{})
	for _, row := range breakdowns {
		if row.Category == combatCategoryMining && row.WeaponType != "" {
			if _, exists := seen[row.WeaponType]; !exists {
				seen[row.WeaponType] = struct{}{}
				names = append(names, row.WeaponType)
			}
		}
	}
	if len(names) == 0 {
		return
	}
	languageID := map[string]string{"english": "en", "chinese": "zh", "russian": "ru", "french": "fr", "german": "de", "japanese": "ja"}[language]
	volumes, err := s.sdeRepo.GetTypeVolumesByNames(names, languageID)
	if err != nil {
		global.Logger.Warn("战斗日志矿石体积解析失败", zap.String("language", language), zap.Error(err))
		return
	}
	for i := range breakdowns {
		if breakdowns[i].Category != combatCategoryMining {
			continue
		}
		breakdowns[i].VolumeM3 = float64(breakdowns[i].Amount) * volumes[breakdowns[i].WeaponType]
		report.MiningM3 += breakdowns[i].VolumeM3
	}
}

func (s *CombatLogService) List(userID uint, includeAll, admin bool, page, pageSize int) (*CombatLogListResult, error) {
	if includeAll && !admin {
		return nil, ErrCombatLogForbidden
	}
	items, total, err := s.repo.List(userID, includeAll, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询战斗日志失败: %w", err)
	}
	return &CombatLogListResult{List: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *CombatLogService) Get(userID, reportID uint, admin bool) (*repository.CombatLogDetail, error) {
	detail, err := s.repo.Get(reportID)
	if err != nil {
		if repository.IsCombatLogNotFound(err) {
			return nil, ErrCombatLogNotFound
		}
		return nil, fmt.Errorf("查询战斗日志详情失败: %w", err)
	}
	if !admin && detail.Report.UserID != userID {
		return nil, ErrCombatLogForbidden
	}
	return detail, nil
}
