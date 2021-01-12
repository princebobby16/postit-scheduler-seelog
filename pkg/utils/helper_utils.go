package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/twinj/uuid"
	"io/ioutil"
	"log"
	"os"
	"scheduler-microservice/pkg/models"
	"time"
)

func HibernateSchedule(connection *sql.DB, schedule models.PostSchedule, namespace string, postsChannel chan <- models.PostsWithPermission) {
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
						&post.ImageExtension,
						pq.Array(&post.HashTags),
						&post.PostPriority,
						&post.PostStatus,
						&post.CreatedOn,
						&post.UpdatedOn,
					)

					posts = append(posts, post)
				}
				log.Println(posts)

				postWithPermission := models.PostsWithPermission{
					Posts:      posts,
					PostToFeed: schedule.PostToFeed,
				}

				postsChannel <- postWithPermission
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
						&post.ImageExtension,
						pq.Array(&post.HashTags),
						&post.PostPriority,
						&post.PostStatus,
						&post.CreatedOn,
						&post.UpdatedOn,
					)

					posts = append(posts, post)
				}
				log.Println(posts)

				postWithPermission := models.PostsWithPermission{
					Posts:      posts,
					PostToFeed: schedule.PostToFeed,
				}

				postsChannel <- postWithPermission
			}
		}
	}
}

func SchedulePosts(posts <- chan models.PostsWithPermission, posted <- chan bool, post chan <- models.SinglePostWithPermission, duration float64) {
	// Listen for posts from the other goroutine
	for p := range posts {

		if p.Posts != nil {
			for _, job := range p.Posts {
				singlePostWithPerm := models.SinglePostWithPermission{
					Post:       job,
					PostToFeed: p.PostToFeed,
				}
				post <- singlePostWithPerm
				time.Sleep(time.Duration(duration) * time.Second)
				status := <- posted
				log.Println(status)
				if status == false {
					p.Posts = append(p.Posts, job)
				}
			}
			close(post)
		}
	}

}

func SendPostToFaceBook(post <- chan models.SinglePostWithPermission, posted chan <- bool, namespace string, connection *sql.DB) {
	for p := range post {
		//log.Println("Posted", p, " From ", namespace)
		err := PostToFacebook(p, namespace, connection)
		if err != nil {
			log.Println(err)
			posted <- false
		} else {
			posted <- true
		}
	}
}

func PostToFacebook(post models.SinglePostWithPermission, namespace string, connection *sql.DB) error {

	// use tenantNamespace to get access token
	stmt := fmt.Sprintf("SELECT user_access_token FROM %s.application_info", namespace)
	row, err := connection.Query(stmt)
	if err != nil {
		return err
	}
	log.Println(stmt)

	var accessToken string
	var accessTokens []string
	for row.Next() {
		err = row.Scan(&accessToken)
		if err != nil {
			return err
		}

		accessTokens = append(accessTokens, accessToken)
	}

	if accessTokens != nil {
		for _, token := range accessTokens {
			log.Println(token)

			if post.PostToFeed {
				err = Feed(post.Post, token)
			}
			err = Page(post.Post, token)
			if err != nil {
				return err
			}

			log.Println("Posted")
		}
	} else {
		return errors.New("no facebook access tokens available")
	}

	return nil
}

func Page(post models.Post, s string) error {
	postMessage, err := GeneratePostMessageWithHashTags(post)
	if err != nil {
		return err
	}

	if post.ImageExtension == "" {
		//_, err = fb.Post("/me", fb.Params{
		//	"message":      postMessage,
		//	"access_token": s,
		//})
		//if err != nil {
		//	return err
		//}
		log.Println("image Extension: ", post.ImageExtension)

	} else if post.ImageExtension != "" {

		log.Println("image Extension: ", post.ImageExtension)

		blob, err := os.Create(uuid.NewV4().String() + "_fb." + post.ImageExtension)
		if err != nil {
			return err
		}
		log.Println(blob.Name())

		err = ioutil.WriteFile(blob.Name(), post.PostImage, os.ModeAppend)
		if err != nil {
			return err
		}

		//_, err = fb.Post("/me/feed", fb.Params{
		//	"message":      postMessage,
		//	//"url": "",
		//	"access_token": s,
		//})
		//if err != nil {
		//	return err
		//}
	}

	return nil
}

func Feed(post models.Post, s string) error {

	postMessage, err := GeneratePostMessageWithHashTags(post)
	if err != nil {
		return err
	}

	log.Println("Posting to feed as well")
	log.Println(postMessage)

	//_, err = fb.Post("/me/feed", fb.Params{
	//	"message":      postMessage,
	//	"access_token": s,
	//})
	//if err != nil {
	//	return err
	//}

	return nil
}

func GeneratePostMessageWithHashTags(post models.Post) (string, error) {
	m := ""
	for i := 0; i < len(post.HashTags); i++ {
		if i == 0 {
			m = post.PostMessage + "\n\n" + post.HashTags[i]
		} else {
			m += "\n" + post.HashTags[i]
		}
	}
	return m, nil
}
