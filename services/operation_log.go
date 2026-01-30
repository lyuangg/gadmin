package services

import (
	"context"
	"time"

	"github.com/lyuangg/gadmin/models"
)

// OperationLogService 操作日志服务
type OperationLogService struct {
	ctx ServiceContext
}

// NewOperationLogService 创建操作日志服务实例
func NewOperationLogService(ctx ServiceContext) *OperationLogService {
	return &OperationLogService{ctx: ctx}
}

// GetOperationLogs 获取操作日志列表（分页和筛选）
// 支持按时间范围、用户名、方法、路径、状态码筛选
func (s *OperationLogService) GetOperationLogs(
	ctx context.Context,
	page, pageSize int,
	filters map[string]string,
) ([]models.OperationLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var total int64
	var logs []models.OperationLog

	query := s.ctx.DB().Model(&models.OperationLog{})

	// 时间范围筛选（created_at）
	if startStr, ok := filters["start_time"]; ok && startStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startStr); err == nil {
			query = query.Where("created_at >= ?", startTime)
		} else {
			s.ctx.Logger().WarnContext(ctx, "解析操作日志开始时间失败", "start_time", startStr, "error", err)
		}
	}
	if endStr, ok := filters["end_time"]; ok && endStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endStr); err == nil {
			query = query.Where("created_at <= ?", endTime)
		} else {
			s.ctx.Logger().WarnContext(ctx, "解析操作日志结束时间失败", "end_time", endStr, "error", err)
		}
	}

	// 其他筛选
	if username, ok := filters["username"]; ok && username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if method, ok := filters["method"]; ok && method != "" {
		query = query.Where("method = ?", method)
	}
	if path, ok := filters["path"]; ok && path != "" {
		query = query.Where("path LIKE ?", "%"+path+"%")
	}
	if statusCode, ok := filters["status_code"]; ok && statusCode != "" {
		query = query.Where("status_code = ?", statusCode)
	}

	// 排序：默认按 id 倒序（新记录在前）
	switch filters["order_by"] {
	case "id", "id_asc":
		query = query.Order("id ASC")
	case "id_desc":
		query = query.Order("id DESC")
	case "created_at", "created_at_asc":
		query = query.Order("created_at ASC")
	case "created_at_desc":
		query = query.Order("created_at DESC")
	default:
		query = query.Order("id DESC")
	}

	// 总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	s.fillLogNicknames(logs)
	return logs, total, nil
}

// CleanOldLogs 清理旧操作日志，仅保留最近 retain 条（按 id 倒序）。返回删除行数。
func (s *OperationLogService) CleanOldLogs(ctx context.Context, retain int) (int64, error) {
	if retain <= 0 {
		return 0, nil
	}
	var minIDToKeep uint
	err := s.ctx.DB().Model(&models.OperationLog{}).Order("id DESC").Offset(retain-1).Limit(1).Pluck("id", &minIDToKeep).Error
	if err != nil {
		return 0, err
	}
	result := s.ctx.DB().Unscoped().Where("id < ?", minIDToKeep).Delete(&models.OperationLog{})
	return result.RowsAffected, result.Error
}

// fillLogNicknames 按 UserID 批量查 users 表取昵称并填到 logs 的 Nickname（原地修改）
func (s *OperationLogService) fillLogNicknames(logs []models.OperationLog) {
	if len(logs) == 0 {
		return
	}
	userIDs := make([]uint, 0, len(logs))
	seen := make(map[uint]struct{})
	for _, l := range logs {
		if l.UserID > 0 {
			if _, ok := seen[l.UserID]; !ok {
				seen[l.UserID] = struct{}{}
				userIDs = append(userIDs, l.UserID)
			}
		}
	}
	if len(userIDs) == 0 {
		return
	}
	var users []struct {
		ID       uint   `gorm:"column:id"`
		Nickname string `gorm:"column:nickname"`
	}
	if err := s.ctx.DB().Model(&models.User{}).Where("id IN ?", userIDs).Select("id", "nickname").Find(&users).Error; err != nil {
		return
	}
	nicknames := make(map[uint]string)
	for _, u := range users {
		nicknames[u.ID] = u.Nickname
	}
	for i := range logs {
		logs[i].Nickname = nicknames[logs[i].UserID]
	}
}
