package graph

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

var version = "0.3a"
var baseUrl = "http://reddit.com/"
var client *http.Client
var urlRegex = regexp.MustCompile(`(http(s)?://(www\.)?reddit.com)?/r/([\w]+)(/)?`)

var numReaders = 1

const numRequests = 30
const queueSize = 256

var queue chan string
var readSetChan chan map[string]bool
var numRequestsChan chan int
var failed chan string
var linkChan chan Link
var subScriberChan chan subScribers
var finish chan bool
var stop chan bool
var idle chan int

type subScribers struct {
	name string
	num  int64
}

func init() {
	//runtime.GOMAXPROCS(runtime.NumCPU())
	structInit()

	client = &http.Client{}

	readSetChan = make(chan map[string]bool, 1)
	readSetChan <- Read

	numRequestsChan = make(chan int, 1)
	numRequestsChan <- numRequests

	subScriberChan = make(chan subScribers, queueSize)

	queue = make(chan string, queueSize)
	failed = make(chan string, queueSize)
	linkChan = make(chan Link, queueSize*2)
	finish = make(chan bool)
	stop = make(chan bool)
	idle = make(chan int, numReaders-1)
}

func read(r string) {
	r = strings.ToLower(r)
	req, err := http.NewRequest("GET", baseUrl+"r/"+r+"/about.json", nil)
	if err != nil {
		fmt.Println("Error", err)
		failed <- r
		return
	}
	req.Header.Set("User-Agent", "graph-it subreddit mapper v"+version)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error", err)
		failed <- r
		return
	}
	defer resp.Body.Close()
	fmt.Println(resp.Header)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error", err)
		failed <- r
		return
	}
	fmt.Println("got answer")
	//fmt.Println("raw:", string(body))
	var js interface{}
	err = json.Unmarshal(body, &js)
	if err != nil {
		fmt.Println("Error", err)
		failed <- r
		return
	}
	//fmt.Println("parsed:")
	//fmt.Println(js)
	info := js.(map[string]interface{})
	//now find links
	for _, v := range info {
		switch v.(type) {
		case map[string]interface{}:
			parse(r, v.(map[string]interface{}))
		default:
			//fmt.Println("ignoring key", k, "in", r)
		}
	}
	fmt.Println()
}

func parse(from string, data map[string]interface{}) {
	for k, v := range data {
		switch k {
		case "description":
			parseDesc(from, v.(string))
		case "subscribers":
			subScriberChan <- subScribers{name: from,
				num: int64(v.(float64))}
		default:
			//fmt.Println("ignoring key", k, "in", from)
		}
	}
}

func parseDesc(from, d string) {
	//fmt.Println("description:",d)
	all := urlRegex.FindAllStringSubmatch(d, -1)
	for _, l := range all {
		to := strings.ToLower(l[4])

		// if not queried yet, try.
		m := <-readSetChan
		if !m[to] {
			m[to] = true
			queue <- to
		}
		readSetChan <- m

		linkChan <- Link{From: from, To: to}
		fmt.Println("from", from, "to", to)
	}
}

func reader() {
	for on, wasIdle := true, false; on; {
		select {
		case r := <-queue:
			if wasIdle {
				<-idle
			}
			left := <-numRequestsChan
			left--
			numRequestsChan <- left
			if left < 0 {
				on = false
				fmt.Println("Out of requests, shutting down...")
				failed <- r
			} else {
				fmt.Print("reading ", r, "... ")
				read(r)
			}
		case idle <- 1:
			wasIdle = true
			fmt.Println("reader idle")
		default:
			if !wasIdle {
				on = false
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
		case link := <-linkChan:
			from, ok := Subs[link.From]
			if !ok {
				from = &Sub{Name: link.From}
				Subs[link.From] = from
			}
			to, ok := Subs[link.To]
			if !ok {
				to = &Sub{Name: link.To}
				Subs[link.To] = to
			}
			from.Out++
			to.In++
			Links = append(Links, link)
		case s := <-subScriberChan:
			sub, ok := Subs[s.name]
			if !ok {
				sub = &Sub{Name: s.name}
				Subs[s.name] = sub
			}
			sub.Subscribers = s.num
		case <-linkerStop:
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

	//put stuff into the queue
	for len(queue) < queueSize && len(Failed) > 0 {
		queue <- Failed[len(Failed)-1]
		Failed = Failed[:len(Failed)-1]
	}

	fmt.Println("starting readers")
	for i := 0; i < numReaders; i++ {
		go reader()
	}

	fmt.Println("starting linker")
	linkerStop := make(chan bool)
	linkerFin := make(chan bool)
	go linker(linkerStop, linkerFin)

	// go on while not all readers are idle
	for numReaders > 0 {
		select {
		case <-finish:
			numReaders--
			fmt.Println("numReaders:", numReaders)
		case s := <-failed:
			Failed = append(Failed, s)
			fmt.Println("URL failed:", s)
		}
	}
	// stop linker
	fmt.Println("stopping linker")
	linkerStop <- true
	// wait for linker
	fmt.Println("waiting for linker")
	<-linkerFin
	//finished.
	fmt.Println("reading finished, draining queue")
	// drain queue
	for len(queue) > 0 {
		Failed = append(Failed, <-queue)
	}
	fmt.Print("ToDo:")
	// remove queue from read
	for _, s := range Failed {
		delete(Read, s)
		fmt.Printf("%s ", s)
	}
	fmt.Println()

}
