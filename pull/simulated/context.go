package simulated

import "github.com/palantir/policy-bot/pull"

type Context struct {
	pull.Context
	options *Options
}

func NewContext(pullContext pull.Context, options *Options) *Context {
	return &Context{Context: pullContext, options: options}
}

func (c *Context) Comments() ([]*pull.Comment, error) {
	comments, err := c.Context.Comments()
	if err != nil {
		return nil, err
	}

	comments = c.options.filterIgnoredComments(comments)
	return comments, nil
}
