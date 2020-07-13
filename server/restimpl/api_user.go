/*
 * Simple blogging API handlers
 */

package restimpl

import (
	"fmt"
	"github.com/gin-gonic/gin"
	restimpl "github.com/gouthams/blogApp/server/model"
	"github.com/gouthams/blogApp/server/utils"
	"github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"mime"
	"net/http"
	"strconv"
	"time"
)

// AddBlogUsers - adds an blogUsers item
func AddBlogUsers(c *gin.Context) {
	logEntry := utils.Log().WithFields(utils.Fields{"url": c.Request.URL,
		"Method": c.Request.Method})
	logEntry.Debug("Post request received.")

	contentType := c.Request.Header.Get("Content-type")
	if contentType, _, err := mime.ParseMediaType(contentType); contentType != "application/json" || err != nil {
		logEntry.Errorf("Unsupported content type : %s", contentType)
		c.JSON(http.StatusUnsupportedMediaType, restimpl.Error{Code: "415", Message: contentType})
		return
	}

	var blogUser restimpl.BlogUser
	err := c.BindJSON(&blogUser)
	if err != nil {
		logEntry.Errorf("Json parsing error %v", err)
		c.JSON(http.StatusBadRequest, restimpl.Error{Code: "400", Message: err.Error()})
		return
	}

	isDup, err := getBlogUserByEmail(blogUser.Email, logEntry)
	if isDup != (restimpl.BlogUser{}) {
		logEntry.Errorf("User already exists %v", err)
		c.JSON(http.StatusConflict, restimpl.Error{Code: "409", Message: "Email address is not unique"})
		return
	}

	//Set the readonly fields
	//Set the time in UTC
	blogUser.LastModifiedDate = time.Now().UTC()
	blogUser.Id = uuid.NewV4().String()

	blogCollection, ctx := utils.GetUserCollection()
	doc, err := blogCollection.InsertOne(ctx, blogUser)
	if err != nil {
		logEntry.Errorf("Insert failed %v", err)
	}

	logEntry.Debugf("Document created %v doc: %v", doc.InsertedID)

	user, err := getBlogUserByid(blogUser.Id, logEntry)
	if err != nil {
		logEntry.Errorf("Retrieval failed!")
		c.JSON(http.StatusInternalServerError, restimpl.Error{Code: "500", Message: err.Error()})
		return
	}

	logEntry.Infof("blogUser with id: %s created!", blogUser.Id)
	c.JSON(http.StatusCreated, user)
	return
}

// GetblogUsers - get a single blogUsers
func GetblogUsers(c *gin.Context) {
	logEntry := utils.Log().WithFields(utils.Fields{"url": c.Request.URL,
		"Method": c.Request.Method})
	logEntry.Debug("Get request received.")

	id := c.Param("id")
	if id, err := uuid.FromString(id); err != nil {
		logEntry.Errorf("Invalid UUID: %s", id.String())
		c.JSON(http.StatusBadRequest, restimpl.Error{Code: "400", Message: err.Error()})
		return
	}

	user, err := getBlogUserByid(id, logEntry)
	if err != nil {
		logEntry.Errorf("Retrieval failed!")
		c.JSON(http.StatusNotFound, restimpl.Error{Code: "404", Message: err.Error()})
		return
	}

	logEntry.Infof("BlogUser with id: %s retrieved ", id)
	c.JSON(http.StatusOK, user)
	return
}

// Helper method to get user based on the id
func getBlogUserByid(id string, logEntry *utils.REntry) (restimpl.BlogUser, error) {
	//Filter with the parameter id from the url
	filter := bson.D{{"id", id}}

	var user restimpl.BlogUser
	blogCollection, ctx := utils.GetUserCollection()
	err := blogCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		logEntry.Errorf("Search failed %v", err)
		return restimpl.BlogUser{}, err
	}
	return user, nil
}

// Helper method to get user based on the email
func getBlogUserByEmail(email string, logEntry *utils.REntry) (restimpl.BlogUser, error) {
	//Filter with the parameter email from the url
	filter := bson.D{{"email", email}}

	var user restimpl.BlogUser
	blogCollection, ctx := utils.GetUserCollection()
	err := blogCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		logEntry.Errorf("Search failed %v", err)
		return restimpl.BlogUser{}, err
	}
	return user, nil
}

// SearchblogUsers - searches blogUsers
func SearchblogUsers(c *gin.Context) {
	logEntry := utils.Log().WithFields(utils.Fields{"url": c.Request.URL,
		"Method": c.Request.Method})
	logEntry.Info("Search request received.")

	//Query string from the url
	query := c.Request.URL.Query()
	var filter bson.D

	name := query.Get("name")
	if name == "" {
		logEntry.Errorf("Invalid name to search search: %s. Ignores this filter", name)
		//Empty filter to get all the records of user in the slice
		filter = bson.D{}
	} else {
		filter = bson.D{{"name", name}}
	}
	logEntry.Debugf("Filter criteria %v", filter)

	findOptions := options.Find()
	pageSize := query.Get("pageSize")
	pageLimit, err := strconv.ParseInt(pageSize, 10, 64)
	if err != nil {
		logEntry.Errorf("Invalid pageSize: %d. Ignores this filter", pageSize)
	}
	findOptions.SetLimit(pageLimit)

	//Explicitly initialize the slice with empty value to return if none found
	res := []restimpl.BlogUser{}
	blogCollection, ctx := utils.GetUserCollection()
	cursor, err := blogCollection.Find(ctx, filter, findOptions)
	if err != nil {
		logEntry.Errorf("Search failed %v", err)
		c.JSON(http.StatusNotFound, restimpl.Error{Code: "404", Message: err.Error()})
		return
	}

	for cursor.Next(ctx) {
		logEntry.Debugf("Documents retrieved %v", cursor.Current)
		var user restimpl.BlogUser
		err := cursor.Decode(&user)
		//If the is issue with one user log the error and continue
		if err != nil {
			logEntry.Errorf("Unable to decode user: %v", err)

		}
		res = append(res, user)
	}

	logEntry.Info("BlogUser document search done!")
	c.JSON(http.StatusOK, res)
	return
}

// UpdateBlogUsers - update an blogUsers item
func UpdateBlogUsers(c *gin.Context) {
	logEntry := utils.Log().WithFields(utils.Fields{"url": c.Request.URL,
		"Method": c.Request.Method})
	logEntry.Debug("Update request received.")

	contentType := c.Request.Header.Get("Content-type")
	if contentType, _, err := mime.ParseMediaType(contentType); contentType != "application/json" || err != nil {
		logEntry.Errorf("Unsupported content type : %s", contentType)
		c.JSON(http.StatusUnsupportedMediaType, restimpl.Error{Code: "415", Message: contentType})
		return
	}

	id := c.Param("id")
	if id, err := uuid.FromString(id); err != nil {
		logEntry.Errorf("Invalid UUID: %s", id.String())
		c.JSON(http.StatusBadRequest, restimpl.Error{Code: "400", Message: err.Error()})
		return
	}

	var blogUser restimpl.BlogUser
	err := c.BindJSON(&blogUser)
	if err != nil {
		logEntry.Errorf("Json parsing error %v", err)
		c.JSON(http.StatusBadRequest, restimpl.Error{Code: "400", Message: err.Error()})
		return
	}

	//update the time in UTC
	blogUser.LastModifiedDate = time.Now().UTC()
	blogUser.Id = id

	blogCollection, ctx := utils.GetUserCollection()
	isDone, err := deleteUserById(id, logEntry)
	if isDone == false {
		logEntry.Errorf("Delete user failed")
		c.JSON(http.StatusInternalServerError, restimpl.Error{Code: "500",
			Message: fmt.Sprintf("Delete user with id: %s failed", id)})
	}

	doc, err := blogCollection.InsertOne(ctx, blogUser)
	if err != nil {
		logEntry.Errorf("Insert failed %v", err)
	}

	logEntry.Infof("Document created %v doc: %v", doc.InsertedID, doc)

	user, err := getBlogUserByid(blogUser.Id, logEntry)
	if err != nil {
		logEntry.Errorf("Retrieval failed!")
		c.JSON(http.StatusInternalServerError, restimpl.Error{Code: "500", Message: err.Error()})
		return
	}

	logEntry.Infof("blogUser with id: %s updated!", blogUser.Id)
	c.JSON(http.StatusOK, user)
	return
}

//Helper function to delete user by the given Id
func deleteUserById(id string, logEntry *utils.REntry) (bool, error) {
	//Delete the blogUser
	deleteFilter := bson.D{{"id", id}}

	blogCollection, ctx := utils.GetUserCollection()
	//Check to see if user exist
	_, err := getBlogUserByid(id, logEntry)
	if err != nil {
		// If get user fails that means the user is not in the system, delete will be treated as success.
		// TODO: If the failure is not able to connect to DB this behaviour needs to be changed.
		logEntry.Errorf("Unable to get the user with id: %s. Error : %v", id, err)
		return true, nil
	}

	deletedUser, err := blogCollection.DeleteOne(ctx, deleteFilter)
	if err != nil {
		logEntry.Errorf("Delete failed %v", err)
	}

	if deletedUser.DeletedCount != 1 {
		logEntry.Errorf("Delete count is not 1 %d", deletedUser.DeletedCount)
		return false, err
	}
	return true, nil
}

// DeleteBlogUsers - deletes an blogUsers item
func DeleteBlogUsers(c *gin.Context) {
	logEntry := utils.Log().WithFields(utils.Fields{"url": c.Request.URL,
		"Method": c.Request.Method})
	logEntry.Debug("Delete request received.")

	contentType := c.Request.Header.Get("Content-type")
	if contentType, _, err := mime.ParseMediaType(contentType); contentType != "application/json" || err != nil {
		logEntry.Errorf("Unsupported content type : %s", contentType)
		c.JSON(http.StatusUnsupportedMediaType, restimpl.Error{Code: "415", Message: contentType})
		return
	}

	id := c.Param("id")
	if id, err := uuid.FromString(id); err != nil {
		logEntry.Errorf("Invalid UUID: %s", id.String())
		c.JSON(http.StatusBadRequest, restimpl.Error{Code: "400", Message: err.Error()})
		return
	}

	isDone, _ := deleteUserById(id, logEntry)
	if isDone == false {
		logEntry.Errorf("Delete user failed")
		c.JSON(http.StatusInternalServerError, restimpl.Error{Code: "500",
			Message: fmt.Sprintf("Delete user with id: %s failed", id)})
	}

	logEntry.Infof("blogUser with id: %s deleted!", id)
	c.JSON(http.StatusNoContent, restimpl.Error{Code: "204",
		Message: fmt.Sprintf("Delete user with id: %s Succeeded",
			id)})
}
