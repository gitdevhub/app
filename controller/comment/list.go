package comment

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/guregu/null.v4"
	"net/http"
	"strings"
)

type PaginationQueryForm struct {
	PostId int64  `form:"post_id" json:"post_id" binding:"required,gt=0"`
	Page        int    `form:"page,default=1" json:"page,default=1" binding:"required,gt=0"`
	Limit       int    `form:"limit,default=20" json:"limit,default=20" binding:"required,gte=10,lte=100"`
	SortBy      string `form:"sort_by,default=id" json:"sort_by,default=id" binding:"oneof=created_at id name"`
	Order       string `form:"order,default=asc" json:"order,default=asc" binding:"oneof=desc asc"`
}

func ListComments(c *gin.Context) {
	userId, err := util.GetAuth(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You need to be authorized to access this route"})
		return
	}

	var form PaginationQueryForm

	if err := c.ShouldBindQuery(&form); err != nil {
		logger.Error("the error form validation: ", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid query provided"})
		return
	}

	// validate page
	pageCount, err := model.PageCountAllCommentsByPostId(form.PostId, form.Limit)
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

	list, err := model.GetAllCommentsByPostId(form.PostId, userId, form.Page, form.Limit)
	if err != nil {
		logger.Error("the error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Comments not found"})
		return
	}

	if len(list) > 0 {
		for i, entity := range list {
			if strings.HasPrefix(entity.User.Image.String, "images/") {
				entity.User.Image = null.StringFrom(fmt.Sprintf("%s/%s", config.Server.GetFullHostName(), entity.User.Image.String))
				list[i] = entity
			}

			if entity.User.Username.IsZero() {
				entity.User.Username = null.StringFrom(fmt.Sprintf("user%d", entity.User.ID))
				list[i] = entity
			}
		}

		// Get second level of comments
		var commentIds []int64
		for _, result := range list {
			if result.ChildrenCount > 0 {
				//commentIds = append(commentIds, result.ID)
			}
		}

		if len(commentIds) > 0 {
			listChild, err := model.GetAllChildrenCommentsByRootIds(commentIds, userId, 10)
			if err != nil {
				logger.Error("the error: ", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Comments not found"})
				return
			}

			listMap := make(map[int64]model.CommentResult, len(list))
			for _, val := range listMap {
				listMap[val.ID] = val
			}

			for _, resultChild := range listChild {
				if result, ok := listMap[resultChild.RootId.Int64]; ok {
					if resultChild.User.Username.IsZero() {
						resultChild.User.Username = null.StringFrom(fmt.Sprintf("user%d", resultChild.User.ID))
					}
					result.ChildrenItems = append(result.ChildrenItems, resultChild)
					listMap[resultChild.RootId.Int64] = result
				}
			}

			for i, val := range list {
				if result, ok := listMap[val.ID]; ok {
					list[i] = result
				}
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
