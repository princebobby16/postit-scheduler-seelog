package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/huandu/facebook"
	"github.com/lib/pq"
	"io/ioutil"
	"os"
	"path/filepath"
	"scheduler-microservice/pkg/logs"
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
				logs.Logger.Info("Schedule is due")
				// Build the query
				stmt := fmt.Sprintf("SELECT * FROM %s.scheduled_post WHERE scheduled_post_id = $1", namespace)
				logs.Logger.Info(stmt)
				//query the db
				rows, err := connection.Query(stmt, schedule.ScheduleId)
				if err != nil {
					err = logs.Logger.Critical(err)
					if err != nil {
						_ = logs.Logger.Error(err)
					}
					return
				}

				if rows.Err() != nil {
					logs.Logger.Error(rows.Err().Error())
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
				logs.Logger.Info("About to wait for schedule for %v Seconds", schedule.From.Sub(time.Now()))
				//	wait till its due before sending
				time.Sleep(schedule.From.Sub(time.Now()))
				logs.Logger.Info("Due now")
				stmt := fmt.Sprintf("SELECT * FROM %s.scheduled_post WHERE scheduled_post_id = $1", namespace)
				logs.Logger.Info(stmt)
				rows, err := connection.Query(stmt, schedule.ScheduleId)
				if err != nil {
					logs.Logger.Error(err)
					return
				}

				if rows.Err() != nil {
					logs.Logger.Error(rows.Err().Error())
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

		if p.Posts != nil {
			for i := 0; i < len(p.Posts); i++ {
				singlePostWithPerm := models.SinglePostWithPermission{
					Post:       p.Posts[i],
					PostToFeed: p.PostToFeed,
				}

				post <- singlePostWithPerm
				status := <- posted
				if !status {
					logs.Logger.Warn("Unable to post... Queueing...")
					p.Posts = append(p.Posts, p.Posts[i])
				}
				time.Sleep(time.Duration(duration) * time.Second)
			}
			close(post)
		}
	}

}

func SendPostToFaceBook(post <- chan models.SinglePostWithPermission, posted chan <- bool, namespace string, connection *sql.DB) {
	for p := range post {
		logs.Logger.Info(p.Post.PostMessage, "Image Extension", p.Post.ImageExtension)

		err := PostToFacebook(p, namespace, connection)
		if err != nil {
			_ = logs.Logger.Critical(err)
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
	logs.Logger.Info(stmt)

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
			logs.Logger.Info(data)

			//Post to page
			logs.Logger.Info("Posting to Page")
			err = Page(post.Post, data.AccessToken, data.UserId)
			if err != nil {
				_ = logs.Logger.Critical(err)
				return err
			}

			logs.Logger.Info("Posted")
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
	logs.Logger.Info(postMessage)

	logs.Logger.Info("Retrieving page info from facebook")
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

	logs.Logger.Info(fbPageData)

	if fbPageData.Data != nil {
		for _, d := range fbPageData.Data {
			if post.ImageExtension == "" {
				logs.Logger.Info("Posting Without Image")
				_res, err := facebook.Post("/" + d.Id + "/feed", facebook.Params{
					"message":      postMessage,
					"access_token": d.AccessToken,
				})
				if err != nil {
					logs.Logger.Error(err)
					return err
				}
				logs.Logger.Info("Posted: ", _res.Get("id"))

			} else if post.ImageExtension != "" {
				logs.Logger.Info("Posting With Image")
				logs.Logger.Info("image Extension: ", post.ImageExtension)

				logs.Logger.Info("Creating image file")
				blob, err := os.Create(post.PostId + "." + post.ImageExtension)
				if err != nil {
					return err
				}

				logs.Logger.Info("Writing image content to file")
				err = ioutil.WriteFile(blob.Name(), post.PostImage, os.ModeAppend)
				if err != nil {
					return err
				}
				logs.Logger.Info(blob.Name())

				logs.Logger.Info("Getting image full path")
				wd, err := os.Getwd()
				if err != nil {
					return err
				}

				completeImagePath := filepath.Join(wd, blob.Name())
				logs.Logger.Info(completeImagePath)

				_res, err := facebook.Post("/" + d.Id + "/photos", facebook.Params{
					"message":      postMessage,
					"file": facebook.File(completeImagePath),
					"access_token": d.AccessToken,
				})
				if err != nil {
					logs.Logger.Error(err)
					return err
				}
				logs.Logger.Info("Posted: ", _res.Get("id"))
				// Delete file
				logs.Logger.Info("Deleting Image File From Directory")
				err = os.Remove(completeImagePath)
				if err != nil {
					return err
				}
			}
		}
	} else {
		logs.Logger.Warn("No Facebook Pages Found")
		return errors.New("no facebook pages found")
	}

	return nil
}

func Feed(post models.Post, s string, id string) error {

	postMessage, err := GeneratePostMessageWithHashTags(post)
	if err != nil {
		return err
	}
	logs.Logger.Info(postMessage)

	if post.ImageExtension != "" {
		logs.Logger.Info("Posting With Image")

		logs.Logger.Info("Creating image file")
		blob, err := os.Create(post.PostId + "." + post.ImageExtension)
		if err != nil {
			return err
		}

		logs.Logger.Info("Writing image content to file")
		err = ioutil.WriteFile(blob.Name(), post.PostImage, os.ModeAppend)
		if err != nil {
			return err
		}
		logs.Logger.Info(blob.Name())

		logs.Logger.Info("Getting image full path")
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		completeImagePath := filepath.Join(wd, blob.Name())
		logs.Logger.Info(completeImagePath)

		_res, err := facebook.Post("/" + id +"/photos", facebook.Params {
			"message":      postMessage,
			"file": facebook.File(completeImagePath),
			"access_token": s,
		} )
		if err != nil {
			logs.Logger.Error(err)
			return err
		}
		err = os.Remove(completeImagePath)
		if err != nil {
			return err
		}
		logs.Logger.Info("Posted: ", _res.Get("id"))
	} else {
		logs.Logger.Info("Posting Without Image")
		_res, err := facebook.Post("/" + id + "/feed", facebook.Params {
			"message": postMessage,
			"access_token": s,
		})
		if err != nil {
			return err
		}
		logs.Logger.Info("Posted: ", _res.Get("id"))
	}

	return nil
}

func GeneratePostMessageWithHashTags(post models.Post) (string, error) {
	m := ""

	if post.HashTags == nil {
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