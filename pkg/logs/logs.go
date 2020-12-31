package logs

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func HandlerLog(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		inner.ServeHTTP(w, r)

		elapsedTime := time.Since(start)

		message := fmt.Sprintf(
			"%s ==> Log Message: %s\t%s\t%s\t%s\t%s\t%s",
			time.Now(),
			r.Method,
			r.RequestURI,
			name,
			elapsedTime,
			r.Header.Get("Content-Type"),
			r.RemoteAddr,
		)
		Log(message)
	})
}

func logToFile(message, filename string) error {
	message += "\n"
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 666)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = file.WriteString(message)
		if err != nil {
			return err
		}

		return nil
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(message)
	if err != nil {
		return err
	}

	return nil
}

func Log(messages ...interface{}) {
		// for each message; store it in a defined format
		message := fmt.Sprintf("%s ==> Log Message: %v\n", time.Now(), messages)
		// write the message to a file
		err := logToFile(message, "log.txt")
		if err != nil {
			log.Println(err)
		}
		log.Println(message)
}
