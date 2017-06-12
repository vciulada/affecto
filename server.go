package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"github.com/streadway/amqp"
	"strconv"
	"bytes"
	"fmt"
	"os"
)

 type Policy struct {
	 Email string `json:"email"`
         Premium int `json:"premium"`
 }

 func sendHandler(w http.ResponseWriter, r *http.Request) {
	 log.Println(r.FormValue("email"))
	 if  r.FormValue("password") == "op" && len(r.FormValue("email"))>0{
		 t, _ := template.ParseFiles("sender.html") //setp 1
		 t.Execute(w,r.FormValue("email")) //step 2
	 }else{
		 mainHandler(w,r)
	 }
 }

func mainHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("index.html") //setp 1
	t.Execute(w,"") //step 2
}

func policyHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("policy.html") //setp 1
	t.Execute(w,"") //step 2
}



 func postMembersHandler(ch *amqp.Channel, q amqp.Queue) func(w http.ResponseWriter, r *http.Request) {
	 return func(w http.ResponseWriter, r *http.Request){
		 var p Policy
		 b, _ := ioutil.ReadAll(r.Body)
		 log.Println(b)
		 json.Unmarshal(b, &p)
		 log.Println(p)
		 send(ch, q, p.Email, p.Premium)
		 content, err := json.Marshal(struct{
			 Premium int `json:"premium"`
		 }{Premium:p.Premium})
		 if err != nil {
			 w.WriteHeader(http.StatusInternalServerError)
		 }
		 w.Header().Set("Content-Type","application/json")
		 w.WriteHeader(200)
		 w.Write(content)
	 }
 }

 func main() {
	 conn, err := amqp.Dial("amqp://tia:tia@195.230.98.205:5672/tia")
	 //conn, err := amqp.Dial("amqp://tia:tia@localhost:5672/tia")
	 failOnError(err, "Failed to connect to RabbitMQ")
	 defer conn.Close()

	 ch, err := conn.Channel()
	 failOnError(err, "Failed to open a channel")
	 defer ch.Close()

	 q, err := ch.QueueDeclare(
		 "hello", // name
		 false, // durable
		 false, // delete when unused
		 false, // exclusive
		 false, // no-wait
		 nil, // arguments
	 )
	 failOnError(err, "Failed to declare a queue")



	 r := mux.NewRouter()
	 r.HandleFunc("/send/", sendHandler).Methods("post")
	 r.HandleFunc("/", mainHandler).Methods("GET")
	 r.HandleFunc("/policy/", policyHandler).Methods("GET")
	 r.HandleFunc("/policy/", postMembersHandler(ch,q)).Methods("POST")
	 http.Handle("/", r)

	 fmt.Println("Listening on port "+os.Getenv("HTTP_PLATFORM_PORT")+"....")
	 http.ListenAndServe(":"+os.Getenv("HTTP_PLATFORM_PORT"), nil)
 }

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

/*func message(email string)[]byte{
	rand.Seed(int64(time.Now().Nanosecond()));
	amount := rand.Intn(500)
	var buffer bytes.Buffer

	buffer.WriteString(`{"prodId":"ZC","name":"Vardas Pavardė","email":"`)
	buffer.WriteString(email)
	buffer.WriteString(`","premiumDiff":`)
	buffer.WriteString(strconv.Itoa(amount))
	buffer.WriteString(`}`)
	return buffer.Bytes()
}*/

func message(email string, premium int)[]byte{
	var buffer bytes.Buffer
	buffer.WriteString(`{"prodId":"ZC","name":"Vardas Pavardė","email":"`)
	buffer.WriteString(email)
	buffer.WriteString(`","premiumDiff":`)
	buffer.WriteString(strconv.Itoa(premium))
	buffer.WriteString(`}`)
	return buffer.Bytes()
}

func send(ch *amqp.Channel, q amqp.Queue, email string, premium int) {


		//

		body := message(email, premium)
		err := ch.Publish(
			"", // exchange
			q.Name, // routing key
			false, // mandatory
			false, // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:   body,
			})
		log.Printf(" [x] Sent %s", body)
		failOnError(err, "Failed to publish a message")
}
