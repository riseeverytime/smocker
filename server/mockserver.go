package server

import (
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/Thiht/smocker/templates"
	"github.com/Thiht/smocker/types"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type MockServer interface {
	AddMock(*types.Mock)
	Mocks() types.Mocks
	Mock(id string) *types.Mock
	History(filterPath string) (types.History, error)
	Sessions() types.Sessions
	Reset()
	Clear()
}

type mockServer struct {
	server   *echo.Echo
	mocks    types.Mocks
	sessions types.Sessions
	mu       sync.Mutex
}

func NewMockServer(port int) MockServer {
	s := &mockServer{
		server:   echo.New(),
		mocks:    types.Mocks{},
		sessions: types.Sessions{{}},
	}

	s.server.HideBanner = true
	s.server.HidePort = true
	s.server.Use(recoverMiddleware(), loggerMiddleware(), s.historyMiddleware())
	s.server.Any("/*", s.genericHandler)

	log.WithField("port", port).Info("Starting mock server")
	go func() {
		if err := s.server.Start(":" + strconv.Itoa(port)); err != nil {
			log.WithError(err).Error("Mock Server execution failed")
		}
	}()

	return s
}

func (s *mockServer) genericHandler(c echo.Context) error {
	s.mu.Lock()
	mocks := make(types.Mocks, len(s.mocks))
	copy(mocks, s.mocks)
	s.mu.Unlock()

	actualRequest := types.HTTPRequestToRequest(c.Request())
	b, _ := yaml.Marshal(actualRequest)
	log.Debugf("Received request:\n---\n%s\n", string(b))

	/* Request matching */

	var (
		matchingMock *types.Mock
		response     *types.MockResponse
		err          error
	)
	exceededMocks := types.Mocks{}
	for _, mock := range mocks {
		if mock.Request.Match(actualRequest) {
			matchingMock = mock
			if matchingMock.Context.Times > 0 && matchingMock.State.TimesCount >= matchingMock.Context.Times {
				b, _ = yaml.Marshal(mock)
				log.Tracef("Times exceeded, skipping mock:\n---\n%s\n", string(b))
				exceededMocks = append(exceededMocks, mock)
				continue
			}
			b, _ = yaml.Marshal(matchingMock)
			log.Debugf("Matching mock:\n---\n%s\n", string(b))
			if mock.DynamicResponse != nil {
				response, err = templates.GenerateMockResponse(mock.DynamicResponse, actualRequest)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, echo.Map{
						"message": err.Error(),
						"request": actualRequest,
					})
				}
			} else if mock.Proxy != nil {
				response, err = mock.Proxy.Redirect(actualRequest)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, echo.Map{
						"message": err.Error(),
						"request": actualRequest,
					})
				}
			} else if mock.Response != nil {
				response = mock.Response
			}
			matchingMock.State.TimesCount++
			c.Set(types.MockIDKey, matchingMock.State.ID)
			break
		} else {
			b, _ = yaml.Marshal(mock)
			log.Tracef("Skipping mock:\n---\n%s\n", string(b))
		}
	}
	if response == nil {
		resp := echo.Map{
			"message": "No mock found matching the request",
			"request": actualRequest,
		}
		if len(exceededMocks) > 0 {
			for _, mock := range exceededMocks {
				mock.State.TimesCount++
			}
			resp["message"] = "Matching mock found but was exceeded"
			resp["nearest"] = exceededMocks
		}
		b, _ = yaml.Marshal(resp)
		log.Debugf("No mock found, returning:\n---\n%s\n", string(b))
		return c.JSON(666, resp)
	}

	/* Response writing */

	// Headers
	for key, values := range response.Headers {
		for _, value := range values {
			c.Response().Header().Add(key, value)
		}
	}

	// Delay
	time.Sleep(response.Delay)

	// Status
	if response.Status == 0 {
		// Fallback to 200 OK
		response.Status = http.StatusOK
	}
	c.Response().WriteHeader(response.Status)

	// Body
	if _, err = c.Response().Write([]byte(response.Body)); err != nil {
		log.WithError(err).Error("Failed to write response body")
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	b, _ = yaml.Marshal(response)
	log.Debugf("Returned response:\n---\n%s\n", string(b))
	return nil
}

func (s *mockServer) AddMock(newMock *types.Mock) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mocks = append(types.Mocks{newMock}, s.mocks...)
}

func (s *mockServer) Mocks() types.Mocks {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mocks
}

func (s *mockServer) Mock(id string) *types.Mock {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, mock := range s.mocks {
		if mock.State.ID == id {
			return mock
		}
	}
	return nil
}

func (s *mockServer) History(filterPath string) (types.History, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	res := types.History{}
	regex, err := regexp.Compile(filterPath)
	if err != nil {
		return res, err
	}
	history := s.sessions[len(s.sessions)-1]
	for _, entry := range history {
		if regex.Match([]byte(entry.Request.Path)) {
			res = append(res, entry)
		}
	}
	return res, nil
}

func (s *mockServer) Sessions() types.Sessions {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sessions
}

func (s *mockServer) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mocks = types.Mocks{}
	s.sessions = append(s.sessions, types.History{})
}

func (s *mockServer) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mocks = types.Mocks{}
	s.sessions = types.Sessions{}
}
