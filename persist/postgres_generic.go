package persist

// import (
// 	"context"
// 	"database/sql"
// 	"fmt"
// 	"log"

// 	"github.com/cheekybits/genny/generic"
// 	"github.com/ourrootsorg/cms-server/model"
// )

// type Entity generic.Type

// // SelectEntities loads all the entities from the database
// func (p PostgresPersister) SelectEntities(ctx context.Context) ([]Entity, error) {
// 	rows, err := p.db.QueryContext(ctx, "SELECT id, body, insert_time, last_update_time FROM entity")
// 	if err != nil {
// 		return nil, translateError(err)
// 	}
// 	defer rows.Close()
// 	cats := make([]Entity, 0)
// 	for rows.Next() {
// 		var dbid int32
// 		var cat Entity
// 		err := rows.Scan(&dbid, &cat.EntityBody, &cat.InsertTime, &cat.LastUpdateTime)
// 		if err != nil {
// 			return nil, translateError(err)
// 		}
// 		cat.ID = p.pathPrefix + fmt.Sprintf(EntityIDFormat, dbid)
// 		cat.Type = "entity"
// 		cats = append(cats, cat)
// 	}
// 	return cats, nil
// }

// // SelectOneEntity loads a single entity from the database
// func (p PostgresPersister) SelectOneEntity(ctx context.Context, id string) (Entity, error) {
// 	var cat Entity
// 	var dbid int32
// 	n, err := fmt.Sscanf(id+"\n", p.pathPrefix+EntityIDFormat+"\n", &dbid)
// 	if err != nil || n != 1 {
// 		// Bad ID
// 		return cat, model.NewError(model.ErrNotFound, id)
// 	}
// 	log.Printf("[DEBUG] id: %s, dbid: %d", id, dbid)
// 	err = p.db.QueryRowContext(ctx, "SELECT id, body, insert_time, last_update_time FROM entity WHERE id=$1", dbid).Scan(
// 		&dbid,
// 		&cat.EntityBody,
// 		&cat.InsertTime,
// 		&cat.LastUpdateTime,
// 	)
// 	if err != nil {
// 		return cat, translateError(err)
// 	}
// 	cat.ID = p.pathPrefix + fmt.Sprintf(EntityIDFormat, dbid)
// 	cat.Type = "entity"
// 	return cat, nil
// }

// // InsertEntity inserts a EntityBody into the database and returns the inserted Entity
// func (p PostgresPersister) InsertEntity(ctx context.Context, in EntityIn) (Entity, error) {
// 	var dbid int32
// 	var cat Entity
// 	row := p.db.QueryRowContext(ctx, "INSERT INTO entity (body) VALUES ($1) RETURNING id, body, insert_time, last_update_time", in)
// 	err := row.Scan(
// 		&dbid,
// 		&cat.EntityBody,
// 		&cat.InsertTime,
// 		&cat.LastUpdateTime,
// 	)
// 	cat.ID = p.pathPrefix + fmt.Sprintf(EntityIDFormat, dbid)
// 	cat.Type = "entity"
// 	return cat, translateError(err)
// }

// // UpdateEntity updates a Entity in the database and returns the updated Entity
// func (p PostgresPersister) UpdateEntity(ctx context.Context, id string, in Entity) (Entity, error) {
// 	var cat Entity
// 	var dbid int32
// 	n, err := fmt.Sscanf(id+"\n", p.pathPrefix+EntityIDFormat+"\n", &dbid)
// 	if err != nil || n != 1 {
// 		// Bad ID
// 		return cat, model.NewError(model.ErrNotFound, id)
// 	}
// 	err = p.db.QueryRowContext(ctx, "UPDATE entity SET body = $1, last_update_time = CURRENT_TIMESTAMP WHERE id = $2 AND last_update_time = $3 RETURNING id, body, insert_time, last_update_time", in.EntityBody, dbid, in.LastUpdateTime).
// 		Scan(
// 			&dbid,
// 			&cat.EntityBody,
// 			&cat.InsertTime,
// 			&cat.LastUpdateTime,
// 		)
// 	if err != nil && err == sql.ErrNoRows {
// 		// Either non-existent or last_update_time didn't match
// 		c, _ := p.SelectOneEntity(ctx, id)
// 		if c.ID == id {
// 			// Row exists, so it must be a non-matching update time
// 			return cat, model.NewError(model.ErrConcurrentUpdate, c.LastUpdateTime.String(), in.LastUpdateTime.String())
// 		}
// 		return cat, model.NewError(model.ErrNotFound, id)
// 	}
// 	cat.ID = p.pathPrefix + fmt.Sprintf(EntityIDFormat, dbid)
// 	cat.Type = "entity"
// 	return cat, translateError(err)
// }

// // DeleteEntity deletes a Entity
// func (p PostgresPersister) DeleteEntity(ctx context.Context, id string) error {
// 	var dbid int32
// 	n, err := fmt.Sscanf(id+"\n", p.pathPrefix+EntityIDFormat+"\n", &dbid)
// 	if err != nil || n != 1 {
// 		// Bad ID
// 		return model.NewError(model.ErrNotFound, id)
// 	}
// 	_, err = p.db.ExecContext(ctx, "DELETE FROM entity WHERE id = $1", dbid)
// 	return translateError(err)
// }
