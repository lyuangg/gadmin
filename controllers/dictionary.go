package controllers

import (
	"strconv"

	"github.com/lyuangg/gadmin/app"
	"github.com/lyuangg/gadmin/errors"

	"github.com/gin-gonic/gin"
)

type DictionaryController struct {
	app *app.App
}

func NewDictionaryController(a *app.App) *DictionaryController {
	return &DictionaryController{app: a}
}

type getDictTypesQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	OrderBy  string `form:"order_by"`
	Code     string `form:"code"`
	Name     string `form:"name"`
}

func (ctrl *DictionaryController) GetTypes(c *gin.Context) {
	var req getDictTypesQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestErr(err))
		return
	}
	filters := map[string]string{
		"order_by": req.OrderBy,
		"code":     req.Code,
		"name":     req.Name,
	}
	list, total, err := ctrl.app.GetDictionaryService().GetTypes(c.Request.Context(), req.Page, req.PageSize, filters)
	if err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	totalPage := 0
	if req.PageSize > 0 {
		totalPage = (int(total) + req.PageSize - 1) / req.PageSize
	}
	ctrl.app.Responder.Success(c, gin.H{
		"data": list,
		"pagination": gin.H{
			"page":       req.Page,
			"page_size":  req.PageSize,
			"total":      total,
			"total_page": totalPage,
		},
	})
}

type CreateTypeRequest struct {
	Code   string `json:"code" binding:"required"`
	Name   string `json:"name" binding:"required"`
	Remark string `json:"remark"`
}

func (ctrl *DictionaryController) CreateType(c *gin.Context) {
	var req CreateTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestErr(err))
		return
	}

	dt, err := ctrl.app.GetDictionaryService().CreateType(c.Request.Context(), req.Code, req.Name, req.Remark)
	if err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}
	ctrl.app.Responder.SuccessWithMsg(c, "创建成功", dt)
}

type UpdateTypeRequest struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Remark string `json:"remark"`
}

func (ctrl *DictionaryController) UpdateType(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestMsg("无效的ID"))
		return
	}

	var req UpdateTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestErr(err))
		return
	}

	dt, err := ctrl.app.GetDictionaryService().UpdateType(c.Request.Context(), uint(id), req.Code, req.Name, req.Remark)
	if err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}
	ctrl.app.Responder.SuccessWithMsg(c, "更新成功", dt)
}

func (ctrl *DictionaryController) DeleteType(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestMsg("无效的ID"))
		return
	}

	if err := ctrl.app.GetDictionaryService().DeleteType(c.Request.Context(), uint(id)); err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}
	ctrl.app.Responder.SuccessWithMsg(c, "删除成功", nil)
}

type getDictItemsQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	TypeID   uint   `form:"type_id"`
	TypeCode string `form:"type_code"`
	Label    string `form:"label"`
	Value    string `form:"value"`
}

func (ctrl *DictionaryController) GetItems(c *gin.Context) {
	var req getDictItemsQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestErr(err))
		return
	}
	filters := map[string]string{
		"label": req.Label,
		"value": req.Value,
	}
	list, total, err := ctrl.app.GetDictionaryService().GetItems(c.Request.Context(), req.TypeID, req.TypeCode, req.Page, req.PageSize, filters)
	if err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}

	totalPage := 0
	if req.PageSize > 0 {
		totalPage = (int(total) + req.PageSize - 1) / req.PageSize
	}
	ctrl.app.Responder.Success(c, gin.H{
		"data": list,
		"pagination": gin.H{
			"page":       req.Page,
			"page_size":  req.PageSize,
			"total":      total,
			"total_page": totalPage,
		},
	})
}

func (ctrl *DictionaryController) GetItemsByCode(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		ctrl.app.Responder.RespondError(c, errors.BadRequestMsg("请提供 code 参数"))
		return
	}

	list, err := ctrl.app.GetDictionaryService().GetItemsByCode(c.Request.Context(), code)
	if err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}
	ctrl.app.Responder.Success(c, gin.H{"data": list})
}

type CreateItemRequest struct {
	TypeID uint   `json:"type_id" binding:"required"`
	Label  string `json:"label" binding:"required"`
	Value  string `json:"value" binding:"required"`
	Sort   int    `json:"sort"`
	Status int    `json:"status"`
	Remark string `json:"remark"`
}

func (ctrl *DictionaryController) CreateItem(c *gin.Context) {
	var req CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestErr(err))
		return
	}
	if req.Status != 0 && req.Status != 1 {
		req.Status = 1
	}

	item, err := ctrl.app.GetDictionaryService().CreateItem(c.Request.Context(), req.TypeID, req.Label, req.Value, req.Sort, req.Status, req.Remark)
	if err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}
	ctrl.app.Responder.SuccessWithMsg(c, "创建成功", item)
}

type UpdateItemRequest struct {
	Label  string `json:"label"`
	Value  string `json:"value"`
	Sort   *int   `json:"sort"`
	Status *int   `json:"status"`
	Remark string `json:"remark"`
}

func (ctrl *DictionaryController) UpdateItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestMsg("无效的ID"))
		return
	}

	var req UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestErr(err))
		return
	}

	item, err := ctrl.app.GetDictionaryService().UpdateItem(c.Request.Context(), uint(id), req.Label, req.Value, req.Sort, req.Status, req.Remark)
	if err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}
	ctrl.app.Responder.SuccessWithMsg(c, "更新成功", item)
}

func (ctrl *DictionaryController) DeleteItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctrl.app.Responder.RespondError(c, errors.BadRequestMsg("无效的ID"))
		return
	}

	if err := ctrl.app.GetDictionaryService().DeleteItem(c.Request.Context(), uint(id)); err != nil {
		ctrl.app.Responder.RespondError(c, err)
		return
	}
	ctrl.app.Responder.SuccessWithMsg(c, "删除成功", nil)
}
