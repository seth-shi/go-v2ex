package pages

import (
	"go.dalton.dog/bubbleup"
)

const (
	alertInfoColor  = "#3CB179"
	alertErrorColor = "#E05263"
)

func registerDefaultAlertTypes(m *bubbleup.AlertModel) {

	infoDef := bubbleup.AlertDefinition{
		Key:       "Info",
		Prefix:    "✓",
		ForeColor: alertInfoColor,
	}
	m.RegisterNewAlertType(infoDef)
	errorDef := bubbleup.AlertDefinition{
		Key:       "Error",
		Prefix:    "!",
		ForeColor: alertErrorColor,
	}
	m.RegisterNewAlertType(errorDef)
}
