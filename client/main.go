
package main

import (
"fmt"
"time"
"bytes"
"io"
"strings"
"net/http"
"io/ioutil"
"bufio"
"log"
"net"
"github.com/gorilla/mux"
"github.com/gorilla/websocket"
"github.com/skratchdot/open-golang/open")

var nodeServer = "";
var upgrader = websocket.Upgrader{}

func main(){
    router := mux.NewRouter()
    router.HandleFunc("/",RootPage).Methods("GET")
    router.HandleFunc("/",FileUpload).Methods("POST")
    router.HandleFunc("/socket",StatusSock).Methods("GET")

    log.Println("Serving on :5000")
    go open.Run("http://localhost:5000")
    log.Fatal(http.ListenAndServe(":5000",router))
}
func RootPage(w http.ResponseWriter, r *http.Request){
	data, err := ioutil.ReadFile("html/index.html")
    if err != nil {
        http.Error(w, "Couldn't read file", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Write(data)	
}

func StatusSock(w http.ResponseWriter, r *http.Request){
    c, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("upgrade:", err)
        fmt.Fprintf(w,"Error Please Check the Logs")
    }

    conn,err:=net.Dial("tcp", nodeServer)

    if err!=nil{
        log.Println(err)
        fmt.Fprintf(w,"Could not connect to the Node Please check your connection")
    }

    defer c.Close()
    defer conn.Close()
    for {
        connbuf := bufio.NewReader(conn)
        str, err:= connbuf.ReadBytes('\n') //it is not a string really
        if err!=nil{
            log.Println(err)
            c.WriteMessage(1,[]byte("Conntection to Printer Closed"))
            break
        }
        c.WriteMessage(1,str)
    }

}


func FileUpload(w http.ResponseWriter, r *http.Request){
	var Buf bytes.Buffer
	r.ParseForm()
	file, _, err := r.FormFile("gcode")
	FileName := r.FormValue("fname")
    FCode := r.FormValue("fcode")
    nodeServer = r.FormValue("ipNode")
	
	if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    io.Copy(&Buf, file)

    contents := Buf.String()
    uploadCode(contents,FCode,FileName,nodeServer,w,r)
    fmt.Fprintf(w,"s")
}

func uploadCode(file string,FCode string,FileName string,nodeServer string,w http.ResponseWriter,r *http.Request){
	var commands int
	var speedmode bool

	GcodeSplit:=strings.Split(file, "\n")
    glen:=len(GcodeSplit)

	conn,err:=net.Dial("tcp", nodeServer+":9999")

	if err!=nil{
		log.Println(err)
		fmt.Fprintf(w,"Could not connect to the Node Please check your connection")
	}
	if (FCode=="M928"){
		fmt.Fprintf(w,"M928 %s \n",FileName) // todo: unsure about this test it please
	}else{
		fmt.Fprintf(w,"M28 %s \n",FileName)
	}

	for index, element := range GcodeSplit{
    	fmt.Fprintf(conn,element+"\n") 
    	commands+=1

        if (commands==500){ //100,80
            commands=0
            speedmode=false
            fmt.Println(float32(index)/float32(glen));
        }else{
            speedmode=true
        }

        if (speedmode!=true){
            time.Sleep(100 * time.Millisecond) //0.01 sec
        }
    }

   fmt.Fprintf(conn,"M29\n")
   if (FCode=="M23"){
   		fmt.Fprintf(conn,"M23 %s \n",FileName)
   }
   fmt.Fprintf(conn,"M27 S2\n")
   conn.Close()
   fmt.Fprintf(w,"200ok")
}
