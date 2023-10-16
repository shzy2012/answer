package checker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsInvalidUsername(t *testing.T) {

	assert.Equal(t, false, IsInvalidUsername("a'lan"))
	assert.Equal(t, false, IsInvalidUsername("HoLo_Debris"))
	assert.Equal(t, false, IsInvalidUsername("Alice-v"))
	assert.Equal(t, false, IsInvalidUsername("wallace.lai"))
	assert.Equal(t, false, IsInvalidUsername("Zhang（张三丰） SanFeng"))
	assert.Equal(t, false, IsInvalidUsername("Zhang(张三丰) SanFeng"))
	assert.Equal(t, false, IsInvalidUsername("·lu_2023"))
	assert.Equal(t, false, IsInvalidUsername("shuaoz Shuao Zhang"))
	assert.Equal(t, false, IsInvalidUsername("十进"))
	assert.Equal(t, false, IsInvalidUsername("PRIVACY"))
}
