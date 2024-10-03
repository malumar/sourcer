package sourcer

import (
	"github.com/stretchr/testify/assert"
	"malumar/sourcer/annotations"
	"testing"
)

func Test1(t *testing.T) {
	assert := assert.New(t)
	assert.True(true)

	annotations.RegisterAnnotationLine("Entity", nil, nil)

	config, err := ParseGoFilesInDir("tesci", "test/")
	assert.NoError(err)
	assert.NotNil(config)
	config.ParsedSources()
	for i, _ := range config.Operations() {
		t.Logf("Config %d %v komentarze: %v", i, config.ParsedSources().Operations[i], config.ParsedSources().Operations[i].CommentLines)
		a, ok := config.Registry().ResolveAnnotationByName(config.ParsedSources().Operations[i].DocLines, "Entity")
		t.Logf("Annotacje %v %v", a, ok)
	}
}
