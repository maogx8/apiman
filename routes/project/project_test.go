package project

import (
	"encoding/json"
	"errors"
	//"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/bitly/go-simplejson"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/will835559313/apiman/models"
	"github.com/will835559313/apiman/pkg/jwt"
	"github.com/will835559313/apiman/pkg/log"
	"github.com/will835559313/apiman/pkg/setting"
	"github.com/will835559313/apiman/routes/team"
	"github.com/will835559313/apiman/routes/user"
	"gopkg.in/go-playground/validator.v9"
)

var (
	projectJSON    = `{"name":"web","description":"web版","avatar_url":"http://www.famulei.com/"}`
	badProjectJSON = `{"name":"web9132481942934623842384283423842","description":"web版","avatar_url":"http://www.famulei.com/"}`

	newProjectJSON    = `{"name":"web_v2","description":"伐木累web版","avatar_url":"http://www.famulei.com/v2/"}`
	badNewProjectJSON = `{"name":"weoiruweru023947210942348230482031","description":"伐木累web版","avatar_url":"http://www.famulei.com/v2/"}`

	authJSON     = `{"name":"will", "password":"mgx123"}`
	access_token = ""
	projectID    = uint(0)

	userJSON = `{"name":"will","nickname":"毛广献","password":"mgx123","avatar_url":"http://ojz1mcltu.bkt.clouddn.com/animals-august2015.jpg"}`
	teamJSON = `{"name":"famulei","description":"team","creator":"will","avatar_url":"http://www.famulei.com/images/index_v4/slogan.png"}`
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func init() {
	// load config
	setting.Conf = "C:/Users/will8/go/src/github.com/will835559313/apiman/conf/app.conf"
	setting.CustomConf = "C:/Users/will8/go/src/github.com/will835559313/apiman/conf/custom.app.conf"
	setting.NewConfig()

	// set logger
	log.LoggerInit()

	// get db connection
	models.Dbinit()

	// migrate tables
	models.DB.DropTableIfExists(&models.Team{}, &models.User{}, &models.Project{}, &models.TeamUser{})
	models.DbMigrate()

	// set jwt
	jwt.JwtInint()
}

func createUser() {
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	req := httptest.NewRequest(echo.POST, "/users", strings.NewReader(userJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	user.CreateUser(c)
}

func createTeam() {
	// Setup
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	req := httptest.NewRequest(echo.POST, "/teams", strings.NewReader(teamJSON))
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+access_token)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	team.CreateTeam(c)
}

func TestGetToken(t *testing.T) {
	// create user
	createUser()

	// Setup
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	req := httptest.NewRequest(echo.POST, "/oauth2/token", strings.NewReader(authJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, user.GetToken(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), "access_token")
	}

	// save token
	js, err := simplejson.NewJson([]byte(rec.Body.String()))
	if err != nil {
		assert.Error(t, errors.New("save token error"))
	}
	access_token, _ = js.Get("access_token").String()
	//fmt.Println(access_token)
}

// normal request
func TestCreateProject(t *testing.T) {
	// create team
	createTeam()

	// Setup
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	req := httptest.NewRequest(echo.POST, "/teams/:teamname/projects", strings.NewReader(projectJSON))
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+access_token)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("teamname")
	c.SetParamValues("famulei")

	// Assertions
	if assert.NoError(t, CreateProject(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		t1 := new(models.Project)
		t2 := new(models.Project)
		json.Unmarshal([]byte(rec.Body.String()), t2)
		json.Unmarshal([]byte(projectJSON), t1)
		assert.Equal(t, t1.Name, t2.Name)
		assert.Equal(t, t1.Description, t2.Description)
		projectID = t2.ID
	}
}

// bad request
func TestBadCreateProject(t *testing.T) {
	// Setup
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	req := httptest.NewRequest(echo.POST, "/teams/:teamname/projects", strings.NewReader(badProjectJSON))
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+access_token)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("teamname")
	c.SetParamValues("famulei")

	// Assertions
	if assert.NoError(t, CreateProject(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

// normal request
func TestGetProject(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/projects/:id", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+access_token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	idstr := strconv.Itoa(int(projectID))
	c.SetParamValues(idstr)

	// Assertions
	if assert.NoError(t, GetProjectByID(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		t1 := new(models.Project)
		t2 := new(models.Project)
		json.Unmarshal([]byte(rec.Body.String()), t2)
		json.Unmarshal([]byte(projectJSON), t1)
		assert.Equal(t, t1.Name, t2.Name)
		assert.Equal(t, t1.Description, t2.Description)
	}
}

// bad request
func TestBadGetProject(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/projects/:id", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+access_token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	idstr := strconv.Itoa(int(projectID + 100))
	c.SetParamValues(idstr)

	// Assertions
	if assert.NoError(t, GetProjectByID(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}

// normal request
func TestUpdateProject(t *testing.T) {
	// Setup
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	req := httptest.NewRequest(echo.PUT, "/", strings.NewReader(newProjectJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+access_token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/projects/:id")
	idstr := strconv.Itoa(int(projectID))
	c.SetParamNames("id")
	c.SetParamValues(idstr)

	// Assertions
	if assert.NoError(t, UpdateProjectByID(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		t1 := new(models.Project)
		t2 := new(models.Project)
		json.Unmarshal([]byte(rec.Body.String()), t2)
		json.Unmarshal([]byte(newProjectJSON), t1)
		assert.Equal(t, t1.Name, t2.Name)
		assert.Equal(t, t1.Description, t2.Description)
	}
}

// bad request
func TestBadUpdateProject(t *testing.T) {
	// Setup
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	req := httptest.NewRequest(echo.PUT, "/", strings.NewReader(badNewProjectJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+access_token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/projects/:id")
	idstr := strconv.Itoa(int(projectID))
	c.SetParamNames("id")
	c.SetParamValues(idstr)

	// Assertions
	if assert.NoError(t, UpdateProjectByID(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

// normal request
func TestDeleteProject(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(echo.DELETE, "/projects/:id", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+access_token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	idstr := strconv.Itoa(int(projectID))
	c.SetParamNames("id")
	c.SetParamValues(idstr)

	// Assertions
	if assert.NoError(t, DeleteProjectByID(c)) {
		assert.Equal(t, http.StatusNoContent, rec.Code)
	}

	// delete team table
	// models.DB.DropTableIfExists(&models.Team{})
}

// bad request
func TestBadDeleteProject(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(echo.DELETE, "/projects/:id", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+access_token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	idstr := strconv.Itoa(int(projectID))
	c.SetParamNames("id")
	c.SetParamValues(idstr)

	// Assertions
	if assert.NoError(t, DeleteProjectByID(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}
