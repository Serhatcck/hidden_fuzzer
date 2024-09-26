package hidden_fuzzer

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"
)

func getURL(str string) (*url.URL, bool) {
	u, err := url.Parse(str)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, true
	}
	return u, false
}

func getHeeaderToString(headers map[string][]string) string {
	var resp = ""
	for name, values := range headers {
		var valueString = ""
		for _, value := range values {
			valueString += value + " "
		}
		resp += name + ":" + valueString
	}
	return resp
}

func readFileLines(filename string) ([]string, error) {
	// Dosyanın var olup olmadığını kontrol etme
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("dosya bulunamadı: %s", filename)
	}

	// Dosyayı okuma
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Dosya içeriğini satır satır ayırma
	lines := strings.Split(string(fileContent), "\n")

	// Satır sonlarındaki boşlukları temizleme
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}

	return lines, nil
}

func makeUrl(path string, endpoint string) string {

	slashStatPath := strings.HasSuffix(path, "/")
	slashStatEndpoint := strings.HasPrefix(endpoint, "/")

	if slashStatPath {
		path = strings.TrimSuffix(path, "/")
	}

	if slashStatEndpoint {
		endpoint = strings.TrimPrefix(endpoint, "/")
	}
	return path + "/" + endpoint
}

func makeUrlWithParameter(path string, parameter string) string {

	newString := ""
	if strings.HasPrefix(parameter, "/") {
		newString = strings.TrimPrefix(parameter, "/")
	} else {
		newString = parameter
	}

	if !strings.HasPrefix(parameter, "?") {
		newString = "?" + newString + "="
	}

	newString = path + newString
	return newString

}

func addExtensionToPath(path string, extension string) string {
	newPath := ""
	dotStat := strings.HasPrefix(extension, ".")
	if !dotStat {
		extension = "." + extension
	}

	slashStat := strings.HasSuffix(path, "/")
	if slashStat {
		newPath = strings.TrimSuffix(path, "/") + extension
	} else {
		newPath = path + extension
	}

	return newPath
}

func hasExtension(u string) bool {
	parsedURL, _ := url.Parse(u)
	segments := strings.Split(parsedURL.Path, "/")
	lastSegment := segments[len(segments)-1]
	ext := path.Ext(lastSegment)
	return ext != ""
}

func slashCounter(url string) int {
	return strings.Count(url, "/")
}
