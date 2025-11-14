package model

import "github.com/liuminhaw/yatijapp-tui/internal/data"

type record interface {
	// GetActualType() data.RecordType
	GetTitle() string
	GetParentsTitle() map[data.RecordType]string
}

type input interface {
	Value() string
}
