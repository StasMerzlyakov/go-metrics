package domain

import "context"

type ChangeListener func(ctx context.Context, m *Metrics) error
