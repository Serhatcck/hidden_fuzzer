package hidden_fuzzer

import (
	"log"
	"math/rand"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type Worker struct {
	Config           *Config
	Runner           *SimpleRunner
	WorkQueue        []WorkQueue
	SleepWg          sync.WaitGroup
	ProssesWg        sync.WaitGroup
	Sleep            bool
	Fail             bool
	Error            error
	RootInformation  Response
	FoundUrls        []FoundUrl
	FoundPaths       []FoundPath
	TargetPaths      []FoundPath
	DuplicateIndexes []DuplicateIndexes

	mainSlash          int
	currentTarget      string
	queueWorkCounter   int
	goroutineCountChan chan int
}

type FoundPath struct {
	Path string
}

type FoundUrl struct {
	Response           Response
	Request            Request
	DuplicateIndexId   int // for analyze end of method
	IsRedirect         bool
	IsAddedFolderArray bool
}

type DuplicateIndexes struct {
	Url     string
	Counter int
	Body    string
	Index   int
}

type WorkQueue struct {
	Url            string
	Req            Request
	RedirectConter int
}

func NewWorker(conf *Config) *Worker {
	return &Worker{
		Config:    conf,
		Sleep:     false,
		WorkQueue: make([]WorkQueue, 0),
		Runner:    NewSimpleRunner(conf, false),
	}
}

func (w *Worker) Start() {
	// Gorutin sayısını takip etmek için bir kanal oluştur

	w.goroutineCountChan = make(chan int)

	go w.showProgress(w.goroutineCountChan)

	err := w.doMainReq()
	if err != nil {
		log.Fatal(err)
	}
	w.mainSlash = slashCounter(w.RootInformation.URL)
	w.currentTarget = w.Config.Url.String()
	w.start(w.Config.Url.String())
	//w.ProssesWg.Wait()

	if len(w.FoundPaths) > 0 {
		w.startSubFolderExection()
	}
}

func (w *Worker) start(path string) {
	w.ProssesWg.Add(1)
	defer w.ProssesWg.Done()

	w.WorkQueue = nil

	for _, newUrl := range w.Config.Wordlist {

		newUrl := makeUrl(path, newUrl)
		w.WorkQueue = append(w.WorkQueue, WorkQueue{Url: newUrl, Req: Request{
			URL:     newUrl,
			Headers: w.Config.Headers,
			Method:  w.Config.Method,
		}, RedirectConter: 0})
	}
	w.startExecution()
	w.AnalyzeSubFolder()

}

func (w *Worker) startSubFolderExection() {

	w.WorkQueue = nil
	for _, path := range w.TargetPaths {
		w.currentTarget = path.Path
		w.start(path.Path)
	}

	w.ProssesWg.Wait()
	if len(w.TargetPaths) > 0 {
		w.startSubFolderExection()
	}
}

func (w *Worker) startExecution() error {
	w.queueWorkCounter = 0
	time.Sleep(time.Second * 1)
	var wg sync.WaitGroup
	threadlimiter := make(chan bool, w.Config.Threads)

	for _, queue := range w.WorkQueue {
		w.SleepWg.Wait()
		if w.Error != nil {

			w.WorkQueue = nil
			return w.Error

		}

		wg.Add(1)
		threadlimiter <- true

		// İstek gönderilmeden önce gorutin sayısını artır
		w.goroutineCountChan <- 1

		go func(queue WorkQueue) {
			defer func() { <-threadlimiter }()
			defer wg.Done()
			w.executeTask(queue)
		}(queue)
	}
	wg.Wait()
	w.AnalyzeDuplicate()
	return nil
}

func (w *Worker) AnalyzeSubFolder() {
	w.TargetPaths = nil
	for idx, url := range w.FoundUrls {
		if url.Response.StatusCode == 403 && !url.IsAddedFolderArray {
			w.FoundUrls[idx].IsAddedFolderArray = true
			w.FoundPaths = append(w.FoundPaths, FoundPath{
				Path: url.Request.URL,
			})
			pathSlash := slashCounter(url.Request.URL)
			//Depth Counter  Check
			if (pathSlash - w.mainSlash) < w.Config.Depth {
				w.TargetPaths = append(w.TargetPaths, FoundPath{
					Path: url.Request.URL,
				})
			}

		}
	}

}

func (w *Worker) showProgress(goroutineCountChan chan int) {
	var reqPerSecondCount int
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// Her bir saniyede bir toplam gorutin sayısını ekrana bas
			//goroutinesPerSecondChan <- reqPerSecondCount
			// counter'ı sıfırla

			WriteStatus(w, reqPerSecondCount, w.queueWorkCounter)
			reqPerSecondCount = 0
		case delta := <-goroutineCountChan:
			// Gorutin sayısını güncelle
			reqPerSecondCount += delta
			w.queueWorkCounter += delta

		}
	}

}

func (w *Worker) AnalyzeDuplicate() {

	newUrls := []FoundUrl{}
	for _, found := range w.FoundUrls {
		for _, duplicate := range w.DuplicateIndexes {
			if found.DuplicateIndexId == duplicate.Index {
				if duplicate.Counter <= w.Config.DuplicateCounter {
					newUrls = append(newUrls, found)
				}
				continue
			}
		}
	}
	w.FoundUrls = newUrls
}

func (w *Worker) executeTask(queue WorkQueue) {
	if queue.RedirectConter > w.Config.RedirectCounter {
		return
	}
	tmpResponse, err := w.Runner.Execute(&queue.Req)
	if err != nil {
		if !w.Fail {
			w.sleep()
			w.failureCheck(0)

		}

	} else {
		if MainCheck(w.RootInformation, tmpResponse) {
			return
		}
		//if redirect
		if tmpResponse.StatusCode >= 300 && tmpResponse.StatusCode < 400 {
			if tmpResponse.Headers["Location"] != nil {
				redirectLocation := tmpResponse.Headers["Location"][0]
				u, err := url.Parse(redirectLocation)

				//URL değilse
				if err != nil || u.Host == "" {

					newUrl := makeUrl(w.Config.Url.String(), tmpResponse.Headers["Location"][0])
					queue.Req.URL = newUrl
					queue.RedirectConter += 1
					w.executeTask(queue)
					return
				} else {

					// redirect aynı URL e
					if u.Host == w.RootInformation.Request.Host {
						queue.Req.URL = redirectLocation
						queue.RedirectConter += 1
						w.executeTask(queue)
						return
					}
				}

			}
		}
		stat, index := DuplicateCheck(tmpResponse, w)
		if stat {
			return
		}

		isHave := IsUrlAppend(tmpResponse, w)
		if isHave {
			return
		}
		w.FoundUrls = append(w.FoundUrls, FoundUrl{
			Response:           tmpResponse,
			Request:            queue.Req,
			DuplicateIndexId:   index,
			IsRedirect:         queue.RedirectConter > 0,
			IsAddedFolderArray: false,
		})
		WriteFound(w, queue.Req.URL+": "+strconv.Itoa(tmpResponse.StatusCode))

	}
}

func (w *Worker) failureCheck(indis int) {

	w.Fail = true

	time.Sleep(time.Second * time.Duration(w.Config.Timeout))
	_, err := sendHTTPRequest(*w.Runner.client, Request{
		URL:     w.Config.Url.String(),
		Headers: w.Config.Headers,
	})

	if err != nil {
		if indis > w.Config.FailureCounter {
			WriteFailure(w, "Failure Check Error:"+err.Error())
			//failure got resume and finish
			w.resume()
			return
		}
		w.Error = err
		w.failureCheck(indis + 1)
	} else {
		w.Error = nil
		w.Fail = false
		w.resume()
	}
}

func (w *Worker) doMainReq() error {

	mainResp, err := w.Runner.Execute(&Request{
		URL:     w.Config.Url.String(),
		Headers: w.Config.Headers,
		Host:    w.Config.Url.Host,
	})
	if err != nil {
		return err
	}
	if mainResp.StatusCode == 403 {
		seed := time.Now().UnixNano()
		r := rand.New(rand.NewSource(seed))

		// Rastgele 5 karakterden oluşan bir dize oluştur
		var randomUrl = string(r.Intn(26)+65) + string(r.Intn(20)+65) + string(r.Intn(26)+65) + string(r.Intn(26)+65)
		newUrl := makeUrl(w.Config.Url.String(), randomUrl)

		mainResp, err = w.Runner.Execute(&Request{
			URL:     newUrl,
			Headers: w.Config.Headers,
		})
		if err != nil {
			return err
		}
		w.RootInformation = mainResp
	} else {
		w.RootInformation = mainResp
	}

	return nil

}

func (w *Worker) sleep() {
	if !w.Sleep {
		w.Sleep = true
		w.SleepWg.Add(1)
	}
}

func (w *Worker) resume() {
	if w.Sleep {
		w.Sleep = false
		w.SleepWg.Done()
	}
}
