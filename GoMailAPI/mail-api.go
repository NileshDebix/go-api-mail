package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"

	"net/http"
	"net/smtp"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// Create a struct that represents the JSON Payload that will be sent to the end user
type TargetForEmail struct {
	TargetUserName      string   `json:"user_name"`
	TargetEmail         []string `json:"target_emailadress"`
	TargetEmailSubjects string   `json:"target_emailsub"`
	TargetEmailBody     string   `json:"target_emailbody"`
}

// Log email information to an log file in the tmp directory
func LogEmailInfo(TargetUserName string, TargetEmail []string) {

	// Log TargetUserName and TargetEmail
	file, err := os.OpenFile("./tmp/log/mail.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer file.Close()
	wrt := io.MultiWriter(os.Stdout, file)
	log.SetOutput(wrt)
	log.Printf("Mail send to: %v target email: %v", TargetUserName, TargetEmail)

}

func SendMail(TargetUserName string, TargetEmail []string, TargetEmailSubjects string, TargetEmailBody string) {

	// Creating a viper instance
	viper_ := viper.New()
	viper_.SetConfigFile("config.yml")
	viper_.ReadInConfig()

	// Let Viper read the configuration
	sender := viper_.GetString("senderEmail")
	host := viper_.GetString("mailServer")
	port := viper_.GetString("tcpPort")
	username := viper_.GetString("userName")
	password := viper_.GetString("password")

	// Insert viper yaml data
	body := []byte(TargetEmailBody)
	target_email := []string(TargetEmail)

	// Create email header
	header := make(map[string]string)
	header["From"] = sender
	header["Subject"] = TargetEmailSubjects
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	// Create the email-body together with the header and convert it into byte-code
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	// Authorize yourself
	auth := smtp.PlainAuth("", username, password, host)
	sendmail := smtp.SendMail(host+":"+port, auth, sender, target_email, []byte(message))

	// If email-error occurs the console will print the error
	if sendmail != nil {

		fmt.Println(sendmail)
		os.Exit(1)

	} else {

		// Message was successfully sent
		fmt.Println("Successfully sent mail")
		LogEmailInfo(TargetUserName, TargetEmail)

	}

}

func RouterPost(ctx *gin.Context) {

	// The incoming JSON data
	IncomingRequest := new(TargetForEmail)

	if err := ctx.BindJSON(&IncomingRequest); err != nil {

		// Error Message for user "HTTP = 200"
		BadRequest := make(map[string]string)

		// Custom error message for the end client
		BadRequest["HTTP_CODE"] = "400"
		BadRequest["HTTP_TEXT"] = "Something went wrong checking your JSON BODY"

		// Response Message for user "HTTP = 400"
		ctx.IndentedJSON(http.StatusBadRequest, BadRequest)

	} else {

		// Send SMS to client
		// This function need the required JSON data, that you passed via the REST endpoint. In my case http://127.0.0.1:5000/api/v1/post/send/sms/client
		go SendMail(IncomingRequest.TargetUserName, IncomingRequest.TargetEmail, IncomingRequest.TargetEmailSubjects, IncomingRequest.TargetEmailBody)

		// Response Message for user "HTTP = 200"
		ctx.IndentedJSON(http.StatusCreated, IncomingRequest)

	}
}

func main() {

	// Create the Go Gin engine instance
	router := gin.Default()

	// Create an API group
	API_v1 := router.Group("/api")

	// Provide and REST endpoint
	API_v1.POST("/v1/post/send/mail/client", RouterPost)

	// Run the Router
	router.SetTrustedProxies([]string{"127.0.0.1"})
	router.Run("127.0.0.1:8080")

}
