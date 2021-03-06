package gameserver

const (
	PLAYER_MESSAGE_TYPE_PLAYER_INIT = 0
	PLAYER_MESSAGE_TYPE_WORLD_STATE = 1
)

type ToPlayerMessage struct {
	Type             uint8         `json:"messageType"`
	RoomState        GameRoomState `json:"room"`
	LeftClientState  ClientState   `json:"leftPlayer"`
	RightClientState ClientState   `json:"rightPlayer"`
}
