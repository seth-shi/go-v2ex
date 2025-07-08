package pages

import (
	"go.dalton.dog/bubbleup"
)

const (
	alertInfoColor  = "#2E8B57" // 原#00FF00 → 深绿色(SeaGreen)，降低77%亮度
	alertErrorColor = "#B22222" // 原#FF0000 → 砖红色(FireBrick)，降低57%亮度
)

func registerDefaultAlertTypes(m *bubbleup.AlertModel) {

	infoDef := bubbleup.AlertDefinition{
		Key:       "Info",
		Prefix:    "♪",
		ForeColor: alertInfoColor,
	}
	m.RegisterNewAlertType(infoDef)
	errorDef := bubbleup.AlertDefinition{
		Key:       "Error",
		Prefix:    "♫",
		ForeColor: alertErrorColor,
	}
	m.RegisterNewAlertType(errorDef)
}
