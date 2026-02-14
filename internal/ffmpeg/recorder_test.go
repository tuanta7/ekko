package ffmpeg

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FFmpegRecorderTestSuite struct {
	suite.Suite
}

func (suite *FFmpegRecorderTestSuite) TestListSources() {
	recorder, _ := NewRecorder()
	ctx := context.Background()

	result, err := recorder.ListSources(ctx)
	assert.Nil(suite.T(), err)
	assert.NotEmpty(suite.T(), result)
}

func TestFFmpegRecorderTestSuite(t *testing.T) {
	suite.Run(t, new(FFmpegRecorderTestSuite))
}
