package utils

import (
	"testing"

	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MockDatabase struct {
	database.DatabaseInterface
}

func TestGetDbFromContextErrDbNotFound(t *testing.T) {
	t.Parallel()

	ctx := gin.Context{}

	db, err := GetDbFromContext(&ctx)
	assert.EqualError(t, err, errDbNotFound.Error())
	assert.Nil(t, db)
}

func TestGetDbFromContextErrDbWrongType(t *testing.T) {
	t.Parallel()

	ctx := gin.Context{}
	ctx.Set("db", "bogus")

	db, err := GetDbFromContext(&ctx)
	assert.EqualError(t, err, errDbWrongType.Error())
	assert.Nil(t, db)
}

func TestGetDbFromContextSuccess(t *testing.T) {
	t.Parallel()

	ctx := gin.Context{}
	ctx.Set("db", MockDatabase{})

	db, err := GetDbFromContext(&ctx)
	assert.NoError(t, err)
	assert.NotNil(t, db)
}
