package messages

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bocha-io/garnet/internal/backend/api"
	"github.com/bocha-io/garnet/internal/database"
	"github.com/bocha-io/garnet/internal/logger"
)

func connectMessage(ws *WebSocketContainer, usersDB *database.InMemoryDatabase, p *[]byte) error {
	var connectMsg ConnectMessage
	err := json.Unmarshal(*p, &connectMsg)
	if err != nil {
		return err
	}

	user, err := usersDB.Login(connectMsg.User, connectMsg.Password)
	if err != nil {
		// TODO: this is only used for the demo, so it autoregisters each user
		client := &http.Client{
			Timeout: 2 * time.Second,
		}

		body := api.RegistationParams{
			Username: connectMsg.User,
			Password: connectMsg.Password,
		}

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return err
		}
		r, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/signup", 6666), bytes.NewBuffer(bodyBytes))
		if err != nil {
			return err
		}
		r.Header.Add("Content-Type", "application/json")

		res, err := client.Do(r)
		if err != nil {
			return fmt.Errorf("error sending the signup request: %s", err.Error())
		}
		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("incorrect response: %d", res.StatusCode)
		}

		// return fmt.Errorf("incorrect credentials")
	}

	user, err = usersDB.Login(connectMsg.User, connectMsg.Password)
	if err != nil {
		return fmt.Errorf("incorrect credentials")
	}

	ws.User = connectMsg.User
	ws.Authenticated = true
	ws.WalletID = user.Index
	ws.WalletAddress = strings.ToLower(user.Address)

	logger.LogInfo(fmt.Sprintf("[backend] user connected: %s (%s)", ws.User, ws.WalletAddress))
	return nil
}
