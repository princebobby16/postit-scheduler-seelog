package controllers

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"net/http"
	"scheduler-microservice/pkg/logs"
)

func ScheduleStatus(w http.ResponseWriter, r *http.Request) {
	 conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		logs.Logger.Info(err)
		return
	}
	go func() {
		defer conn.Close()

		for {
			msg, op, err := wsutil.ReadClientData(conn)
			if err != nil {
				_ = logs.Logger.Error(err)
				return
			}
			logs.Logger.Info(string(msg) + " : " + string(op))

			err = wsutil.WriteClientMessage(conn, op, msg)
			if err != nil {
				_ = logs.Logger.Error(err)
				return
			}
		}
	}()
}

//func Writer(conn net.Conn, post chan models.Post) {
//	for {
//		ticker := time.NewTicker(1 * time.Second)
//
//		for t := range ticker.C {
//
//		}
//	}
//}
