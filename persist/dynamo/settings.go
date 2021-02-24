package dynamo

const settingsType = "settings"

// SelectSettings selects the Settings object if it exists or returns ErrNoRows
//func (p Persister) SelectSettings(ctx context.Context) (*model.Settings, error) {
//	var settings model.Settings
//	gii := &dynamodb.GetItemInput{
//		TableName: p.tableName,
//		Key: map[string]*dynamodb.AttributeValue{
//			pkName: {
//				S: aws.String(settingsType),
//			},
//			skName: {
//				S: aws.String(settingsType),
//			},
//		},
//	}
//	// log.Printf("[DEBUG] SelectSettings: GetItem(): %#v", gii)
//	gio, err := p.svc.GetItem(gii)
//	// log.Printf("[DEBUG] SelectSettings: gio: %#v, err: %#v", gio, err)
//	if err != nil {
//		log.Printf("[ERROR] Failed to get settings. qi: %#v err: %v", gio, err)
//		return nil, model.NewError(model.ErrOther, err.Error())
//	}
//	if gio.Item == nil {
//		return nil, model.NewError(model.ErrNotFound, settingsType)
//	}
//	err = dynamodbattribute.UnmarshalMap(gio.Item, &settings)
//	if err != nil {
//		log.Printf("[ERROR] Failed to unmarshal. qo: %#v err: %v", gio, err)
//		return nil, model.NewError(model.ErrOther, err.Error())
//	}
//	return &settings, nil
//}

// UpsertSettings updates or inserts a Settings object in the database and returns the updated Settings
//func (p Persister) UpsertSettings(ctx context.Context, in model.Settings) (*model.Settings, error) {
//	in.ID = settingsType
//	in.Sk = settingsType
//	settings := in
//	now := time.Now().Truncate(0)
//	// Since we're doing upsert, can't easily distinguish insert and update time
//	settings.InsertTime = now
//	settings.LastUpdateTime = now
//
//	avs, err := dynamodbattribute.MarshalMap(settings)
//	if err != nil {
//		log.Printf("[ERROR] Failed to marshal settings %#v: %v", settings, err)
//		return nil, model.NewError(model.ErrOther, err.Error())
//	}
//	lastUpdateTime := in.LastUpdateTime.Format(time.RFC3339Nano)
//	pii := &dynamodb.PutItemInput{
//		TableName:           p.tableName,
//		Item:                avs,
//		ConditionExpression: aws.String("attribute_not_exists(pk) OR last_update_time = :lastUpdateTime"), // Either an insert or last_update_time must match
//		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
//			":lastUpdateTime": {S: &lastUpdateTime},
//		},
//		ReturnValues: aws.String("ALL_OLD"),
//	}
//	// log.Printf("[DEBUG] UpsertSettings: PutItem(): %#v", pii)
//	_, err = p.svc.PutItem(pii)
//	// log.Printf("[DEBUG] UpsertSettings: pio: %#v, err: %#v", pio, err)
//	if err != nil {
//		if compareToAWSError(err, dynamodb.ErrCodeConditionalCheckFailedException) {
//			return nil, model.NewError(model.ErrConcurrentUpdate /* *pio.Attributes["lastUpdateTime"].S */, "", lastUpdateTime)
//		}
//		log.Printf("[ERROR] Failed to update settings %#v. pii: %#v err: %v", settings, pii, err)
//		return nil, model.NewError(model.ErrOther, err.Error())
//	}
//	return p.SelectSettings(ctx)
//	// err := p.db.QueryRowContext(ctx,
//	// 	`UPDATE settings SET body = $1, last_update_time = CURRENT_TIMESTAMP
//	// 	 WHERE id = $2 AND last_update_time = $3
//	// 	 RETURNING body, insert_time, last_update_time`,
//	// 	in.SettingsBody, SettingsID, in.LastUpdateTime).
//	// 	Scan(
//	// 		&settings.SettingsBody,
//	// 		&settings.InsertTime,
//	// 		&settings.LastUpdateTime,
//	// 	)
//	// if err != nil && err == sql.ErrNoRows {
//	// 	// Either non-existent or last_update_time didn't match
//	// 	s, err := p.SelectSettings(ctx)
//	// 	if err == nil {
//	// 		// Row exists, so it must be a non-matching update time
//	// 		return nil, model.NewError(model.ErrConcurrentUpdate, s.LastUpdateTime.String(), in.LastUpdateTime.String())
//	// 	} else if model.ErrNotFound.Matches(err) {
//	// 		// row doesn't exist; need to insert
//	// 		err := p.db.QueryRowContext(ctx,
//	// 			`INSERT INTO settings (id, body)
//	// 			VALUES ($1, $2)
//	// 	 		RETURNING body, insert_time, last_update_time`,
//	// 			SettingsID, in.SettingsBody).
//	// 			Scan(
//	// 				&settings.SettingsBody,
//	// 				&settings.InsertTime,
//	// 				&settings.LastUpdateTime,
//	// 			)
//	// 		id := uint32(SettingsID)
//	// 		return &settings, translateError(err, &id, nil, "")
//	// 	}
//	// }
//}
