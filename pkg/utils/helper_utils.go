package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/huandu/facebook"
	"github.com/lib/pq"
	"io/ioutil"
	"log"
	"os"
	"scheduler-microservice/pkg/models"
	"time"
)

func HibernateSchedule(connection *sql.DB, schedule models.PostSchedule, namespace string, postsChannel chan <- *models.PostsWithPermission) {
	defer close(postsChannel)
	if schedule.ScheduleId != "" {
		/* Get all schedules that aren't due yet */
		if !schedule.To.Before(time.Now()) {
			// if the schedule is due
			if schedule.From.Before(time.Now()) || schedule.From.Equal(time.Now()) {
				// Do this
				log.Println("Schedule is due")
				// Build the query
				stmt := fmt.Sprintf("SELECT * FROM %s.scheduled_post WHERE scheduled_post_id = $1", namespace)
				log.Println(stmt)
				//query the db
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

				postWithPermission := &models.PostsWithPermission{
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

				postWithPermission := &models.PostsWithPermission{
					Posts:      posts,
					PostToFeed: schedule.PostToFeed,
				}

				postsChannel <- postWithPermission
			}
		}
	}
}

func SchedulePosts(posts <- chan *models.PostsWithPermission, posted <- chan bool, post chan <- models.SinglePostWithPermission, duration float64) {
	// Listen for posts from the other goroutine
	for p := range posts {
		log.Println("Post With Permissions: ", p)

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
		log.Println(p)

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
	stmt := fmt.Sprintf("SELECT user_id, user_access_token FROM %s.application_info", namespace)
	row, err := connection.Query(stmt)
	if err != nil {
		return err
	}
	log.Println(stmt)

	var userData models.FacebookUserData
	var userDataS []models.FacebookUserData
	for row.Next() {
		err = row.Scan(&userData.UserId, &userData.AccessToken)
		if err != nil {
			return err
		}

		userDataS = append(userDataS, userData)
	}

	if userDataS != nil {
		for _, data := range userDataS {
			log.Println(data)

			// Post ti the user's feed if post.PostToFeed is true
			if post.PostToFeed {
				err = Feed(post.Post, data.AccessToken, data.UserId)
			}

			// Post to page
			err = Page(post.Post, data.AccessToken, data.UserId)
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

func Page(post models.Post, token string, id string) error {

	postMessage, err := GeneratePostMessageWithHashTags(post)
	if err != nil {
		return err
	}

	log.Println(postMessage)
	log.Println("Posting to page")

	// Get a list of pages first
	result, err := facebook.Get("/" + id + "/accounts",
		facebook.Params {
			"access_token": token,
		},
	)
	if err != nil {
		return err
	}

	// Decode the data into fbPageData object
	var fbPageData models.FBPData
	err = result.Decode(&fbPageData)
	if err != nil {
		return err
	}

	log.Println(fbPageData)

	for _, d := range fbPageData.Data {
		if post.ImageExtension == "" {
			_, err = facebook.Post("/" + d.Id + "/feed", facebook.Params{
				"message":      postMessage,
				"access_token": d.AccessToken,
			})
			if err != nil {
				return err
			}
			log.Println("image Extension: ", post.ImageExtension)

		} else if post.ImageExtension != "" {

			log.Println("image Extension: ", post.ImageExtension)

			log.Println("Post image")

			blob, err := os.Create("pkg/img/" + post.PostId + "." + post.ImageExtension)
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(blob.Name(), post.PostImage, os.ModeAppend)
			if err != nil {
				return err
			}
			log.Println(blob.Name())

			_, err = facebook.Post("/" + d.Id + "/photos", facebook.Params{
				"message":      postMessage,
				"file": facebook.File(blob.Name()),
				"access_token": d.AccessToken,
			})
			if err != nil {
				return err
			}
			// Delete file
			err = os.Remove(blob.Name())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func Feed(post models.Post, s string, id string) error {

	postMessage, err := GeneratePostMessageWithHashTags(post)
	if err != nil {
		return err
	}

	log.Println("Posting to feed")
	log.Println(postMessage)

	if post.ImageExtension != "" {

		log.Println("Post image")

		blob, err := os.Create("pkg/img/" + post.PostId + "." + post.ImageExtension)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(blob.Name(), post.PostImage, os.ModeAppend)
		if err != nil {
			return err
		}
		log.Println(blob.Name())

		_, err = facebook.Post("/" + id + "/photos", facebook.Params {
			"message":      postMessage,
			"file": facebook.File(blob.Name()),
			"access_token": s,
		} )
		if err != nil {
			return err
		}
	} else {
		log.Println("No Post Image")
		_, err = facebook.Post("/" + id, facebook.Params {
			"message":      postMessage,
			"access_token": s,
		} )
	}

	return nil
}

func GeneratePostMessageWithHashTags(post models.Post) (string, error) {
	m := ""

	if post.HashTags == nil {
		log.Println("Empty hastags")
		return post.PostMessage, nil
	}

	for i := 0; i < len(post.HashTags); i++ {

		if i == 0 {
			m = post.PostMessage + "\n\n" + post.HashTags[i]
		} else {
			m += "\n" + post.HashTags[i]
		}

	}

	return m, nil
}