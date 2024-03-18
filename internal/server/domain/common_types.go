package domain

type ChangeListener interface {
	Refresh(metrics *Metrics) error
}
