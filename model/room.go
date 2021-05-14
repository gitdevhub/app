package model


type Room struct {
	ID        int64       `json:"id" gorm:"primaryKey;autoIncrement:true;"`
	RoomType  int64       `json:"room_type" gorm:"not null;"` // 0 - private or 1 - group
	Name      null.String `json:"name" gorm:"size:50;null;"`
	Image     null.String `json:"image" gorm:"size:255;null;"`
	CreatedAt time.Time   `json:"created_at" gorm:"type:DATETIME;NOT NULL;"`
}

func (Room) TableName() string {
	return "rooms"
}

func GetAllRooms(senderId int64, senderType int64, page int, limit int) ([]RoomInfoResult, error) {
	var list []RoomInfoResult
	offset := (page - 1) * limit

	rows, err := db.GetDB().
		Raw(
			`SELECT
			r2.user_id,
			r2.user_type,
			m.user_id,
			m.user_type,
			r2.room_id,
			r.room_type,
			m.id,
			m.message,
			m.created_at,
			CASE
				WHEN r.room_type = @room_type_private
					THEN
						CASE WHEN r2.user_type = @user_type_user THEN u2.name ELSE c2.name END
				ELSE r.name
			END as name,
			CASE
				WHEN r.room_type = @room_type_private
					THEN
						CASE WHEN r2.user_type = @user_type_user THEN u2.image ELSE c2.image END
				ELSE r.image
			END as image
			FROM (
				WITH ranked_rooms AS (
					SELECT
					ur1.user_id, ur1.user_type, ur1.room_id,
					ROW_NUMBER() OVER (PARTITION BY ur1.room_id ORDER BY ur1.id DESC) AS rn
					FROM user_room AS ur1
					INNER JOIN user_room ur2 ON ur2.room_id = ur1.room_id AND ur2.id != ur1.id
					WHERE ur2.user_id = @sender_id AND ur2.user_type = @sender_type
				)
				SELECT * FROM ranked_rooms WHERE rn = 1
			) r2
			INNER JOIN rooms r ON r.id = r2.room_id
			INNER JOIN (
				SELECT MAX(id) maxid, room_id
				FROM messages
				GROUP BY room_id
			) m2 ON m2.room_id = r2.room_id
			INNER JOIN messages m ON m.id = m2.maxid
			LEFT JOIN users u2 ON r2.user_type = @user_type_user AND u2.id = r2.user_id
			LEFT JOIN companies c2 ON r2.user_type = @user_type_company AND c2.id = r2.user_id
			ORDER BY m2.maxid DESC
			LIMIT @limit
			OFFSET @offset`,
			sql.Named("room_type_private", RoomTypePrivate),
			sql.Named("user_type_user", RoomUserTypeUser),
			sql.Named("user_type_company", RoomUserTypeCompany),
			sql.Named("sender_id", senderId),
			sql.Named("sender_type", senderType),
			sql.Named("limit", limit),
			sql.Named("offset", offset),
		).
		Rows()

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var result RoomInfoResult
		err := rows.Scan(&result.RecipientId, &result.RecipientType, &result.SenderId, &result.SenderType, &result.RoomId, &result.RoomType, &result.MessageId, &result.Message, &result.MessageDate, &result.Name, &result.Image)
		if err != nil {
			return nil, err
		}

		if result.SenderId == senderId && result.SenderType == senderType {
			logger.Info("IsSenderMessage")
			result.IsSenderMessage = true
		}

		list = append(list, result)
	}

	return list, nil
}
