package libfgiu

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"strconv"
	"sync"
)

var errUnauth = errors.New("unauthorized")

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

func (f *Fuzzer) startHttpApi(apiConf *WebApiConfig) error {
	addr, startTLS, certFileName, certKeyName := getApiConfig(apiConf)

	acToken := common.RandMarker()
	e := echo.New()

	getJobById := func(c echo.Context) error {
		tokenHeader := c.Request().Header.Get("Access-Token")
		if tokenHeader != acToken {
			return c.NoContent(401)
		}
		jid, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse id: %w", err)
		}
		jc, exist := f.GetJob(int(jid))
		if jc != nil {
			defer jc.Release()
		}
		if !exist {
			return fmt.Errorf("job#%d does not exist", jid)
		}
		return c.JSON(200, jc.Job)
	}

	delJobById := func(c echo.Context) error {
		tokenHeader := c.Request().Header.Get("Access-Token")
		if tokenHeader != acToken {
			return c.NoContent(401)
		}
		jid, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse id: %w", err)
		}
		return f.StopJob(int(jid))
	}

	submitJob := func(c echo.Context) error {
		tokenHeader := c.Request().Header.Get("Access-Token")
		if tokenHeader != acToken {
			return c.NoContent(401)
		}
		newJob := new(fuzzTypes.Fuzz)
		err := c.Bind(newJob)
		if err != nil {
			return err
		}
		return f.Submit(newJob)
	}

	getJids := func(c echo.Context) error {
		tokenHeader := c.Request().Header.Get("Access-Token")
		if tokenHeader != acToken {
			return c.NoContent(401)
		}
		return c.JSON(200, f.GetJobIds())
	}

	// restful api
	e.GET("/job/:id", getJobById)
	e.DELETE("/job/:id", delJobById)
	e.POST("/job", submitJob)
	e.GET("/jobIds", getJids)

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
