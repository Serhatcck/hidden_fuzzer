package hidden_fuzzer

import (
	"github.com/hyperjumptech/beda"
)

var notFoundStatus = [2]int{404, 429}

func isSimilar(str1 string, str2 string) bool {
	sd := beda.NewStringDiff(str1, str2)
	//lDist := sd.LevenshteinDistance()
	//tDiff := sd.TrigramCompare()
	jDiff := sd.JaroDistance()
	jwDiff := sd.JaroWinklerDistance(0.1)

	//fmt.Printf("Levenshtein Distance is %d \n", lDist)
	//fmt.Printf("Trigram Compare is is %f \n", tDiff)
	//fmt.Printf("Jaro Distance is is %f \n", jDiff)        // > %0.80
	//fmt.Printf("Jaro Wingkler Distance is %f \n", jwDiff) // > %0.90*/
	//fmt.Printf("%d ,%d \n", jDiff, jwDiff)

	//TO DO set this percents from conf
	if jDiff > 0.80 && jwDiff > 0.90 {
		return true
	} else {
		return false
	}
}

// if matched so if not show return true
func MainCheck(rootInfo Response, newInfo Response) bool {
	//main response and this response is smilar?
	if isSimilar(rootInfo.Body, newInfo.Body) {
		return true
	}
	//this response status code is not found?
	for _, stat := range notFoundStatus {
		if newInfo.StatusCode == stat {
			return true
		}
	}
	return false
}

// if not matcher return index value of response in duplicateIndexes array
func DuplicateCheck(resp Response, w *Worker) (bool, int) {
	for idx, duplicate := range w.DuplicateIndexes {
		if resp.Body == "" {
			//if response has not body set header to body
			if len(resp.Headers["Location"]) > 0 {
				resp.Body = resp.Headers["Location"][0] //# for lowercase check it
			} else if len(resp.Headers["location"]) > 0 {
				resp.Body = resp.Headers["location"][0] //# for lowercase check it
			} else {
				resp.Body = "EMPTY"
			}
			//resp.Body = getHeeaderToString(resp.Headers)
		}
		if isSimilar(resp.Body, duplicate.Body) {
			w.DuplicateIndexes[idx].Counter++
			return w.DuplicateIndexes[idx].Counter > w.Config.DuplicateCounter, duplicate.Index // if duplicate counter is a 49+ duplicate check is matched return true / else duplicate check is not matched return false
		}
	}
	//no match found
	index := AppendDuplicateIndex(resp, w)
	return false, index
}

func AppendDuplicateIndex(resp Response, w *Worker) int {
	index := len(w.DuplicateIndexes) + 1
	tmp := DuplicateIndexes{
		Url:     resp.URL,
		Counter: 1,
		Body:    resp.Body,
		Index:   index,
	}
	w.DuplicateIndexes = append(w.DuplicateIndexes, tmp)
	return index
}

func IsUrlAppend(resp Response, w *Worker) bool {

	for _, url := range w.FoundUrls {
		if resp.Request.URL == url.Request.URL {
			return true
		}
	}
	return false
}
