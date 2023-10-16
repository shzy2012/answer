package question

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/answerdev/answer/plugin"
	"github.com/segmentfault/pacman/log"
	"strings"
	"time"
	"unicode"

	"github.com/answerdev/answer/internal/base/handler"
	"xorm.io/builder"

	"github.com/answerdev/answer/internal/base/constant"
	"github.com/answerdev/answer/internal/base/data"
	"github.com/answerdev/answer/internal/base/pager"
	"github.com/answerdev/answer/internal/base/reason"
	"github.com/answerdev/answer/internal/entity"
	"github.com/answerdev/answer/internal/schema"
	questioncommon "github.com/answerdev/answer/internal/service/question_common"
	"github.com/answerdev/answer/internal/service/unique"
	"github.com/answerdev/answer/pkg/htmltext"
	"github.com/answerdev/answer/pkg/uid"

	"github.com/segmentfault/pacman/errors"
)

// questionRepo question repository
type questionRepo struct {
	data         *data.Data
	uniqueIDRepo unique.UniqueIDRepo
}

// NewQuestionRepo new repository
func NewQuestionRepo(
	data *data.Data,
	uniqueIDRepo unique.UniqueIDRepo,
) questioncommon.QuestionRepo {
	return &questionRepo{
		data:         data,
		uniqueIDRepo: uniqueIDRepo,
	}
}

// AddQuestion add question
func (qr *questionRepo) AddQuestion(ctx context.Context, question *entity.Question) (err error) {
	question.ID, err = qr.uniqueIDRepo.GenUniqueIDStr(ctx, question.TableName())
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	_, err = qr.data.DB.Context(ctx).Insert(question)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		question.ID = uid.EnShortID(question.ID)
	}
	_ = qr.updateSearch(ctx, question.ID)
	return
}

// RemoveQuestion delete question
func (qr *questionRepo) RemoveQuestion(ctx context.Context, id string) (err error) {
	id = uid.DeShortID(id)
	_, err = qr.data.DB.Context(ctx).Where("id =?", id).Delete(&entity.Question{})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateQuestion update question
func (qr *questionRepo) UpdateQuestion(ctx context.Context, question *entity.Question, Cols []string) (err error) {
	question.ID = uid.DeShortID(question.ID)
	_, err = qr.data.DB.Context(ctx).Where("id =?", question.ID).Cols(Cols...).Update(question)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		question.ID = uid.EnShortID(question.ID)
	}
	_ = qr.updateSearch(ctx, question.ID)
	return
}

func (qr *questionRepo) UpdatePvCount(ctx context.Context, questionID string) (err error) {
	questionID = uid.DeShortID(questionID)
	question := &entity.Question{}
	_, err = qr.data.DB.Context(ctx).Where("id =?", questionID).Incr("view_count", 1).Update(question)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	_ = qr.updateSearch(ctx, question.ID)
	return nil
}

func (qr *questionRepo) UpdateAnswerCount(ctx context.Context, questionID string, num int) (err error) {
	questionID = uid.DeShortID(questionID)
	question := &entity.Question{}
	question.AnswerCount = num
	_, err = qr.data.DB.Context(ctx).Where("id =?", questionID).Cols("answer_count").Update(question)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	_ = qr.updateSearch(ctx, question.ID)
	return nil
}

func (qr *questionRepo) UpdateCollectionCount(ctx context.Context, questionID string, num int) (err error) {
	questionID = uid.DeShortID(questionID)
	question := &entity.Question{}
	_, err = qr.data.DB.Context(ctx).Where("id =?", questionID).Incr("collection_count", num).Update(question)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (qr *questionRepo) UpdateQuestionStatus(ctx context.Context, question *entity.Question) (err error) {
	question.ID = uid.DeShortID(question.ID)
	now := time.Now()
	question.UpdatedAt = now
	_, err = qr.data.DB.Context(ctx).Where("id =?", question.ID).Cols("status", "updated_at").Update(question)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	_ = qr.updateSearch(ctx, question.ID)
	return nil
}

func (qr *questionRepo) UpdateQuestionStatusWithOutUpdateTime(ctx context.Context, question *entity.Question) (err error) {
	question.ID = uid.DeShortID(question.ID)
	_, err = qr.data.DB.Context(ctx).Where("id =?", question.ID).Cols("status").Update(question)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	_ = qr.updateSearch(ctx, question.ID)
	return nil
}

func (qr *questionRepo) UpdateQuestionOperation(ctx context.Context, question *entity.Question) (err error) {
	question.ID = uid.DeShortID(question.ID)
	_, err = qr.data.DB.Context(ctx).Where("id =?", question.ID).Cols("pin", "show").Update(question)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (qr *questionRepo) UpdateAccepted(ctx context.Context, question *entity.Question) (err error) {
	question.ID = uid.DeShortID(question.ID)
	_, err = qr.data.DB.Context(ctx).Where("id =?", question.ID).Cols("accepted_answer_id").Update(question)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	_ = qr.updateSearch(ctx, question.ID)
	return nil
}

func (qr *questionRepo) UpdateLastAnswer(ctx context.Context, question *entity.Question) (err error) {
	question.ID = uid.DeShortID(question.ID)
	_, err = qr.data.DB.Context(ctx).Where("id =?", question.ID).Cols("last_answer_id").Update(question)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	_ = qr.updateSearch(ctx, question.ID)
	return nil
}

// GetQuestion get question one
func (qr *questionRepo) GetQuestion(ctx context.Context, id string) (
	question *entity.Question, exist bool, err error,
) {
	id = uid.DeShortID(id)
	question = &entity.Question{}
	question.ID = id
	exist, err = qr.data.DB.Context(ctx).Where("id = ?", id).Get(question)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		question.ID = uid.EnShortID(question.ID)
	}
	return
}

// GetQuestionsByTitle get question list by title
func (qr *questionRepo) GetQuestionsByTitle(ctx context.Context, title string, pageSize int) (
	questionList []*entity.Question, err error) {
	questionList = make([]*entity.Question, 0)
	session := qr.data.DB.Context(ctx)
	session.Where("status != ?", entity.QuestionStatusDeleted)
	session.Where("title like ?", "%"+title+"%")
	session.Limit(pageSize)
	err = session.Find(&questionList)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		for _, item := range questionList {
			item.ID = uid.EnShortID(item.ID)
		}
	}
	return
}

func (qr *questionRepo) FindByID(ctx context.Context, id []string) (questionList []*entity.Question, err error) {
	for key, itemID := range id {
		id[key] = uid.DeShortID(itemID)
	}
	questionList = make([]*entity.Question, 0)
	err = qr.data.DB.Context(ctx).Table("question").In("id", id).Find(&questionList)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		for _, item := range questionList {
			item.ID = uid.EnShortID(item.ID)
		}
	}
	return
}

// GetQuestionList get question list all
func (qr *questionRepo) GetQuestionList(ctx context.Context, question *entity.Question) (questionList []*entity.Question, err error) {
	question.ID = uid.DeShortID(question.ID)
	questionList = make([]*entity.Question, 0)
	err = qr.data.DB.Context(ctx).Find(questionList, question)
	if err != nil {
		return questionList, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	for _, item := range questionList {
		item.ID = uid.DeShortID(item.ID)
	}
	return
}

func (qr *questionRepo) GetQuestionCount(ctx context.Context) (count int64, err error) {
	session := qr.data.DB.Context(ctx)
	session.In("status", []int{entity.QuestionStatusAvailable, entity.QuestionStatusClosed})
	count, err = session.Count(&entity.Question{Show: entity.QuestionShow})
	if err != nil {
		return 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return count, nil
}

func (qr *questionRepo) GetUserQuestionCount(ctx context.Context, userID string) (count int64, err error) {
	session := qr.data.DB.Context(ctx)
	session.In("status", []int{entity.QuestionStatusAvailable, entity.QuestionStatusClosed})
	count, err = session.Count(&entity.Question{UserID: userID})
	if err != nil {
		return count, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (qr *questionRepo) SitemapQuestions(ctx context.Context, page, pageSize int) (
	questionIDList []*schema.SiteMapQuestionInfo, err error) {
	page = page - 1
	questionIDList = make([]*schema.SiteMapQuestionInfo, 0)

	// try to get sitemap data from cache
	cacheKey := fmt.Sprintf(constant.SiteMapQuestionCacheKeyPrefix, page)
	cacheData, exist, err := qr.data.Cache.GetString(ctx, cacheKey)
	if err == nil && exist {
		_ = json.Unmarshal([]byte(cacheData), &questionIDList)
		return questionIDList, nil
	}

	// get sitemap data from db
	rows := make([]*entity.Question, 0)
	session := qr.data.DB.Context(ctx)
	session.Select("id,title,created_at,post_update_time")
	session.Where("`show` = ?", entity.QuestionShow)
	session.Where("status = ? OR status = ?", entity.QuestionStatusAvailable, entity.QuestionStatusClosed)
	session.Limit(pageSize, page*pageSize)
	session.Asc("created_at")
	err = session.Find(&rows)
	if err != nil {
		return questionIDList, err
	}

	// warp data
	for _, question := range rows {
		item := &schema.SiteMapQuestionInfo{ID: question.ID}
		if handler.GetEnableShortID(ctx) {
			item.ID = uid.EnShortID(question.ID)
		}
		item.Title = htmltext.UrlTitle(question.Title)
		if question.PostUpdateTime.IsZero() {
			item.UpdateTime = question.CreatedAt.Format(time.RFC3339)
		} else {
			item.UpdateTime = question.PostUpdateTime.Format(time.RFC3339)
		}
		questionIDList = append(questionIDList, item)
	}

	// set sitemap data to cache
	cacheDataByte, _ := json.Marshal(questionIDList)
	if err := qr.data.Cache.SetString(ctx, cacheKey, string(cacheDataByte), constant.SiteMapQuestionCacheTime); err != nil {
		log.Error(err)
	}
	return questionIDList, nil
}

// GetQuestionPage query question page
func (qr *questionRepo) GetQuestionPage(ctx context.Context, page, pageSize int, userID, tagID, orderCond string, inDays int) (
	questionList []*entity.Question, total int64, err error) {
	questionList = make([]*entity.Question, 0)

	session := qr.data.DB.Context(ctx).Where("question.status = ? OR question.status = ?",
		entity.QuestionStatusAvailable, entity.QuestionStatusClosed)
	if len(tagID) > 0 {
		session.Join("LEFT", "tag_rel", "question.id = tag_rel.object_id")
		session.And("tag_rel.tag_id = ?", tagID)
		session.And("tag_rel.status = ?", entity.TagRelStatusAvailable)
	}
	if len(userID) > 0 {
		session.And("question.user_id = ?", userID)
	} else {
		session.And("question.show = ?", entity.QuestionShow)
	}
	if inDays > 0 {
		session.And("question.created_at > ?", time.Now().AddDate(0, 0, -inDays))
	}

	switch orderCond {
	case "newest":
		session.OrderBy("question.pin desc,question.created_at DESC")
	case "active":
		session.OrderBy("question.pin desc,question.post_update_time DESC, question.updated_at DESC")
	case "frequent":
		session.OrderBy("question.pin desc,question.view_count DESC")
	case "score":
		session.OrderBy("question.pin desc,question.vote_count DESC, question.view_count DESC")
	case "unanswered":
		session.Where("question.last_answer_id = 0")
		session.OrderBy("question.pin desc,question.created_at DESC")
	}

	total, err = pager.Help(page, pageSize, &questionList, &entity.Question{}, session)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		for _, item := range questionList {
			item.ID = uid.EnShortID(item.ID)
		}
	}
	return questionList, total, err
}

func (qr *questionRepo) AdminQuestionPage(ctx context.Context, search *schema.AdminQuestionPageReq) ([]*entity.Question, int64, error) {
	var (
		count   int64
		err     error
		session = qr.data.DB.Context(ctx).Table("question")
	)

	session.Where(builder.Eq{
		"status": search.Status,
	})

	rows := make([]*entity.Question, 0)
	if search.Page > 0 {
		search.Page = search.Page - 1
	} else {
		search.Page = 0
	}
	if search.PageSize == 0 {
		search.PageSize = constant.DefaultPageSize
	}

	// search by question title like or question id
	if len(search.Query) > 0 {
		// check id search
		var (
			idSearch = false
			id       = ""
		)

		if strings.Contains(search.Query, "question:") {
			idSearch = true
			id = strings.TrimSpace(strings.TrimPrefix(search.Query, "question:"))
			id = uid.DeShortID(id)
			for _, r := range id {
				if !unicode.IsDigit(r) {
					idSearch = false
					break
				}
			}
		}

		if idSearch {
			session.And(builder.Eq{
				"id": id,
			})
		} else {
			session.And(builder.Like{
				"title", search.Query,
			})
		}
	}

	offset := search.Page * search.PageSize

	session.OrderBy("created_at desc").
		Limit(search.PageSize, offset)
	count, err = session.FindAndCount(&rows)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return rows, count, err
	}
	if handler.GetEnableShortID(ctx) {
		for _, item := range rows {
			item.ID = uid.EnShortID(item.ID)
		}
	}
	return rows, count, nil
}

// updateSearch update search, if search plugin not enable, do nothing
func (qr *questionRepo) updateSearch(ctx context.Context, questionID string) (err error) {
	// check search plugin
	var s plugin.Search
	_ = plugin.CallSearch(func(search plugin.Search) error {
		s = search
		return nil
	})
	if s == nil {
		return
	}
	questionID = uid.DeShortID(questionID)
	question, exist, err := qr.GetQuestion(ctx, questionID)
	if !exist {
		return
	}
	if err != nil {
		return err
	}

	// get tags
	var (
		tagListList = make([]*entity.TagRel, 0)
		tags        = make([]string, 0)
	)
	session := qr.data.DB.Context(ctx).Where("object_id = ?", questionID)
	session.Where("status = ?", entity.TagRelStatusAvailable)
	err = session.Find(&tagListList)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	for _, tag := range tagListList {
		tags = append(tags, tag.TagID)
	}
	content := &plugin.SearchContent{
		ObjectID:    questionID,
		Title:       question.Title,
		Type:        constant.QuestionObjectType,
		Content:     question.ParsedText,
		Answers:     int64(question.AnswerCount),
		Status:      plugin.SearchContentStatus(question.Status),
		Tags:        tags,
		QuestionID:  questionID,
		UserID:      question.UserID,
		Views:       int64(question.ViewCount),
		Created:     question.CreatedAt.Unix(),
		Active:      question.UpdatedAt.Unix(),
		Score:       int64(question.VoteCount),
		HasAccepted: question.AcceptedAnswerID != "" && question.AcceptedAnswerID != "0",
	}
	err = s.UpdateContent(ctx, content)
	return
}
