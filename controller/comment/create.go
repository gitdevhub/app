package comment

type CreateCommentForm struct {
	PostId      int64  `form:"post_id" json:"post_id" binding:"required,gt=0"`
	ParentId    int64  `form:"parent_id" json:"parent_id" binding:"omitempty,gt=0"`
	Message     string `form:"message" json:"message" binding:"required,max=1024"`
}

func CreateComment(c *gin.Context) {
	userId, err := util.GetAuth(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You need to be authorized to access this route"})
		return
	}

	var form CreateCommentForm

	if err := c.ShouldBindJSON(&form); err != nil {
		logger.Error("the error form validation: ", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid json provided"})
		return
	}

	var rootId *int64
	var parentId *int64

	if form.ParentId > 0 {
		comment, err := model.GetCommentByIDAndPostId(form.ParentId, form.PostId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
			return
		}

		parentId = &form.ParentId
		if comment.RootId.Int64 > 0 {
			rootId = &comment.RootId.Int64
		} else {
			rootId = &comment.ID
		}
	}

	v := model.Comment{
		UserId:      userId,
		RootId:      null.IntFromPtr(rootId),
		ParentId:    null.IntFromPtr(parentId),
		PostId:      form.PostId,
		Message:     html.EscapeString(strings.TrimSpace(form.Message)),
		Status:      model.CommentActive,
		CreatedAt:   time.Now(),
	}

	_, err = model.CreateComment(&v)
	if err != nil {
		logger.Error("the error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error. Try again"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             v.ID,
		"root_id":        v.RootId,
		"parent_id":      v.ParentId,
		"user_id":        v.UserId,
		"message":        v.Message,
		"status":         v.Status,
		"created_at":     v.CreatedAt,
	})
}
