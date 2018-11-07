
package main

import (
"fmt"
"time"
"bytes"
"io"
"strings"
"strconv"
"net/http"
"bufio"
"log"
"net"
"github.com/gorilla/mux"
"github.com/gorilla/websocket"
"github.com/skratchdot/open-golang/open")

var nodeServer = ""
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

func parse_Status(str []byte) ([]byte,bool){
    st:=string(str);
    if strings.HasPrefix(st, "SD printing"){
        splitStatus:=strings.Split(strings.Split(st," ")[3],"/")
        st,_:=strconv.ParseFloat(splitStatus[0], 32)
        nd,_:=strconv.ParseFloat(splitStatus[1], 32)
        return []byte(fmt.Sprintf("%.2f", nd/st)),true
    }
    return []byte("No update"),false
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
        perc,Send:=parse_Status(str);
        if Send {
            c.WriteMessage(1,perc);
        }
    }

}


func TrimSuffix(s, suffix string) string {
    if strings.HasSuffix(s, suffix) {
        s = s[:len(s)-len(suffix)]
    }
    return s
}

func FileUpload(w http.ResponseWriter, r *http.Request){
	var Buf bytes.Buffer
	r.ParseForm()
	file, _, err := r.FormFile("gcode")
	FileName := TrimSuffix(r.FormValue("fname"),"de")
    
    FCode := r.FormValue("fcode")
    nodeServer = r.FormValue("ipNode")
	
	if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    io.Copy(&Buf, file)

    contents := Buf.String()
    uploadCode(contents,FCode,FileName,w,r)
    
}

func uploadCode(file string,FCode string,FileName string,w http.ResponseWriter,r *http.Request){
	var commands int
	var speedmode bool

	GcodeSplit:=strings.Split(file, "\n")
    glen:=len(GcodeSplit)

	conn,err:=net.Dial("tcp", nodeServer+":9999")

    time.Sleep(5000 * time.Millisecond)
	if err!=nil{
		log.Println(err)
		fmt.Fprintf(w,"Could not connect to the Node Please check your connection")
	}
    fmt.Fprintf(w,"Done! Your file has been uploaded <a href='/'>Click Here</a> to go back")
	if (FCode=="M928"){
        fmt.Fprintf(conn,"\r\nM928 "+FileName+"\r\n") // todo: unsure about this test it please
    }else{
        fmt.Fprintf(conn,"\r\nM28 "+FileName+"\r\n")
    }
    time.Sleep(1000 * time.Millisecond) 
    
    for index, element := range GcodeSplit{
    	fmt.Fprintf(conn,"\r\n"+element+"\n\r") 
    	commands+=1

        if (commands==80){
            commands=0
            speedmode=false
            fmt.Println(float32(index)/float32(glen));
        }else{
            speedmode=true
            time.Sleep(1 *time.Millisecond)
        }

        if (speedmode==false){
            time.Sleep(100 * time.Millisecond) //0.1 sec
        }
    }

   fmt.Fprintf(conn,"\r\n M29 "+FileName+"\r\n") //somehow doesn't work unless I do this,Even though similar Projects don't do the same
   time.Sleep(1000 *time.Millisecond)
   if (FCode=="M23"){
   		fmt.Fprintf(conn,"\r\nM23 "+FileName+"\r\n")
        fmt.Fprintf(conn,"\r\nM24\r\n")
   }
   if (FCode=="M23" || FCode=="M928"){
        fmt.Fprintf(conn,"\r\nM27 S2\r\n") 
   }
   conn.Close()
}
