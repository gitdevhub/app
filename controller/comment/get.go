package comment

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/guregu/null.v4"
	"net/http"
	"strconv"
	"strings"
)

type GetPaginationQueryForm struct {
	Page   int    `form:"page,default=1" json:"page,default=1" binding:"required,gt=0"`
	Limit  int    `form:"limit,default=5" json:"limit,default=5" binding:"required,gte=5,lte=100"`
	SortBy string `form:"sort_by,default=id" json:"sort_by,default=id" binding:"oneof=created_at id name"`
	Order  string `form:"order,default=asc" json:"order,default=asc" binding:"oneof=desc asc"`
}

func GetComments(c *gin.Context) {
	userId, err := util.GetAuth(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You need to be authorized to access this route"})
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request. Try again"})
		return
	}

	var form GetPaginationQueryForm

	if err := c.ShouldBindQuery(&form); err != nil {
		logger.Error("the error form validation: ", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid query provided"})
		return
	}

	// validate page
	pageCount, err := model.PageCountAllChildrenCommentsByRootId(id, form.Limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	if pageCount <= 0 {
		response := map[string]interface{}{
			"list": []interface{}{},
			"pagination": map[string]interface{}{
				"totalPages": 0,
			},
		}

		c.JSON(http.StatusOK, response)
		return
	}

	if form.Page > pageCount {
		logger.Error("the error: form.Page > pageCount")
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("page is out of range. page range is 1-%v", pageCount)})
		return
	}

	list, err := model.GetAllChildrenCommentsByRootId(id, userId, form.Page, form.Limit)
	if err != nil {
		logger.Error("the error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Comments not found"})
		return
	}

	if len(list) > 0 {
		for i, entity := range list {
			if strings.HasPrefix(entity.User.Image.String, "images/") {
				list[i].User.Image = null.StringFrom(fmt.Sprintf("%s/%s", config.Server.GetFullHostName(), entity.User.Image.String))
			}

			if entity.User.Username.IsZero() {
				list[i].User.Username = null.StringFrom(fmt.Sprintf("user%d", entity.User.ID))
			}
		}
	}

	response := map[string]interface{}{
		"list": list,
		"pagination": map[string]interface{}{
			"totalPages": pageCount,
		},
	}

	c.JSON(http.StatusOK, response)
}
