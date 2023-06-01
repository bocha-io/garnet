package actions

const WorldID = "0x5FbDB2315678afecb367f032d93F642f64180aa3"

type BasicAction struct {
	MsgType string `json:"msgtype"`
	CardID  string `json:"id"`
	X       int64  `json:"x"`
	Y       int64  `json:"y"`
}

type PlaceCard struct {
	BasicAction
}

type Attack struct {
	BasicAction
}

type MoveCard struct {
	BasicAction
}

type EndTurn struct {
	MsgType string `json:"msgtype"`
	MatchID string `json:"id"`
}
