package controller

import (
	"fmt"

	"github.com/answerdev/answer/internal/base/handler"
	"github.com/answerdev/answer/internal/base/middleware"
	"github.com/answerdev/answer/internal/base/reason"
	"github.com/answerdev/answer/internal/base/translator"
	"github.com/answerdev/answer/internal/base/validator"
	"github.com/answerdev/answer/internal/entity"
	"github.com/answerdev/answer/internal/schema"
	"github.com/answerdev/answer/internal/service"
	"github.com/answerdev/answer/internal/service/action"
	"github.com/answerdev/answer/internal/service/permission"
	"github.com/answerdev/answer/internal/service/rank"
	"github.com/answerdev/answer/pkg/uid"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/errors"
)

// AnswerController answer controller
type AnswerController struct {
	answerService *service.AnswerService
	rankService   *rank.RankService
	actionService *action.CaptchaService
}

// NewAnswerController new controller
func NewAnswerController(
	answerService *service.AnswerService,
	rankService *rank.RankService,
	actionService *action.CaptchaService,
) *AnswerController {
	return &AnswerController{
		answerService: answerService,
		rankService:   rankService,
		actionService: actionService,
	}
}

// RemoveAnswer delete answer
// @Summary delete answer
// @Description delete answer
// @Tags api-answer
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.RemoveAnswerReq true "answer"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/answer [delete]
func (ac *AnswerController) RemoveAnswer(ctx *gin.Context) {
	req := &schema.RemoveAnswerReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.ID = uid.DeShortID(req.ID)
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin {
		captchaPass := ac.actionService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionDelete, req.UserID, req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(handler.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}

	objectOwner := ac.rankService.CheckOperationObjectOwner(ctx, req.UserID, req.ID)
	canList, err := ac.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.AnswerDelete,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	req.CanDelete = canList[0] || objectOwner
	if !req.CanDelete {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	err = ac.answerService.RemoveAnswer(ctx, req)
	if !isAdmin {
		ac.actionService.ActionRecordAdd(ctx, entity.CaptchaActionDelete, req.UserID)
	}
	handler.HandleResponse(ctx, err, nil)
}

// Get godoc
// @Summary Get Answer
// @Description Get Answer
// @Tags api-answer
// @Accept  json
// @Produce  json
// @Param id query string true "Answer TagID"  default(1)
// @Router  /answer/api/v1/answer/info [get]
// @Success 200 {string} string ""
func (ac *AnswerController) Get(ctx *gin.Context) {
	id := ctx.Query("id")
	id = uid.DeShortID(id)
	userID := middleware.GetLoginUserIDFromContext(ctx)

	info, questionInfo, has, err := ac.answerService.Get(ctx, id, userID)
	if err != nil {
		handler.HandleResponse(ctx, err, gin.H{})
		return
	}
	if !has {
		handler.HandleResponse(ctx, fmt.Errorf(""), gin.H{})
		return
	}
	handler.HandleResponse(ctx, err, gin.H{
		"info":     info,
		"question": questionInfo,
	})
}

// Add godoc
// @Summary Insert Answer
// @Description Insert Answer
// @Tags api-answer
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param data body schema.AnswerAddReq  true "AnswerAddReq"
// @Success 200 {string} string ""
// @Router /answer/api/v1/answer [post]
func (ac *AnswerController) Add(ctx *gin.Context) {
	req := &schema.AnswerAddReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.QuestionID = uid.DeShortID(req.QuestionID)
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	canList, err := ac.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.AnswerEdit,
		permission.AnswerDelete,
		permission.LinkUrlLimit,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}

	linkUrlLimitUser := canList[2]
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin || !linkUrlLimitUser {
		captchaPass := ac.actionService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionAnswer, req.UserID, req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(handler.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}

	can, err := ac.rankService.CheckOperationPermission(ctx, req.UserID, permission.AnswerAdd, "")
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !can {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	answerID, err := ac.answerService.Insert(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !isAdmin || !linkUrlLimitUser {
		ac.actionService.ActionRecordAdd(ctx, entity.CaptchaActionAnswer, req.UserID)
	}
	info, questionInfo, has, err := ac.answerService.Get(ctx, answerID, req.UserID)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !has {
		handler.HandleResponse(ctx, nil, nil)
		return
	}

	objectOwner := ac.rankService.CheckOperationObjectOwner(ctx, req.UserID, info.ID)
	req.CanEdit = canList[0] || objectOwner
	req.CanDelete = canList[1] || objectOwner
	if !can {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}
	info.MemberActions = permission.GetAnswerPermission(ctx, req.UserID, info.UserID, req.CanEdit, req.CanDelete)
	handler.HandleResponse(ctx, nil, gin.H{
		"info":     info,
		"question": questionInfo,
	})
}

// Update godoc
// @Summary Update Answer
// @Description Update Answer
// @Tags api-answer
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param data body schema.AnswerUpdateReq  true "AnswerUpdateReq"
// @Success 200 {string} string ""
// @Router /answer/api/v1/answer [put]
func (ac *AnswerController) Update(ctx *gin.Context) {
	req := &schema.AnswerUpdateReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	canList, err := ac.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.AnswerEdit,
		permission.AnswerEditWithoutReview,
		permission.LinkUrlLimit,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	req.QuestionID = uid.DeShortID(req.QuestionID)
	linkUrlLimitUser := canList[2]
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin || !linkUrlLimitUser {
		captchaPass := ac.actionService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionEdit, req.UserID, req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(handler.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}

	objectOwner := ac.rankService.CheckOperationObjectOwner(ctx, req.UserID, req.ID)
	req.CanEdit = canList[0] || objectOwner
	req.NoNeedReview = canList[1] || objectOwner
	if !req.CanEdit {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	_, err = ac.answerService.Update(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !isAdmin || !linkUrlLimitUser {
		ac.actionService.ActionRecordAdd(ctx, entity.CaptchaActionEdit, req.UserID)
	}
	_, _, _, err = ac.answerService.Get(ctx, req.ID, req.UserID)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	handler.HandleResponse(ctx, nil, &schema.AnswerUpdateResp{WaitForReview: !req.NoNeedReview})
}

// AnswerList godoc
// @Summary AnswerList
// @Description AnswerList <br> <b>order</b> (default or updated)
// @Tags api-answer
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param question_id query string true "question_id"
// @Param order query string true "order"
// @Param page query string true "page"
// @Param page_size query string true "page_size"
// @Success 200 {string} string ""
// @Router /answer/api/v1/answer/page [get]
func (ac *AnswerController) AnswerList(ctx *gin.Context) {
	req := &schema.AnswerListReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	req.QuestionID = uid.DeShortID(req.QuestionID)

	canList, err := ac.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.AnswerEdit,
		permission.AnswerDelete,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	req.CanEdit = canList[0]
	req.CanDelete = canList[1]

	list, count, err := ac.answerService.SearchList(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	handler.HandleResponse(ctx, nil, gin.H{
		"list":  list,
		"count": count,
	})
}

// Accepted godoc
// @Summary Accepted
// @Description Accepted
// @Tags api-answer
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param data body schema.AcceptAnswerReq  true "AcceptAnswerReq"
// @Success 200 {string} string ""
// @Router /answer/api/v1/answer/acceptance [post]
func (ac *AnswerController) Accepted(ctx *gin.Context) {
	req := &schema.AcceptAnswerReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	req.AnswerID = uid.DeShortID(req.AnswerID)
	req.QuestionID = uid.DeShortID(req.QuestionID)
	can, err := ac.rankService.CheckOperationPermission(ctx, req.UserID, permission.AnswerAccept, req.QuestionID)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !can {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	err = ac.answerService.AcceptAnswer(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// AdminSetAnswerStatus godoc
// @Summary AdminSetAnswerStatus
// @Description Status:[available,deleted]
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.AdminSetAnswerStatusRequest true "AdminSetAnswerStatusRequest"
// @Router /answer/admin/api/answer/status [put]
// @Success 200 {object} handler.RespBody
func (ac *AnswerController) AdminSetAnswerStatus(ctx *gin.Context) {
	req := &schema.AdminSetAnswerStatusRequest{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.AnswerID = uid.DeShortID(req.AnswerID)

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	err := ac.answerService.AdminSetAnswerStatus(ctx, req)
	handler.HandleResponse(ctx, err, gin.H{})
}
