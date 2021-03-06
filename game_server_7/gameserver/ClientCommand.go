package gameserver

import (
	"encoding/json"
)

const (
	CLIENT_COMMAND_TYPE_MOVE uint8 = 0
	CLIENT_COMMAND_TYPE_HIT  uint8 = 1
)

type ClientCommandHitInfo struct {
	ID     uint32 `json:"id"`
	Damage int16  `json:"damage"`
}

type ClientCommand struct {
	ID             uint32                 `json:"id"`
	CommandType    uint8                  `json:"type"`
	RotationX      float64                `json:"rx"`
	RotationY      float64                `json:"ry"`
	RotationZ      float64                `json:"rz"`
	X              float64                `json:"x"`
	Y              float64                `json:"y"`
	VX             float64                `json:"vx"`
	VY             float64                `json:"vy"`
	Duration       float64                `json:"duration"`
	VisualState    uint8                  `json:"visualState"`
	AnimName       string                 `json:"animName"`
	StartSkillName string                 `json:"startSkillName"`
	HitMonsters    []ClientCommandHitInfo `json:"hitMonsters"`
}

func NewClientCommand(data []byte) (*ClientCommand, error) {
	command := &ClientCommand{}
	err := json.Unmarshal(data, command)
	return command, err
}
