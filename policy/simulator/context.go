package simulator

import "github.com/palantir/policy-bot/pull"

type SimulatedContext struct {
	pull.Context
	options *Options
}

func NewSimulatedContext(pullContext pull.Context, options *Options) *SimulatedContext {
	return &SimulatedContext{Context: pullContext, options: options}
}

func (c *SimulatedContext) Comments() ([]*pull.Comment, error) {
	comments, err := c.Context.Comments()
	if err != nil {
		return nil, err
	}

	comments = c.options.filterIgnoredComments(comments)
	return comments, nil
}
