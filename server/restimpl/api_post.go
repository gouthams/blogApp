package restimpl

import (
	"fmt"
	restimpl "github.com/gouthams/blogApp/server/model"
	"github.com/gouthams/blogApp/server/utils"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AddblogPosts - adds an blogPosts item
func AddblogPosts(c *gin.Context) {
	logEntry := utils.Log().WithFields(utils.Fields{"url": c.Request.URL,
		"Method": c.Request.Method})
	logEntry.Debug("Post request received.")

	contentType := c.Request.Header.Get("Content-type")
	if contentType, _, err := mime.ParseMediaType(contentType); contentType != "application/json" || err != nil {
		logEntry.Errorf("Unsupported content type : %s", contentType)
		c.JSON(http.StatusUnsupportedMediaType, restimpl.Error{Code: "415", Message: contentType})
		return
	}

	var blogPost restimpl.BlogPost
	err := c.BindJSON(&blogPost)
	if err != nil {
		logEntry.Errorf("Json parsing error %v", err)
		c.JSON(http.StatusBadRequest, restimpl.Error{Code: "400", Message: err.Error()})
		return
	}

	_, err = getBlogUserByid(blogPost.UserId, logEntry)
	if err != nil {
		logEntry.Errorf("Retrieval failed!")
		c.JSON(http.StatusBadRequest, restimpl.Error{Code: "400", Message: "UserId is not valid."})
		return
	}

	//Set the readonly fields
	//Set the time in UTC
	blogPost.LastModifiedDate = time.Now().UTC()
	blogPost.Id = uuid.NewV4().String()

	blogCollection, ctx := utils.GetPostCollection()
	doc, err := blogCollection.InsertOne(ctx, blogPost)
	if err != nil {
		logEntry.Errorf("Insert failed %v", err)
	}

	logEntry.Debugf("Document created %v doc: %v", doc.InsertedID)

	post, err := getBlogPostByid(blogPost.Id, logEntry)
	if err != nil {
		logEntry.Errorf("Retrieval failed!")
		c.JSON(http.StatusInternalServerError, restimpl.Error{Code: "500", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, post)
	return
}

// DeleteBlogPosts - deletes an blogPosts item
func DeleteBlogPosts(c *gin.Context) {
	logEntry := utils.Log().WithField("url", c.Request.URL)
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

	isDone, _ := deletePostById(id, logEntry)
	if isDone == false {
		logEntry.Errorf("Delete post failed")
		c.JSON(http.StatusInternalServerError, restimpl.Error{Code: "500",
			Message: fmt.Sprintf("Delete post with id: %s failed", id)})
	}

	c.JSON(http.StatusNoContent, restimpl.Error{Code: "500",
		Message: fmt.Sprintf("Delete post with id: %s Succeeded",
			id)})
}

//Helper function to delete post by the given Id
func deletePostById(id string, logEntry *utils.REntry) (bool, error) {
	//Delete the blogPost
	deleteFilter := bson.D{{"id", id}}

	postCollection, ctx := utils.GetPostCollection()
	//Check to see if post exist
	_, err := getBlogPostByid(id, logEntry)
	if err != nil {
		// If get post fails that means the post is not in the system, delete will be treated as success.
		// TODO: If the failure is not able to connect to DB this behaviour needs to be changed.
		logEntry.Errorf("Unable to get the post with id: %s", id)
		return true, nil
	}

	deletedPost, err := postCollection.DeleteOne(ctx, deleteFilter)
	if err != nil {
		logEntry.Errorf("Delete failed %v", err)
	}

	if deletedPost.DeletedCount != 1 {
		logEntry.Errorf("Delete count is not 1 %d", deletedPost.DeletedCount)
		return false, err
	}
	return true, nil
}

// GetblogPosts - get a single blogPosts
func GetblogPosts(c *gin.Context) {
	logEntry := utils.Log().WithFields(utils.Fields{"url": c.Request.URL,
		"Method": c.Request.Method})
	logEntry.Debug("Get request received.")

	id := c.Param("id")
	if id, err := uuid.FromString(id); err != nil {
		logEntry.Errorf("Invalid UUID: %s", id.String())
		c.JSON(http.StatusBadRequest, restimpl.Error{Code: "400", Message: err.Error()})
		return
	}

	post, err := getBlogPostByid(id, logEntry)
	if err != nil {
		logEntry.Errorf("Retrieval failed!")
		c.JSON(http.StatusNotFound, restimpl.Error{Code: "404", Message: err.Error()})
		return
	}

	logEntry.Infof("Documents retrieved %v")
	c.JSON(http.StatusOK, post)
	return
}

// Helper method to get psot based on the id
func getBlogPostByid(id string, logEntry *utils.REntry) (restimpl.BlogPost, error) {
	//Filter with the parameter id from the url
	filter := bson.D{{"id", id}}

	var post restimpl.BlogPost
	blogCollection, ctx := utils.GetPostCollection()
	err := blogCollection.FindOne(ctx, filter).Decode(&post)
	if err != nil {
		logEntry.Errorf("Search failed %v", err)
		return restimpl.BlogPost{}, err
	}
	return post, nil
}

// SearchblogPosts - searches blogPosts
func SearchblogPosts(c *gin.Context) {
	logEntry := utils.Log().WithFields(utils.Fields{"url": c.Request.URL,
		"Method": c.Request.Method})
	logEntry.Info("Search request received.")

	//Query string from the url
	query := c.Request.URL.Query()
	var filter bson.D

	userId := query.Get("userId")
	if userId, err := uuid.FromString(userId); err != nil {
		logEntry.Errorf("Invalid userId to search: %s. Ignores this filter", userId)
		//Empty filter to get all the records of post in the slice
		filter = bson.D{}
	} else {
		filter = bson.D{{"userid", userId.String()}}
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
	var res []restimpl.BlogPost
	postCollection, ctx := utils.GetPostCollection()
	cursor, err := postCollection.Find(ctx, filter, findOptions)
	if err != nil {
		logEntry.Errorf("Search failed %v", err)
		c.JSON(http.StatusNotFound, restimpl.Error{Code: "404", Message: err.Error()})
		return
	}

	logEntry.Debugf("Cursor from DB: %v", cursor)
	for cursor.Next(ctx) {
		logEntry.Infof("Documents retrieved %v", cursor.Current)
		var post restimpl.BlogPost
		err := cursor.Decode(&post)
		//If the is issue with one post log the error and continue
		if err != nil {
			logEntry.Errorf("Unable to decode post: %v", err)

		}
		res = append(res, post)
	}

	c.JSON(http.StatusOK, res)
	return
}

// UpdateblogPosts - update an blogPosts item
func UpdateblogPosts(c *gin.Context) {
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

	var blogPost restimpl.BlogPost
	err := c.BindJSON(&blogPost)
	if err != nil {
		logEntry.Errorf("Json parsing error %v", err)
		c.JSON(http.StatusBadRequest, restimpl.Error{Code: "400", Message: err.Error()})
		return
	}

	//update the time in UTC
	blogPost.LastModifiedDate = time.Now().UTC()
	blogPost.Id = id

	blogCollection, ctx := utils.GetPostCollection()
	isDone, err := deletePostById(id, logEntry)
	if isDone == false {
		logEntry.Errorf("Delete post failed")
		c.JSON(http.StatusInternalServerError, restimpl.Error{Code: "500",
			Message: fmt.Sprintf("Delete post with id: %s failed", id)})
	}

	doc, err := blogCollection.InsertOne(ctx, blogPost)
	if err != nil {
		logEntry.Errorf("Insert failed %v", err)
	}

	logEntry.Infof("Document created %v doc: %v", doc.InsertedID, doc)

	post, err := getBlogPostByid(blogPost.Id, logEntry)
	if err != nil {
		logEntry.Errorf("Retrieval failed!")
		c.JSON(http.StatusInternalServerError, restimpl.Error{Code: "500", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, post)
	return
}
