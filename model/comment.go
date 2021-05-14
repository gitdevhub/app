package model

import (
	"app/db"
	"database/sql"
	"gopkg.in/guregu/null.v4"
	"math"
	"time"
)

type Comment struct {
	
}


func PageCountAllRecursiveCommentsById(id int64, limit int) (int, error) {
	var count int64

	err := db.GetDB().
		Raw(`WITH RECURSIVE cte (id, message, parent_id, created_at) AS
			(
				SELECT id, message, parent_id, created_at
					FROM comments
					WHERE parent_id = ?
				UNION ALL
				SELECT c.id, c.message, c.parent_id, c.created_at
					FROM cte
				INNER JOIN comments as c ON cte.id = c.parent_id

				LIMIT ?
			)
			SELECT COUNT(*) FROM cte`, id, limit).
		Scan(&count).
		Error
	if err != nil {
		return 0, err
	}

	pageCount := math.Ceil(float64(count) / float64(limit))

	return int(pageCount), nil
}

func GetAllRecursiveCommentsByParentId(parentId int64, userId int64, page int, limit int) ([]CommentResult, error) {
	var list []CommentResult
	offset := (page - 1) * limit

	rows, err := db.GetDB().
		Raw(`WITH RECURSIVE cte (id, parent_id, user_id, message, status, created_at) AS
			(
				SELECT id, parent_id, user_id, message, status, created_at
					FROM comments
					WHERE parent_id = @parent_id
				UNION ALL
				SELECT c.id, c.parent_id, c.user_id, c.message, c.status, c.created_at
					FROM cte
				INNER JOIN comments as c ON cte.id = c.parent_id
				
				LIMIT @limit
				OFFSET @offset
			)
			SELECT
			cte.id,
			cte.parent_id,
			cte.message,
			cte.status,
			cte.created_at,
			(
				SELECT COUNT(*)
				FROM comment_likes
				WHERE comment_likes.comment_id = cte.id 
					AND comment_likes.status = @comment_like_status
			) likes,
			(
				SELECT COUNT(*) > 0
				FROM comment_likes
				WHERE comment_likes.comment_id = cte.id 
					AND comment_likes.status = @comment_like_status 
					AND comment_likes.user_id = @user_id
			) liked,
			u.id as user_id,
			u.name,
			u.username,
			u.image
			FROM cte
			INNER JOIN users u ON u.id = cte.user_id
			ORDER BY created_at`,
			sql.Named("user_id", userId),
			sql.Named("parent_id", parentId),
			sql.Named("comment_like_status", CommentLikeActive),
			sql.Named("limit", limit),
			sql.Named("offset", offset),
		).
		Rows()

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var result CommentResult
		err := rows.Scan(
			&result.ID,
			&result.ParentId,
			&result.Message,
			&result.Status,
			&result.CreatedAt,
			&result.Likes,
			&result.Liked,
			&result.User.ID,
			&result.User.Name,
			&result.User.Username,
			&result.User.Image,
		)
		if err != nil {
			return nil, err
		}

		list = append(list, result)
	}

	return list, nil
}
