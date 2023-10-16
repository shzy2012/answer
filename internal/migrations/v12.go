package migrations

import (
	"fmt"
	"time"

	"github.com/answerdev/answer/internal/entity"
	"github.com/segmentfault/pacman/log"
	"xorm.io/xorm"
)

type QuestionPostTime struct {
	ID               string    `xorm:"not null pk BIGINT(20) id"`
	CreatedAt        time.Time `xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt        time.Time `xorm:"updated_at TIMESTAMP"`
	UserID           string    `xorm:"not null default 0 BIGINT(20) INDEX user_id"`
	LastEditUserID   string    `xorm:"not null default 0 BIGINT(20) last_edit_user_id"`
	Title            string    `xorm:"not null default '' VARCHAR(150) title"`
	OriginalText     string    `xorm:"not null LONGTEXT original_text"`
	ParsedText       string    `xorm:"not null LONGTEXT parsed_text"`
	Status           int       `xorm:"not null default 1 INT(11) status"`
	Pin              int       `xorm:"not null default 1 INT(11) pin"`
	Show             int       `xorm:"not null default 1 INT(11) show"`
	ViewCount        int       `xorm:"not null default 0 INT(11) view_count"`
	UniqueViewCount  int       `xorm:"not null default 0 INT(11) unique_view_count"`
	VoteCount        int       `xorm:"not null default 0 INT(11) vote_count"`
	AnswerCount      int       `xorm:"not null default 0 INT(11) answer_count"`
	CollectionCount  int       `xorm:"not null default 0 INT(11) collection_count"`
	FollowCount      int       `xorm:"not null default 0 INT(11) follow_count"`
	AcceptedAnswerID string    `xorm:"not null default 0 BIGINT(20) accepted_answer_id"`
	LastAnswerID     string    `xorm:"not null default 0 BIGINT(20) last_answer_id"`
	PostUpdateTime   time.Time `xorm:"post_update_time TIMESTAMP"`
	RevisionID       string    `xorm:"not null default 0 BIGINT(20) revision_id"`
}

func (QuestionPostTime) TableName() string {
	return "question"
}

func updateQuestionPostTime(x *xorm.Engine) error {
	questionList := make([]QuestionPostTime, 0)
	err := x.Find(&questionList, &entity.Question{})
	if err != nil {
		return fmt.Errorf("get questions failed: %w", err)
	}
	for _, item := range questionList {
		if item.PostUpdateTime.IsZero() {
			if !item.UpdatedAt.IsZero() {
				item.PostUpdateTime = item.UpdatedAt
			} else if !item.CreatedAt.IsZero() {
				item.PostUpdateTime = item.CreatedAt
			}
			if _, err = x.Update(item, &QuestionPostTime{ID: item.ID}); err != nil {
				log.Errorf("update %+v config failed: %s", item, err)
				return fmt.Errorf("update question failed: %w", err)
			}
		}

	}

	return nil
}
