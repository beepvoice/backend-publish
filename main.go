package main;

import (
  "context"
  "encoding/json"
  "io/ioutil"
  "net/http"
  "log"
  "os"
  "regexp"
  "strconv"

  . "publish/backend-protobuf/go"

  "github.com/joho/godotenv"
  "github.com/golang/protobuf/proto"
  "github.com/julienschmidt/httprouter"
  "github.com/nats-io/go-nats"
)

const MaxBiteSize = 1024 * 1024 * 10

var listen string
var natsHost string
var secret []byte

var nats_conn *nats.Conn

func main() {
  // Load .env
  err := godotenv.Load()
  if err != nil {
    log.Fatal("Error loading .env file")
  }
  listen = os.Getenv("LISTEN")
  natsHost = os.Getenv("NATS")
  s := os.Getenv("SECRET")

  secret = []byte(s)

  //NATS
  n, err := nats.Connect(natsHost)
  if err != nil {
    log.Fatal(err)
    return
  }
  nats_conn = n

  // Routes
	router := httprouter.New()

  router.PUT("/conversation/:key/start/:start", AuthMiddleware(PutBite)) // bites
  router.PUT("/conversation/:key/start/:start/user", AuthMiddleware(PutBiteUser)) // bite_users

  // Start server
  log.Printf("starting server on %s", listen)
	log.Fatal(http.ListenAndServe(listen, router))
}

type RawClient struct {
  UserId string `json:"userid"`
  ClientId string `json:"clientid"`
}
func AuthMiddleware(next httprouter.Handle) httprouter.Handle {
  return func (w http.ResponseWriter, r *http.Request, p httprouter.Params) {
    ua := r.Header.Get("X-User-Claim")
    if ua == "" {
      http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
  		return
    }

    var client RawClient
    err := json.Unmarshal([]byte(ua), &client)

    if err != nil {
      http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
  		return
    }

    context := context.WithValue(r.Context(), "user", client)
    next(w, r.WithContext(context), p)
  }
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
  client := r.Context().Value("user").(RawClient)
  c := Client {
    Key: client.UserId,
    Client: client.ClientId,
  }

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
    Client: &c,
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
  client := r.Context().Value("user").(RawClient)
  c := Client {
    Key: client.UserId,
    Client: client.ClientId,
  }

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
    Client: &c,
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
