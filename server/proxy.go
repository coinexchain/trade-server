package server

import (
	"io/ioutil"
	"net/http"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	SwaggerPath = "/swagger/swagger.yaml"
)

type RestSwagger struct {
	Paths map[string]interface{} `yaml:"paths"`
}

func registerProxyHandlerLegacy(prefix, lcd string, router *mux.Router) error {
	paths, err := getRestPaths(lcd)
	if err != nil {
		log.WithError(err).Fatal("get rest paths failed")
		return err
	}
	for _, path := range paths {
		if strings.HasPrefix(path, "/txs/") ||
			strings.HasPrefix(path, "/blocks/") ||
			strings.HasPrefix(path, "/validatorsets/") {
			router.HandleFunc(prefix+path, httpProxy(lcd, prefix))
		}
	}
	return nil
}

func registerProxyHandler(lcd string, router *mux.Router) error {
	paths, err := getRestPaths(lcd)
	if err != nil {
		log.WithError(err).Fatal("get rest paths failed")
		return err
	}
	for _, path := range paths {
		router.HandleFunc(path, httpProxy(lcd, ""))
	}
	return nil
}

func getRestPaths(lcd string) ([]string, error) {
	url := lcd + SwaggerPath
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	restSwagger := &RestSwagger{}
	if err = yaml.Unmarshal(body, restSwagger); err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(restSwagger.Paths))
	for path := range restSwagger.Paths {
		paths = append(paths, path)
	}
	return paths, nil
}

func httpProxy(lcd, prefix string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		client := &http.Client{}
		path := strings.TrimPrefix(request.URL.Path, prefix)
		req, err := http.NewRequest("GET", lcd+path, nil)
		if err != nil {
			log.WithError(err).Error("http new request error")
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			log.WithError(err).Error("http client do failed")
			return
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.WithError(err).Error("read response body failed")
			return
		}

		writer.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		if _, err = writer.Write(body); err != nil {
			log.WithError(err).Error("write response failed")
			return
		}
	}
}
