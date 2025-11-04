package libfgiu

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"strconv"
	"sync"
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

func getApiConfig(config *WebApiConfig) (servAddr string, startTLS bool, certFile string, certKey string) {
	if config == nil {
		servAddr, certFile, certKey = defServAddr, defApiCertFileName, defApiCertKeyName
		return
	}
	if servAddr = config.ServAddr; servAddr == "" {
		servAddr = defServAddr
	}
	if certFile = config.CertFileName; certFile == "" {
		certFile = defApiCertFileName
	}
	if certKey = config.CertKeyName; certKey == "" {
		certKey = defApiCertKeyName
	}
	startTLS = config.Tls
	return
}

func errorMsg(err error) map[string]string {
	return map[string]string{"error": err.Error()}
}

func (f *Fuzzer) startHttpApi(apiConf *WebApiConfig) error {
	addr, startTLS, certFileName, certKeyName := getApiConfig(apiConf)

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
			return c.JSON(400, errorMsg(fmt.Errorf("failed to parse id: %w", err)))
		}
		jc, exist := f.GetJob(int(jid))
		if !exist {
			return c.JSON(404, errorMsg(fmt.Errorf("job#%d not found", jid)))
		}
		defer jc.Release()
		return c.JSON(200, jc.Job)
	}

	delJobById := func(c echo.Context) error {
		jid, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.JSON(400, errorMsg(fmt.Errorf("failed to parse job id: %w", err)))
		}
		err = f.StopJob(int(jid))
		if err != nil {
			return c.JSON(500, errorMsg(fmt.Errorf("failed to stop job#%d: %w", jid, err)))
		}
		return c.NoContent(204)
	}

	submitJob := func(c echo.Context) error {
		newJob := new(fuzzTypes.Fuzz)
		err := c.Bind(newJob)
		if err != nil {
			return c.JSON(400, errorMsg(fmt.Errorf("failed to unmarshal job: %w", err)))
		}
		jid, err := f.Submit(newJob)
		if err != nil {
			return c.JSON(500, errorMsg(fmt.Errorf("failed to submit job: %w", err)))
		}
		return c.JSON(200, map[string]int{"jid": jid})
	}

	getJids := func(c echo.Context) error {
		return c.JSON(200, f.GetJobIds())
	}

	// restful api
	g := e.Group("/", auth)
	g.GET("job/:id", getJobById)
	g.DELETE("job/:id", delJobById)
	g.POST("job", submitJob)
	g.GET("jobIds", getJids)

	f.s = &httpService{
		e:           e,
		addr:        addr,
		tls:         startTLS,
		accessToken: acToken,
	}

	go func() {
		f.s.wg.Add(1)
		defer f.s.wg.Done()
		if startTLS {
			e.Logger.Fatal(e.StartTLS(addr, certFileName, certKeyName))
		} else {
			e.Logger.Fatal(e.Start(addr))
		}
	}()
	return nil
}
