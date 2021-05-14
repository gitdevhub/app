package comment

func DeleteComment(c *gin.Context) {
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

	v, err := model.GetCommentByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	if v.UserId != userId {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Access error"})
		return
	}

	_, err = v.DeleteComment()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error. Try again"})
		return
	}

	c.Status(http.StatusOK)
}
