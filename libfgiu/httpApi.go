package libfgiu

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"log"
	"strconv"
	"sync"
	"time"
)

const (
	defServAddr        = "127.0.0.1:11451"
	defApiCertFileName = "fuzzgiu.cert"
	defApiCertKeyName  = "fuzzgiu.key"
)

type httpService struct {
	addr        string
	tls         bool
	accessToken string
	e           *echo.Echo
	wg          sync.WaitGroup
}

func (s *httpService) wait() {
	s.wg.Wait()
}

func getApiConfig(config WebApiConfig) (servAddr string, startTLS bool, certFile string, certKey string) {
	if servAddr = config.ServAddr; servAddr == "" {
		servAddr = defServAddr
	}
	if certFile = config.CertFileName; certFile == "" {
		certFile = defApiCertFileName
	}
	if certKey = config.CertKeyName; certKey == "" {
		certKey = defApiCertKeyName
	}
	startTLS = config.TLS
	return
}

func errorMsg(err error) map[string]string {
	return map[string]string{"error": err.Error()}
}

func (f *Fuzzer) startHttpApi(apiConf WebApiConfig) error {
	addr, startTLS, certFileName, certKeyName := getApiConfig(apiConf)

	stopChan := make(chan struct{})

	go func() {
		<-stopChan
		time.Sleep(20 * time.Millisecond)
		f.Stop()
	}()

	acToken := common.RandMarker()
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	auth := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenHeader := c.Request().Header.Get("Access-Token")
			if tokenHeader != acToken {
				return c.NoContent(401)
			}
			return next(c)
		}
	}

	getJobById := func(c echo.Context) error {
		jid, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.JSON(400, errorMsg(fmt.Errorf("failed to parse id: %v", err)))
		}
		jc, exist := f.GetJob(int(jid))
		if !exist {
			return c.JSON(404, errorMsg(fmt.Errorf("job#%d not found", jid)))
		}
		cloned := jc.Job.Clone()
		jc.Release()
		return c.JSON(200, cloned)
	}

	delJobById := func(c echo.Context) error {
		jid, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.JSON(400, errorMsg(fmt.Errorf("failed to parse job id: %v", err)))
		}
		err = f.StopJob(int(jid))
		if err != nil {
			return c.JSON(500, errorMsg(fmt.Errorf("failed to stop job#%d: %v", jid, err)))
		}
		return c.NoContent(204)
	}

	submitJob := func(c echo.Context) error {
		newJob := new(fuzzTypes.Fuzz)
		err := c.Bind(newJob)
		if err != nil {
			return c.JSON(400, errorMsg(fmt.Errorf("failed to unmarshal job: %v", err)))
		}
		jid, err := f.Submit(newJob)
		if err != nil {
			return c.JSON(500, errorMsg(fmt.Errorf("failed to submit job: %v", err)))
		}
		return c.JSON(200, map[string]int{"jid": jid})
	}

	getJids := func(c echo.Context) error {
		return c.JSON(200, f.GetJobIds())
	}

	stopFuzzer := func(c echo.Context) error {
		stopChan <- struct{}{}
		return c.JSON(200, map[string]string{"status": "stopped"})
	}

	// restful api
	g := e.Group("/", auth)
	g.GET("job/:id", getJobById)
	g.DELETE("job/:id", delJobById)
	g.POST("job", submitJob)
	g.GET("jobIds", getJids)
	g.GET("stop", stopFuzzer)

	f.s = &httpService{
		e:           e,
		addr:        addr,
		tls:         startTLS,
		accessToken: acToken,
	}

	go func() {
		f.s.wg.Add(1)
		defer f.s.wg.Done()
		var err error
		if startTLS {
			err = e.StartTLS(addr, certFileName, certKeyName)
		} else {
			err = e.Start(addr)
		}
		if err != nil && err.Error() != "http: Server closed" {
			log.Fatal(err)
		}
	}()
	return nil
}
