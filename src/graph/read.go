package graph

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"runtime"
	"regexp"
)

var version = "0.1a"
var baseUrl = "http://reddit.com/"
var client *http.Client
var urlRegex = regexp.MustCompile(`\(http(s)?://(www\.)?reddit.com/r/([\w]+)(/)?\)`)

var numReaders = 1
const queueSize = 256

var allInQueue map[string]bool

var queue chan string
var allQueueChan chan map[string]bool
var numRequestsChan chan int
var failed chan string
var subChan chan Sub
var linkChan chan Link
var finish chan bool
var stop chan bool
var idle chan int

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	client = &http.Client{}
	
	allInQueue = make(map[string]bool)
	allQueueChan = make(chan map[string]bool, 1)
	allQueueChan <- allInQueue
	
	numRequestsChan = make(chan int, 1)
	numRequestsChan <- 30
	
	queue = make(chan string, queueSize)
	failed = make(chan string, queueSize)
	subChan = make(chan Sub, queueSize)
	linkChan = make(chan Link, queueSize*2)
	finish = make(chan bool)
	stop = make(chan bool)
	idle = make(chan int, numReaders)
}

func read(r string) {
	req, err := http.NewRequest("GET", baseUrl+"/r/"+r+"/about.json", nil)
	if err != nil {
		fmt.Println("Error", err);
		failed <- r
		return
	}
	req.Header.Set("User-Agent", "graph-it subreddit mapper v"+version)
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error", err);
		failed <- r
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error", err);
		failed <- r
		return
	}
	//fmt.Println("raw:", string(body))
	
	var js interface{}
	err = json.Unmarshal(body, &js)
	if err != nil {
		fmt.Println("Error",err)
		failed <- r
		return
	}
	fmt.Println()
	//fmt.Println("parsed:")
	//fmt.Println(js)
	info := js.(map[string]interface{})
	sub := &Sub{Name: r}
	
	//now find links
	for k, v := range info {
		switch v.(type) {
			case map[string]interface{}:
				sub.parse(r, v.(map[string]interface{}))
			default:
				fmt.Println("ignoring key",k)
		}
	}
	fmt.Println(sub)
}

func (sub *Sub) parse(from string, data map[string]interface{}) {
	for k, v := range data {
		switch k {
			case "description":
				sub.parseDesc(from, v.(string))
			case "subscribers":
				sub.Subscribers = int64(v.(float64))
			default:
				fmt.Println("ignoring key",k)
		}
	}
}

func (sub *Sub) parseDesc(from, d string) {
	fmt.Println("description:",d)
	all := urlRegex.FindAllStringSubmatch(d,-1)
	for _, l := range all {
		to := l[3]
		
		// if not queried yet, try.
		m := <- allQueueChan
		if !m[to] {
			m[to] = true
			queue <- to
		}
		allQueueChan <- m
		
		linkChan <- Link{From: from, To: to}
		fmt.Println("from", from, "to", to)
	}
}

func reader() {
	for on, wasIdle := true, false; on; {
		select {
			case <- stop:
				on = false
			case r := <- queue:
				if wasIdle {
					<- idle
				}
				left := <- numRequestsChan
				left--
				numRequestsChan <- left
				if left < 0 {
					on = false
					fmt.Println("Out of requests, shutting down...")
					failed <- r
				} else {
					fmt.Println("reader reading",r)
					read(r)
				}
			case idle <- 1:
				wasIdle = true
				fmt.Println("reader idle")
			default:
				// this should only happen when the queue is empty and idle is full
				on = false
				select {
					case stop <- true:
					default:
				}
				
		}
	}
	finish <- true
}

func linker(linkerStop chan bool, linkerFin chan bool) {
	readyToStop := false
	for on := true; on; {
		// linker runs as long if there are things in queue
		select {
			case sub := <- subChan:
				Subs[sub.Name] = &sub
			case link := <- linkChan:
				Subs[link.From].Out++
				Subs[link.To].In++
				Links[link] = true
			case <- linkerStop:
				readyToStop = true
			default:
				// queues are emtpy
				if readyToStop {
					on = false
				}
		}
	}
	linkerFin <- true
}

func Start() {
	fmt.Println("putting stuff")
	// append from failed to queue until the queue is full
	i := len(failed)
	for num := 0; i > 0 && num <= queueSize; i, num = i-1, num-1 {
		queue <- Failed[i]
	}
	// re-slice Failed
	Failed = Failed[:i]

	fmt.Println("starting readers")
	for i := 0; i < numReaders;i++ {
		go reader()
	}
	
	fmt.Println("starting linker")
	linkerStop := make(chan bool)
	linkerFin := make(chan bool)
	go linker(linkerStop, linkerFin)
	
	// go on while not all readers are idle
	for ; numReaders > 0; {
		select {
			case <- finish:
				numReaders--
				fmt.Println("numReaders:",numReaders)
			case s := <- failed:
				Failed = append(Failed,s)
				fmt.Println("URL failed:",s)
		}
	}
	// stop linker
	fmt.Println("stopping linker")
	linkerStop <- true
	// wait for linker
	fmt.Println("waiting for linker")
	<- linkerFin
	//finished.
	fmt.Println("reading finished")
}
