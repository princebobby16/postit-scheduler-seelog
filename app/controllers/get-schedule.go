package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"github.com/twinj/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"scheduler-microservice/db"
	"scheduler-microservice/pkg/logs"
	"scheduler-microservice/pkg/models"
	"time"
)

func GetSchedule(w http.ResponseWriter, r *http.Request) {
	transactionId := uuid.NewV4().String()
	logs.Log("Transaction Id:", transactionId)

	tenantNamespace := r.Header.Get("tenant-namespace")
	logs.Log("Tenant Namespace:", tenantNamespace)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Println(string(body))

	var schedule models.PostSchedule
	err = json.Unmarshal(body, &schedule)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logs.Log("Schedule: ", schedule)

	post := make(chan models.Post, 1)
	posts := make(chan []models.Post)
	posted := make(chan bool)

	go HibernateSchedule(db.Connection, schedule, tenantNamespace, posts)
	go SchedulePosts(posts, posted, post, schedule.Duration)
	go SendPostToFaceBook(post, posted, tenantNamespace)

	var response = models.StandardResponse{
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

	_ = json.NewEncoder(w).Encode(response)
}

func HibernateSchedule(connection *sql.DB, schedule models.PostSchedule, namespace string, postsChannel chan<- []models.Post) {
	defer close(postsChannel)
	if schedule.ScheduleId != "" {
		/* Get all schedules that aren't due yet */
		if !schedule.To.Before(time.Now()) {
			//time.Sleep(time.Minute * time.Duration(schedule.From.Minute()))
			// if the schedule is due
			if schedule.From.Before(time.Now()) || schedule.From.Equal(time.Now()) {
				// Do this
				log.Println("Schedule is due")
				stmt := fmt.Sprintf("SELECT * FROM %s.scheduled_post WHERE scheduled_post_id = $1", namespace)
				log.Println(stmt)
				rows, err := connection.Query(stmt, schedule.ScheduleId)
				if err != nil {
					log.Println(err)
					return
				}

				if rows.Err() != nil {
					log.Println(rows.Err().Error())
					return
				}

				var posts []models.Post
				for rows.Next() {
					var post models.Post
					err = rows.Scan(
						&post.ScheduleId,
						&post.PostId,
						&post.PostMessage,
						&post.PostImage,
						pq.Array(&post.HashTags),
						&post.PostPriority,
						&post.PostStatus,
						&post.CreatedOn,
						&post.UpdatedOn,
					)

					posts = append(posts, post)
				}
				log.Println(posts)

				postsChannel <- posts

			} else {
				log.Printf("About to wait for schedule for %v Seconds", schedule.From.Sub(time.Now()))
				//	wait till its due before sending
				time.Sleep(schedule.From.Sub(time.Now()))
				log.Println("Due now")
				stmt := fmt.Sprintf("SELECT * FROM %s.scheduled_post WHERE scheduled_post_id = $1", namespace)
				log.Println(stmt)
				rows, err := connection.Query(stmt, schedule.ScheduleId)
				if err != nil {
					log.Println(err)
					return
				}

				if rows.Err() != nil {
					log.Println(rows.Err().Error())
					return
				}

				var posts []models.Post
				for rows.Next() {
					var post models.Post
					err = rows.Scan(
						&post.ScheduleId,
						&post.PostId,
						&post.PostMessage,
						&post.PostImage,
						pq.Array(&post.HashTags),
						&post.PostPriority,
						&post.PostStatus,
						&post.CreatedOn,
						&post.UpdatedOn,
					)

					posts = append(posts, post)
				}
				log.Println(posts)

				postsChannel <- posts
			}
		}
	}
}

func SchedulePosts(posts chan []models.Post, posted chan bool, post chan <- models.Post, duration float64) {
	// Listen for schedules on the channel and store it in a variable (currentSchedule)
	//var currentSchedule PostSchedule
	var postArray []models.Post

	for p := range posts {
		//currentSchedule = s
		log.Println("Current Post Array: ", p)

		for _, job := range p {
			post <- job
			time.Sleep(time.Duration(duration) * time.Second)
			status := <- posted
			if !status {
				postArray = append(postArray, job)
			}
		}
		close(post)
	}
}

func SendPostToFaceBook(post chan models.Post, posted chan bool, namespace string) {
	for p := range post {
		log.Println("Posted", p, " From ", namespace)
		posted <- true
	}
	close(posted)
}
