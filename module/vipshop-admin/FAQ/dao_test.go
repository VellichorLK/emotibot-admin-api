package FAQ

import (
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestEscape(t *testing.T) {
	msgTemplate := "should escape image tag, but get: %s"
	source1 := "<img src=\"some_img_url.jpg\">"
	target1 := "[图片]"
	result1 := Escape(source1)

	msg1 := fmt.Sprintf(msgTemplate, result1)
	assert.Equal(t, result1, target1, msg1)

	source2 := "this string does not need escaping"
	target2 := source2
	result2 := Escape(source2)
	assert.Equal(t, result2, target2, "should not escape")
}

func fakeTagMapFactory() map[string]Tag {
	tagMap := make(map[string]Tag)

	tag1 := Tag {
		Type: 1,
		Content: "tag1",
	}

	tag2 := Tag {
		Type: 2,
		Content: "tag2",
	}

	tag3 := Tag {
		Type: 4,
		Content: "tag3",
	}

	tagMap["1"] = tag1
	tagMap["2"] = tag2
	tagMap["3"] = tag3
	return tagMap
}

func TestFormDimension(t *testing.T) {
	source := []string {"1", "2", "3"}
	target := []string {"tag1", "tag2", "", "tag3", ""}
	tagMap := fakeTagMapFactory()
	result := FormDimension(source, tagMap)

	for i := 0; i < 5; i++ {
		assert.Equal(t, target[0], result[0], "should be same")
	}
}