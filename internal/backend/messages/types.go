package messages

import "github.com/bocha-io/garnet/internal/indexer/data"

type BasicMessage struct {
	MsgType string `json:"msgtype"`
}

type ConnectMessage struct {
	MsgType  string `json:"msgtype"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type JoinMatch struct {
	MsgType string `json:"msgtype"`
	MatchID string `json:"id"`
}

// TODO: remove this
type EndTurn struct {
	MsgType string `json:"msgtype"`
	MatchID string `json:"id"`
}

// TODO: remove this
type PlaceCard struct {
	MsgType string `json:"msgtype"`
	CardID  string `json:"id"`
	X       int64  `json:"x"`
	Y       int64  `json:"y"`
}

// TODO: remove this
type MoveCard struct {
	MsgType string `json:"msgtype"`
	CardID  string `json:"id"`
	X       int64  `json:"x"`
	Y       int64  `json:"y"`
}

// TODO: remove this
type Attack struct {
	MsgType string `json:"msgtype"`
	CardID  string `json:"id"`
	X       int64  `json:"x"`
	Y       int64  `json:"y"`
}

type Match struct {
	Id        string `json:"id"`
	Creator   string `json:"creator"`
	PlayerTwo string `json:"playertwo"`
}

type MatchList struct {
	MsgType string  `json:"msgtype"`
	Matches []Match `json:"matches"`
}

type GetMatchStatus struct {
	MsgType string `json:"msgtype"`
	MatchID string `json:"id"`
}

type BoardStatus struct {
	MsgType string         `json:"msgtype"`
	Status  data.MatchData `json:"status"`
}

// type RespMatchStatus struct {
// 	MsgType   string    `json:"msgtype"`
// 	MatchData MatchData `json:"value"`
// }
