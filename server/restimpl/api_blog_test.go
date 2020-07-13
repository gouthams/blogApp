package restimpl

import (
	"bytes"
	"encoding/json"
	"fmt"
	restimpl "github.com/gouthams/blogApp/server/model"
	"github.com/gouthams/blogApp/server/utils"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const httpProtocol = "http"
const localhost = "0.0.0.0"

type RestImplTestSuite struct {
	suite.Suite
	MockPost restimpl.BlogPost
	MockUser restimpl.BlogUser
}

func TestRestImplTestSuite(t *testing.T) {
	testSuite := &RestImplTestSuite{}
	suite.Run(t, testSuite)
}

func (suite *RestImplTestSuite) SetupSuite() {
	suite.TearDownSuite()
	suite.MockPost = restimpl.BlogPost{UserId: "", Topic: "TestTopic", Content: "TestContent"}
	suite.MockUser = restimpl.BlogUser{Name: "David", Email: "david@abc.com"}
}

func (suite *RestImplTestSuite) AfterTest(_, _ string) {
	gock.Off()
	err := utils.FlushCollections()
	if err != nil {
		log.Printf("Flush Collection failed: %v", err)
	}
}

func (suite *RestImplTestSuite) TearDownSuite() {
	err := utils.FlushCollections()
	if err != nil {
		log.Printf("Flush Collection failed: %v", err)
	}
}

func getBlogPostUrl(resourceId string) string {
	url := fmt.Sprintf("%s://%s/blogPosts",
		httpProtocol,
		localhost)
	if resourceId != "" {
		url = url + "/" + resourceId
	}
	return url
}

func getBlogUserUrl(resourceId string) string {
	url := fmt.Sprintf("%s://%s/blogUsers",
		httpProtocol,
		localhost)
	if resourceId != "" {
		url = url + "/" + resourceId
	}
	return url
}

func getHostPath(t *testing.T, uri string) (string, string) {
	u, err := url.Parse(uri)

	assert.Equal(t, err, nil)

	startUri := u.Scheme + "://" + u.Host

	path := uri[len(startUri):]

	return startUri, path
}

func mockGet(host, path string, httpStatus int, mockBody interface{}) {
	gock.New(host).Get(path).Reply(httpStatus).JSON(mockBody)
}

func mockPost(host, path string, httpStatus int, mockBody interface{}) {
	gock.New(host).Post(path).Reply(httpStatus).JSON(mockBody)
}

func mockPut(host, path string, httpStatus int, mockBody interface{}) {
	gock.New(host).Put(path).Reply(httpStatus).JSON(mockBody)
}

func mockDelete(host, path string, httpStatus int, mockBody interface{}) {
	gock.New(host).Delete(path).Reply(httpStatus).JSON(mockBody)
}

func PerformRequest(router http.Handler, method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	logEntry := utils.Log()
	var buffer *bytes.Buffer
	if body != nil {
		bites, err := json.Marshal(body)
		if err != nil {
			logEntry.WithFields(utils.Fields{"error": err, "path": path,
				"body": body}).Error("FAILED TO PERFORM THE REQUEST")
		}
		buffer = bytes.NewBuffer(bites)
	} else {
		buffer = bytes.NewBuffer(make([]byte, 0))
	}
	req, err := http.NewRequest(method, path, buffer)
	if err != nil {
		logEntry.WithField("error", err).WithField("path", path).Error("FAILED TO PERFORM THE REQUEST")
	}

	for key, val := range headers {
		req.Header.Add(key, val)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

//Positive test cases
func (suite *RestImplTestSuite) TestCRUDBlogUsers() {

	_, path := getHostPath(suite.T(), getBlogUserUrl(""))
	router := NewRouter()
	header := map[string]string{"Content-Type": "application/json"}

	response := PerformRequest(router, http.MethodPost, path, suite.MockUser, header)
	assert.Equal(suite.T(), http.StatusCreated, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

	blogUserResp := restimpl.BlogUser{}
	err := json.Unmarshal(response.Body.Bytes(), &blogUserResp)
	if err != nil {
		log.Fatalf("Unmarshall Error %v", err)
	}
	assert.Equal(suite.T(), suite.MockUser.Name, blogUserResp.Name)

	response = PerformRequest(router, http.MethodGet, path, "", header)
	assert.Equal(suite.T(), http.StatusOK, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

	path = getBlogUserUrl(blogUserResp.Id)
	response = PerformRequest(router, http.MethodGet, path, "", header)
	assert.Equal(suite.T(), http.StatusOK, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

	suite.MockUser.Name = "Matt"
	suite.MockUser.Email = "matt@abc.com"
	response = PerformRequest(router, http.MethodPut, path, suite.MockUser, header)
	assert.Equal(suite.T(), http.StatusOK, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

	err = json.Unmarshal(response.Body.Bytes(), &blogUserResp)
	if err != nil {
		log.Fatalf("Unmarshall Error %v", err)
	}
	assert.Equal(suite.T(), "Matt", blogUserResp.Name)
	assert.Equal(suite.T(), "matt@abc.com", blogUserResp.Email)

	response = PerformRequest(router, http.MethodDelete, path, "", header)
	assert.Equal(suite.T(), http.StatusNoContent, response.Code)
}

func (suite *RestImplTestSuite) TestCRUDBlogPosts() {
	router := NewRouter()
	header := map[string]string{"Content-Type": "application/json"}

	//Create a blog User to get the userId
	userResponse := PerformRequest(router, http.MethodPost, getBlogUserUrl(""), suite.MockUser, header)
	assert.Equal(suite.T(), http.StatusCreated, userResponse.Code)
	assert.NotNil(suite.T(), userResponse.Result().Body)

	blogUserResp := restimpl.BlogUser{}
	err := json.Unmarshal(userResponse.Body.Bytes(), &blogUserResp)
	if err != nil {
		log.Fatalf("Unmarshall Error %v", err)
	}
	assert.Equal(suite.T(), suite.MockUser.Name, blogUserResp.Name)
	_, path := getHostPath(suite.T(), getBlogPostUrl(""))

	postBody := restimpl.BlogPost{UserId: blogUserResp.Id, Topic: "OriginalTopic", Content: "OriginalContent"}
	response := PerformRequest(router, http.MethodPost, path, postBody, header)
	assert.Equal(suite.T(), http.StatusCreated, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

	blogPostResp := restimpl.BlogPost{}
	err = json.Unmarshal(response.Body.Bytes(), &blogPostResp)
	if err != nil {
		log.Fatalf("Unmarshall Error %v", err)
	}
	assert.Equal(suite.T(), blogUserResp.Id, blogPostResp.UserId)

	response = PerformRequest(router, http.MethodGet, path, "", header)
	assert.Equal(suite.T(), http.StatusOK, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

	path = getBlogPostUrl(blogPostResp.Id)
	response = PerformRequest(router, http.MethodGet, path, "", header)
	assert.Equal(suite.T(), http.StatusOK, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

	postBody.Topic = "UpdatedTopic"
	postBody.Content = "UpdatedContent"
	response = PerformRequest(router, http.MethodPut, path, postBody, header)
	assert.Equal(suite.T(), http.StatusOK, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

	err = json.Unmarshal(response.Body.Bytes(), &blogPostResp)
	if err != nil {
		log.Fatalf("Unmarshall Error %v", err)
	}
	assert.Equal(suite.T(), "UpdatedTopic", blogPostResp.Topic)
	assert.Equal(suite.T(), "UpdatedContent", blogPostResp.Content)

	response = PerformRequest(router, http.MethodDelete, path, "", header)
	assert.Equal(suite.T(), http.StatusNoContent, response.Code)
}

//Negative test cases for blogUser
func (suite *RestImplTestSuite) TestInvalidBlogUsers() {
	_, path := getHostPath(suite.T(), getBlogUserUrl(""))
	router := NewRouter()
	header := map[string]string{"Content-Type": ""}

	response := PerformRequest(router, http.MethodPost, path, suite.MockUser, header)
	assert.Equal(suite.T(), http.StatusUnsupportedMediaType, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

	header = map[string]string{"Content-Type": "application/json"}
	blogUser := restimpl.BlogUser{Name: "Invalid"}
	response = PerformRequest(router, http.MethodPost, path, blogUser, header)
	assert.Equal(suite.T(), http.StatusBadRequest, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

}

func (suite *RestImplTestSuite) TestDuplicateBlogUsers() {
	_, path := getHostPath(suite.T(), getBlogUserUrl(""))
	router := NewRouter()
	header := map[string]string{"Content-Type": "application/json"}

	response := PerformRequest(router, http.MethodPost, path, suite.MockUser, header)
	assert.Equal(suite.T(), http.StatusCreated, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

	response = PerformRequest(router, http.MethodPost, path, suite.MockUser, header)
	assert.Equal(suite.T(), http.StatusConflict, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

}

func (suite *RestImplTestSuite) TestGetInvalidBlogUsers() {
	_, path := getHostPath(suite.T(), getBlogUserUrl("12345"))
	router := NewRouter()
	header := map[string]string{"Content-Type": "application/json"}

	response := PerformRequest(router, http.MethodGet, path, "", header)
	assert.Equal(suite.T(), http.StatusBadRequest, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

	//Random UUID fetch
	response = PerformRequest(router, http.MethodGet, getBlogUserUrl(uuid.NewV4().String()), "", header)
	assert.Equal(suite.T(), http.StatusNotFound, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)
}

func (suite *RestImplTestSuite) TestInvalidBlogUserUpdate() {
	_, path := getHostPath(suite.T(), getBlogUserUrl(uuid.NewV4().String()))
	router := NewRouter()
	header := map[string]string{"Content-Type": ""}

	response := PerformRequest(router, http.MethodPut, path, suite.MockUser, header)
	assert.Equal(suite.T(), http.StatusUnsupportedMediaType, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

	header = map[string]string{"Content-Type": "application/json"}
	blogUser := restimpl.BlogUser{Name: "Invalid"}
	response = PerformRequest(router, http.MethodPut, path, blogUser, header)
	assert.Equal(suite.T(), http.StatusBadRequest, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

	response = PerformRequest(router, http.MethodPut, getBlogUserUrl("snjsnd"), blogUser, header)
	assert.Equal(suite.T(), http.StatusBadRequest, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

}

func (suite *RestImplTestSuite) TestInvalidBlogUserDelete() {
	_, path := getHostPath(suite.T(), getBlogUserUrl(uuid.NewV4().String()))
	router := NewRouter()
	header := map[string]string{"Content-Type": ""}

	response := PerformRequest(router, http.MethodDelete, path, "", header)
	assert.Equal(suite.T(), http.StatusUnsupportedMediaType, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

	header = map[string]string{"Content-Type": "application/json"}
	response = PerformRequest(router, http.MethodDelete, getBlogUserUrl("snjsnd"), "", header)
	assert.Equal(suite.T(), http.StatusBadRequest, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

}

//Negative test cases for blogPost
func (suite *RestImplTestSuite) TestInvalidBlogPosts() {
	_, path := getHostPath(suite.T(), getBlogPostUrl(""))
	router := NewRouter()
	header := map[string]string{"Content-Type": ""}

	response := PerformRequest(router, http.MethodPost, path, suite.MockPost, header)
	assert.Equal(suite.T(), http.StatusUnsupportedMediaType, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

	header = map[string]string{"Content-Type": "application/json"}
	blogPost := restimpl.BlogPost{Topic: "Invalid"}
	response = PerformRequest(router, http.MethodPost, path, blogPost, header)
	assert.Equal(suite.T(), http.StatusBadRequest, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

}

func (suite *RestImplTestSuite) TestGetInvalidBlogPosts() {
	_, path := getHostPath(suite.T(), getBlogPostUrl("12345"))
	router := NewRouter()
	header := map[string]string{"Content-Type": "application/json"}

	response := PerformRequest(router, http.MethodGet, path, "", header)
	assert.Equal(suite.T(), http.StatusBadRequest, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

	//Random UUID fetch
	response = PerformRequest(router, http.MethodGet, getBlogPostUrl(uuid.NewV4().String()), "", header)
	assert.Equal(suite.T(), http.StatusNotFound, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)
}

func (suite *RestImplTestSuite) TestInvalidBlogPostUpdate() {
	_, path := getHostPath(suite.T(), getBlogPostUrl(uuid.NewV4().String()))
	router := NewRouter()
	header := map[string]string{"Content-Type": ""}

	response := PerformRequest(router, http.MethodPut, path, suite.MockPost, header)
	assert.Equal(suite.T(), http.StatusUnsupportedMediaType, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

	header = map[string]string{"Content-Type": "application/json"}
	BlogPost := restimpl.BlogPost{Topic: "Invalid"}
	response = PerformRequest(router, http.MethodPut, path, BlogPost, header)
	assert.Equal(suite.T(), http.StatusBadRequest, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

	response = PerformRequest(router, http.MethodPut, getBlogPostUrl("snjsnd"), BlogPost, header)
	assert.Equal(suite.T(), http.StatusBadRequest, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

}

func (suite *RestImplTestSuite) TestInvalidBlogPostDelete() {
	_, path := getHostPath(suite.T(), getBlogPostUrl(uuid.NewV4().String()))
	router := NewRouter()
	header := map[string]string{"Content-Type": ""}

	response := PerformRequest(router, http.MethodDelete, path, "", header)
	assert.Equal(suite.T(), http.StatusUnsupportedMediaType, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

	header = map[string]string{"Content-Type": "application/json"}
	response = PerformRequest(router, http.MethodDelete, getBlogPostUrl("snjsnd"), "", header)
	assert.Equal(suite.T(), http.StatusBadRequest, response.Code)
	assert.NotNil(suite.T(), response.Result().Body)

}
