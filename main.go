package main;

import (
  "flag"
  "io/ioutil"
  "net/http"
  "log"
  "regexp"
  "strconv"

  "github.com/golang/protobuf/proto"
  "github.com/julienschmidt/httprouter"
  "github.com/nats-io/go-nats"
)

const MaxBiteSize = 1024 * 1024 * 10

var listen string
var natsHost string

var nats_conn *nats.Conn

func main() {
  // Parse flags
	flag.StringVar(&listen, "listen", ":8080", "host and port to listen on")
  flag.StringVar(&natsHost, "nats", "nats://localhost:4222", "host and port of NATS")
  flag.Parse()

  //NATS
  n, err := nats.Connect(natsHost)
  if err != nil {
    log.Fatal(err)
    return
  }
  nats_conn = n

  // Routes
	router := httprouter.New()

  router.PUT("/conversation/:key/start/:start", PutBite) // bites
  router.PUT("/conversation/:key/start/:start/user", PutBiteUser) // bite_users

  // Start server
  log.Printf("starting server on %s", listen)
	log.Fatal(http.ListenAndServe(listen, router))
}

// TODO: ensure security of regexp
var validConversationRegexp = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

func validConversation(conversation string) bool {
	return validConversationRegexp.MatchString(conversation)
}

func ParseStartString(start string) (uint64, error) {
	return strconv.ParseUint(start, 10, 64)
}

// Route handlers
func PutBite(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
  start, err := ParseStartString(p.ByName("start"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	key := p.ByName("key")
	if !validConversation(key) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	reader := http.MaxBytesReader(w, r.Body, MaxBiteSize)
	body, err := ioutil.ReadAll(reader)
  if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

  b := Bite {
    Start: start,
    Key: key,
    Data: body,
  }
  out, err := proto.Marshal(&b)
  if err != nil {
    http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
    log.Print(err)
    return
	}
  nats_conn.Publish("new_bite", out)

  w.WriteHeader(200)
}

func PutBiteUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
  start, err := ParseStartString(p.ByName("start"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

  key := p.ByName("key")
	if !validConversation(key) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	reader := http.MaxBytesReader(w, r.Body, MaxBiteSize)
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

  b := Bite {
    Start: start,
    Key: key,
    Data: body,
  }
  out, err := proto.Marshal(&b)
  if err != nil {
    http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
    log.Print(err)
    return
	}
  nats_conn.Publish("new_bite_user", out)

  w.WriteHeader(200)
}
