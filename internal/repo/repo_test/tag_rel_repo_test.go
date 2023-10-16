package repo_test

import (
	"context"
	"log"
	"sync"
	"testing"

	"github.com/answerdev/answer/internal/repo/unique"

	"github.com/answerdev/answer/internal/entity"
	"github.com/answerdev/answer/internal/repo/tag"
	"github.com/stretchr/testify/assert"
)

var (
	tagRelOnce     sync.Once
	testTagRelList = []*entity.TagRel{
		{
			ObjectID: "10010000000000001",
			TagID:    "10030000000000001",
			Status:   entity.TagRelStatusAvailable,
		},
		{
			ObjectID: "10010000000000002",
			TagID:    "10030000000000002",
			Status:   entity.TagRelStatusAvailable,
		},
	}
)

func addTagRelList() {
	tagRelRepo := tag.NewTagRelRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
	err := tagRelRepo.AddTagRelList(context.TODO(), testTagRelList)
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func Test_tagListRepo_BatchGetObjectTagRelList(t *testing.T) {
	tagRelOnce.Do(addTagRelList)
	tagRelRepo := tag.NewTagRelRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
	relList, err :=
		tagRelRepo.BatchGetObjectTagRelList(context.TODO(), []string{testTagRelList[0].ObjectID, testTagRelList[1].ObjectID})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(relList))
}

func Test_tagListRepo_CountTagRelByTagID(t *testing.T) {
	tagRelOnce.Do(addTagRelList)
	tagRelRepo := tag.NewTagRelRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
	count, err := tagRelRepo.CountTagRelByTagID(context.TODO(), "10030000000000001")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func Test_tagListRepo_GetObjectTagRelList(t *testing.T) {
	tagRelOnce.Do(addTagRelList)
	tagRelRepo := tag.NewTagRelRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	relList, err :=
		tagRelRepo.GetObjectTagRelList(context.TODO(), testTagRelList[0].ObjectID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(relList))
}

func Test_tagListRepo_GetObjectTagRelWithoutStatus(t *testing.T) {
	tagRelOnce.Do(addTagRelList)
	tagRelRepo := tag.NewTagRelRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	relList, err :=
		tagRelRepo.BatchGetObjectTagRelList(context.TODO(), []string{testTagRelList[0].ObjectID, testTagRelList[1].ObjectID})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(relList))

	ids := []int64{relList[0].ID, relList[1].ID}
	err = tagRelRepo.RemoveTagRelListByIDs(context.TODO(), ids)
	assert.NoError(t, err)

	count, err := tagRelRepo.CountTagRelByTagID(context.TODO(), "10030000000000001")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	_, exist, err := tagRelRepo.GetObjectTagRelWithoutStatus(context.TODO(), relList[0].ObjectID, relList[0].TagID)
	assert.NoError(t, err)
	assert.True(t, exist)

	err = tagRelRepo.EnableTagRelByIDs(context.TODO(), ids)
	assert.NoError(t, err)

	count, err = tagRelRepo.CountTagRelByTagID(context.TODO(), "10030000000000001")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}
