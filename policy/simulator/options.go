package simulator

import (
	"github.com/palantir/policy-bot/pull"
)

type Options struct {
	IgnoreCommentsFrom []string
}

func (o *Options) filterIgnoredComments(comments []*pull.Comment) []*pull.Comment {
	if len(o.IgnoreCommentsFrom) <= 0 {
		return comments
	}

	ignoreCommentFromMap := make(map[string]bool)
	for _, name := range o.IgnoreCommentsFrom {
		ignoreCommentFromMap[name] = true
	}

	var filteredComments []*pull.Comment
	for _, comment := range comments {
		if ignoreCommentFromMap[comment.Author] {
			continue
		}

		filteredComments = append(filteredComments, comment)
	}

	return filteredComments
}
