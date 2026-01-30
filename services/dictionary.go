package services

import (
	"context"
	stderrors "errors"

	"github.com/lyuangg/gadmin/errors"
	"github.com/lyuangg/gadmin/models"

	"gorm.io/gorm"
)

// DictionaryService 字典服务
type DictionaryService struct {
	ctx ServiceContext
}

// NewDictionaryService 创建字典服务实例
func NewDictionaryService(ctx ServiceContext) *DictionaryService {
	return &DictionaryService{ctx: ctx}
}

// GetTypes 获取字典类型列表（分页和筛选）
func (s *DictionaryService) GetTypes(ctx context.Context, page, pageSize int, filters map[string]string) ([]models.DictType, int64, error) {
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
	var list []models.DictType

	query := s.ctx.DB().Model(&models.DictType{})

	if code := filters["code"]; code != "" {
		query = query.Where("code LIKE ?", "%"+code+"%")
	}
	if name := filters["name"]; name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	orderBy := filters["order_by"]
	if orderBy == "id" || orderBy == "id_asc" {
		query = query.Order("id ASC")
	} else if orderBy == "id_desc" {
		query = query.Order("id DESC")
	} else {
		query = query.Order("id DESC")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// CreateType 创建字典类型
func (s *DictionaryService) CreateType(ctx context.Context, code, name, remark string) (*models.DictType, error) {
	var existing models.DictType
	if err := s.ctx.DB().Where("code = ?", code).First(&existing).Error; err != nil {
		if !stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	} else {
		return nil, errors.BadRequestMsg("字典类型编码已存在")
	}

	dt := models.DictType{
		Code:   code,
		Name:   name,
		Remark: remark,
	}
	if err := s.ctx.DB().Create(&dt).Error; err != nil {
		return nil, err
	}
	return &dt, nil
}

// UpdateType 更新字典类型
func (s *DictionaryService) UpdateType(ctx context.Context, id uint, code, name, remark string) (*models.DictType, error) {
	var dt models.DictType
	if err := s.ctx.DB().Where("id = ?", id).First(&dt).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFoundMsg("字典类型不存在")
		}
		return nil, err
	}

	if code != "" {
		var other models.DictType
		if err := s.ctx.DB().Where("code = ? AND id != ?", code, id).First(&other).Error; err != nil {
			if !stderrors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
		} else {
			return nil, errors.BadRequestMsg("字典类型编码已存在")
		}
		dt.Code = code
	}
	if name != "" {
		dt.Name = name
	}
	dt.Remark = remark

	if err := s.ctx.DB().Save(&dt).Error; err != nil {
		return nil, err
	}
	return &dt, nil
}

// DeleteType 删除字典类型（同时删除其下所有字典项）
func (s *DictionaryService) DeleteType(ctx context.Context, id uint) error {
	var dt models.DictType
	if err := s.ctx.DB().Where("id = ?", id).First(&dt).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFoundMsg("字典类型不存在")
		}
		return err
	}

	// 先删除该类型下所有字典项
	if err := s.ctx.DB().Where("type_id = ?", id).Delete(&models.DictItem{}).Error; err != nil {
		return err
	}
	if err := s.ctx.DB().Delete(&dt).Error; err != nil {
		return err
	}
	return nil
}

// GetItems 获取字典项列表（按 type_id 或 type_code 筛选，支持分页）
func (s *DictionaryService) GetItems(ctx context.Context, typeID uint, typeCode string, page, pageSize int, filters map[string]string) ([]models.DictItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := s.ctx.DB().Model(&models.DictItem{})

	if typeID > 0 {
		query = query.Where("type_id = ?", typeID)
	} else if typeCode != "" {
		var dt models.DictType
		if err := s.ctx.DB().Where("code = ?", typeCode).First(&dt).Error; err != nil {
			if stderrors.Is(err, gorm.ErrRecordNotFound) {
				return nil, 0, errors.NotFoundMsg("字典类型不存在")
			}
			return nil, 0, err
		}
		query = query.Where("type_id = ?", dt.ID)
	} else {
		return nil, 0, errors.BadRequestMsg("请指定 type_id 或 type_code")
	}

	if label := filters["label"]; label != "" {
		query = query.Where("label LIKE ?", "%"+label+"%")
	}
	if value := filters["value"]; value != "" {
		query = query.Where("value LIKE ?", "%"+value+"%")
	}

	query = query.Order("sort ASC, id ASC")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []models.DictItem
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// GetItemsByCode 根据类型编码获取所有启用的字典项（不分页，供下拉等使用）
func (s *DictionaryService) GetItemsByCode(ctx context.Context, typeCode string) ([]models.DictItem, error) {
	var dt models.DictType
	if err := s.ctx.DB().Where("code = ?", typeCode).First(&dt).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFoundMsg("字典类型不存在")
		}
		return nil, err
	}

	var list []models.DictItem
	if err := s.ctx.DB().Where("type_id = ? AND status = ?", dt.ID, 1).Order("sort ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// CreateItem 创建字典项
func (s *DictionaryService) CreateItem(ctx context.Context, typeID uint, label, value string, sort int, status int, remark string) (*models.DictItem, error) {
	var dt models.DictType
	if err := s.ctx.DB().Where("id = ?", typeID).First(&dt).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFoundMsg("字典类型不存在")
		}
		return nil, err
	}

	// 同类型下 value 唯一
	var existing models.DictItem
	if err := s.ctx.DB().Where("type_id = ? AND value = ?", typeID, value).First(&existing).Error; err != nil {
		if !stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	} else {
		return nil, errors.BadRequestMsg("该类型下字典项值已存在")
	}

	item := models.DictItem{
		TypeID: typeID,
		Label:  label,
		Value:  value,
		Sort:   sort,
		Status: status,
		Remark: remark,
	}
	if err := s.ctx.DB().Create(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// UpdateItem 更新字典项
func (s *DictionaryService) UpdateItem(ctx context.Context, id uint, label, value string, sort *int, status *int, remark string) (*models.DictItem, error) {
	var item models.DictItem
	if err := s.ctx.DB().Where("id = ?", id).First(&item).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFoundMsg("字典项不存在")
		}
		return nil, err
	}

	if label != "" {
		item.Label = label
	}
	if value != "" {
		var other models.DictItem
		if err := s.ctx.DB().Where("type_id = ? AND value = ? AND id != ?", item.TypeID, value, id).First(&other).Error; err != nil {
			if !stderrors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
		} else {
			return nil, errors.BadRequestMsg("该类型下字典项值已存在")
		}
		item.Value = value
	}
	if sort != nil {
		item.Sort = *sort
	}
	if status != nil {
		item.Status = *status
	}
	item.Remark = remark

	if err := s.ctx.DB().Save(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// DeleteItem 删除字典项
func (s *DictionaryService) DeleteItem(ctx context.Context, id uint) error {
	var item models.DictItem
	if err := s.ctx.DB().Where("id = ?", id).First(&item).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFoundMsg("字典项不存在")
		}
		return err
	}
	return s.ctx.DB().Delete(&item).Error
}
