package netrc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNetRC(t *testing.T) {
	n, err := parseNetRC(``)
	assert.Error(t, err)
	assert.Nil(t, n)

	n, err = parseNetRC(`machine example.org`)
	assert.Error(t, err)
	assert.Nil(t, n)

	n, err = parseNetRC(`
machine test.com
password someTestPWD

login awesomeUser
`)
	assert.NoError(t, err)
	assert.EqualValues(t, &NetRC{Machine: "test.com", Login: "awesomeUser", Password: "someTestPWD"}, n)
}
