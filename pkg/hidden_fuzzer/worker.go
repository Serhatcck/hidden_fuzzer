package hidden_fuzzer

import (
	"log"
	"math/rand"
	"net/url"
	"strings"
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

	mainSlash           int
	currentTarget       string
	queueWorkCounter    int
	goroutineCountChan  chan int
	isrunning           bool
	isInteractiveWindow bool
	skipQueue           bool
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
	Url        string
	StatusCode int
	Counter    int
	Body       string
	Index      int
}

type WorkQueue struct {
	Url            string
	Req            Request
	RedirectConter int
	RedirectQueue  []Response
}

func NewWorker(conf *Config) *Worker {
	return &Worker{
		Config:    conf,
		Sleep:     false,
		WorkQueue: make([]WorkQueue, 0),
		Runner:    NewSimpleRunner(conf, false),
	}
}

func (w *Worker) Start() error {
	go func() {
		err := HandleInteractive(w)
		if err != nil {
			log.Printf("Error while trying to initialize interactive session: %s", err)
		}
	}()
	// Gorutin sayısını takip etmek için bir kanal oluştur
	w.isInteractiveWindow = false
	w.isrunning = true
	w.goroutineCountChan = make(chan int)

	go w.showProgress(w.goroutineCountChan)
	w.mainSlash = slashCounter(w.Config.Url.String())
	w.currentTarget = w.Config.Url.String()
	err := w.start(w.Config.Url.String())
	if err != nil {
		w.isrunning = false
		return err
	}
	//w.ProssesWg.Wait()

	if len(w.FoundPaths) > 0 {
		subFulderErr := w.startSubFolderExection()
		if subFulderErr != nil {
			w.isrunning = false
			return subFulderErr
		}
	}
	w.isrunning = false
	return nil
}

func (w *Worker) start(path string) error {
	err := w.doMainReq(path)
	if err != nil {
		return err
	}
	w.ProssesWg.Add(1)
	defer w.ProssesWg.Done()

	w.WorkQueue = nil

	for _, newUrl := range w.Config.Wordlist {

		if w.Config.ParamFuzing {
			newUrl = makeUrlWithParameter(path, newUrl)
			newUrl += w.Config.ParamValue
		} else {
			newUrl = makeUrl(path, newUrl)
		}

		w.WorkQueue = append(w.WorkQueue, WorkQueue{Url: newUrl, Req: Request{
			URL:     newUrl,
			Headers: w.Config.Headers,
			Method:  w.Config.Method,
		}, RedirectConter: 0})
	}
	execErr := w.startExecution()
	if execErr != nil {
		return execErr
	}
	w.AnalyzeSubFolder()
	return nil

}

func (w *Worker) startSubFolderExection() error {

	w.WorkQueue = nil
	for _, path := range w.TargetPaths {
		w.currentTarget = path.Path
		err := w.start(path.Path)
		if err != nil {
			return err
		}
	}

	w.ProssesWg.Wait()
	if len(w.TargetPaths) > 0 {
		err := w.startSubFolderExection()
		if err != nil {
			return err
		}
	}
	return nil
}

// TO DO Rate Limiti Daha İyi Hale Getir
func (w *Worker) startExecution() error {
	w.queueWorkCounter = 0
	time.Sleep(time.Second * 1)
	var wg sync.WaitGroup
	threadlimiter := make(chan bool, w.Config.Threads)

	// Rate limit için gerekli olan parametreler
	// performansı düşürüyor

	rateLimit := time.Second / time.Duration(w.Config.RateLimit) // İstenen hızda bir zaman dilimi
	lastRequestTime := time.Now()                                // Son istek zamanı

	for _, queue := range w.WorkQueue {

		// Rate limit kontrolü
		if w.Config.UseRateLimit {
			elapsed := time.Since(lastRequestTime)
			if elapsed < rateLimit {
				// Beklemek gerekiyor
				time.Sleep(rateLimit - elapsed)
			}
			lastRequestTime = time.Now()
		}

		w.SleepWg.Wait()
		if w.Error != nil {
			w.WorkQueue = nil
			return w.Error
		}

		wg.Add(1)
		threadlimiter <- true

		// İstek gönderilmeden önce gorutin sayısını artır
		w.goroutineCountChan <- 1

		if w.skipQueue {
			//queue skipped
			w.skipQueue = false
			return nil
		}

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
		//if url has an extension
		//.gitignore
		//.git
		//test.aspx
		// is not an folder. It's f/p
		//For this reason code use hasExtension helper func
		if url.Response.StatusCode == 403 && !url.IsAddedFolderArray && !hasExtension(url.Request.URL) {

			w.FoundUrls[idx].IsAddedFolderArray = true
			w.FoundPaths = append(w.FoundPaths, FoundPath{
				Path: url.Request.URL,
			})
			pathSlash := slashCounter(url.Request.URL)
			//Depth Counter  Check
			if (pathSlash - w.mainSlash) <= w.Config.Depth {
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
			if !w.isrunning {
				return
			}

			if !w.isInteractiveWindow {
				WriteStatus(w, reqPerSecondCount, w.queueWorkCounter)
			}
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
	//wait for anything
	//w.SleepWg.Wait()
	if queue.RedirectConter > w.Config.RedirectCounter {
		return
	}
	tmpResponse, err := w.Runner.Execute(&queue.Req)
	if err != nil {
		if !w.Fail {
			w.sleep()
			w.failureCheck(0)

		} /*else {
			w.SleepWg.Add(1)
			//got error try once
			w.executeTask(queue)
		}*/

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

				var newUrl = ""
				if err != nil || u.Host == "" {
					if strings.HasPrefix("/", u.Host) {
						//location -> /test -> http://host/test
						//location -> test -> http://host/path/test
						newUrl = makeUrl(w.RootInformation.Request.Schema+"://"+w.RootInformation.Request.Host, u.Path)
					} else {
						newUrl = makeUrl(w.Config.Url.String(), tmpResponse.Headers["Location"][0])

					}
				} else {
					// redirect aynı URL e
					if u.Hostname() == w.RootInformation.Request.Host {
						newUrl = redirectLocation
					}
				}

				if newUrl != "" {
					queue.Req.URL = newUrl
					queue.RedirectConter += 1
					queue.RedirectQueue = append(queue.RedirectQueue, tmpResponse)
					w.executeTask(queue)
					return
				}

			}
		}
		index := 0
		//if tmpResponse.StatusCode != 403 {
		stat, idx := DuplicateCheck(tmpResponse, w)
		index = idx
		if stat {
			return
		}
		//}

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
		WriteFound(w, queue, tmpResponse)

	}
}

func (w *Worker) failureCheck(indis int) {

	w.Fail = true

	time.Sleep(time.Second * time.Duration(w.Config.FailureCheckTimeout))
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

func (w *Worker) doMainReq(path string) error {
	//change mainResp for all subfolder
	mainResp, err := w.Runner.Execute(&Request{
		URL:     path,
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
		newUrl := makeUrl(path, randomUrl)

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
