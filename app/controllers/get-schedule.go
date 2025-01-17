package controllers

import (
	"encoding/json"
	"github.com/twinj/uuid"
	"io/ioutil"
	"net/http"
	"scheduler-microservice/db"
	"scheduler-microservice/pkg/logs"
	"scheduler-microservice/pkg/models"
	"scheduler-microservice/pkg/utils"
	"time"
)

func GetSchedule(w http.ResponseWriter, r *http.Request) {
	transactionId := uuid.NewV4().String()
	logs.Logger.Info("Transaction Id:", transactionId)

	tenantNamespace := r.Header.Get("tenant-namespace")
	logs.Logger.Info("Tenant Namespace:", tenantNamespace)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		_ = logs.Logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	logs.Logger.Info(string(body))

	var schedule models.PostSchedule
	err = json.Unmarshal(body, &schedule)
	if err != nil {
		_ = logs.Logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	logs.Logger.Info("Schedule: ", schedule)

	post := make(chan models.SinglePostWithPermission, 1)
	posts := make(chan *models.PostsWithPermission)
	posted := make(chan bool)

	go utils.HibernateSchedule(db.Connection, schedule, tenantNamespace, posts)
	go utils.SchedulePosts(posts, posted, post, schedule.Duration)
	go utils.SendPostToFaceBook(post, posted, tenantNamespace, db.Connection)

	var response = models.StandardResponse {
		Data: models.Data{
			Id:        transactionId,
			UiMessage: "Schedule received and being worked on",
		},
		Meta: models.Meta{
			Timestamp:     time.Now(),
			TransactionId: transactionId,
			TraceId:       "",
			Status:        "SUCCESS",
		},
	}

	_ = json.NewEncoder(w).Encode(&response)
}