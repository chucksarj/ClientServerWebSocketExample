package main

// dependencies 
import (
    "fmt"
    "os"
    "log"
    "net/http"
    "sync"
    "github.com/gorilla/mux"
    "github.com/gorilla/websocket"
    "github.com/golang/protobuf/proto"
)

type CMsg struct {
    Client *websocket.Conn
    Data []byte
}

// we need to take a mutex log when updating our file/db
var mu = &sync.Mutex{}
var i int = 1 // record #
var logfile *os.File

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan CMsg)
// upgrade our normal HTTP connection to websocket
var upgrader = websocket.Upgrader {
    ReadBufferSize:  1024, // set read buffer limits
    WriteBufferSize: 1024, // set write buffer limits 
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func cleanup(client *websocket.Conn) {
    client.Close()
    delete(clients, client)
}

func handledata() {
	for {
		// Grab the next messages acynchronously from the broadcast channel
		msg := <-broadcast

        // store for easy access
        c := msg.Client // client WebSocket connection 
        d := msg.Data // data
        Nproto := &SensorData{}
        // success flag
        flag := int32(1)
        // internal use
        debug := true
        tmp := ""

        if err := proto.Unmarshal(d, Nproto); err != nil {
            fmt.Println(err)
        }

        // deal with async writes by taking a lock when updating file
        // or db or whatever we want to update -> this is our critical region
        // UPDATE: not sure if necessary but just to prevent data race for global record
        // number i, putting it in critical region
        mu.Lock()
        fmt.Fprintf(logfile, "Record Number #%d: %v\n", i, Nproto)
        // update our record number
        i += 1
        mu.Unlock()

        if (debug) {
            fmt.Println("Recieved:", Nproto)
        }

        // part two: send back ack to client 
        Cproto :=  &SensorData {
            ApName: &tmp,  // intialize for syntax reasons  
            Status: &flag, // success
        }

        // syntax reasons we have to set like this
        *Cproto.ApName = Nproto.GetApName()

        // marshal our structure into protobuf codec
        data, err := proto.Marshal(Cproto)
        if err != nil {
            log.Fatal("marshaling error: ", err)
            cleanup(c)
            continue
        }

        // send to client
        err = c.WriteMessage(websocket.BinaryMessage, data)
        if err != nil {
            log.Fatal("WriteMessage failed ", err)
            cleanup(c)
            continue
        }
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
    // upgrade to websocket
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Fatal(err)
    }

    // Make sure we close the connection when the function returns
	defer ws.Close()

    // register client
    clients[ws] = true

    log.Println("Connection Opened")
    for {
        var msg CMsg

        _, tmpData, err := ws.ReadMessage()
        if err != nil {
            log.Printf("Connection Closed: %v", err)

            cleanup(ws)
            break
        }

        // populate our structure before tossing it onto the channel
        msg.Client = ws
        msg.Data = tmpData

        // if no errors, throw our protobuf data onto the channel
        broadcast <- msg
    }
}


func main() {
    port := ":8081"
    f, err := os.Create("log.txt")
    if err != nil {
        panic(err)
    }
    defer f.Close()

    logfile = f

    // gorilla mux router for endpoints -> 
    // in our case we only have one endpoint but easy to scale
    router := mux.NewRouter()

    // endpoints
    router.HandleFunc("/ws", handleConnections)

    // asynchronous execution 
    go handledata()


    fmt.Printf("Starting server at %s\n", port[1:])
    // bind and listen on 8080 by default
    // Server using our gorilla mux router
    if err := http.ListenAndServe(port, router); err != nil {
        fmt.Printf("%s\n", err)
    }
}

