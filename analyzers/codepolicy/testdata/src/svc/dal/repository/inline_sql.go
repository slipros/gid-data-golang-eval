package repository

import (
	"context"
	"errors"
	"fmt"
)

type connection struct{}

func (connection) ScanRow(context.Context, []any, string, any) error { return nil }

func inlineQueries(ctx context.Context, conn connection, table string) error {
	var id string
	args := map[string]any{"id": 1}
	// want +1 `GID-297:`
	err := conn.ScanRow(ctx, []any{&id}, `
 INSERT INTO public.work_order_inbox (message_id)
 VALUES (@id) ON CONFLICT DO NOTHING RETURNING id::text`, args)
	if err != nil {
		return err
	}
	// want +1 `GID-297:`
	if err := conn.ScanRow(ctx, []any{&id}, `
 SELECT id FROM public.work_order_inbox WHERE message_id=@id`, args); err != nil {
		return err
	}
	conn.ScanRow(ctx, nil, "SELECT 1", args)         // want `GID-297:`
	conn.ScanRow(ctx, nil, "SELECT "+shortSQL, args) // want `GID-297:`
	conn.ScanRow(ctx, nil, ("SELECT 1"), args)       // want `GID-297:`
	conn.ScanRow(ctx, nil, string("SELECT 1"), args) // want `GID-297:`
	const query = "SELECT 1"
	conn.ScanRow(ctx, nil, query, args)
	conn.ScanRow(ctx, nil, (query), args)
	conn.ScanRow(ctx, nil, "SELECT * FROM "+table, args)
	conn.ScanRow(ctx, nil, "not a SQL query", args)
	return nil
}

func queryErrors(err error) {
	errors.New("insert delivery inbox")
	fmt.Errorf("insert delivery inbox: %w", err)
	errors.New("update delivery inbox")
	errors.New("delete delivery inbox")
	errors.New("with delivery inbox")
}

func statementShapes(conn connection) {
	conn.ScanRow(nil, nil, "/* delivery */ INSERT /* inbox */ INTO inbox (id) VALUES (@id)", nil) // want `GID-297:`
	conn.ScanRow(nil, nil, "WITH inbox AS NOT MATERIALIZED (SELECT 1) SELECT * FROM inbox", nil)  // want `GID-297:`
	conn.ScanRow(nil, nil, "insert into inbox (id) values (@id)", nil)                            // want `GID-297:`
	conn.ScanRow(nil, nil, "UPDATE inbox SET id=@id", nil)                                        // want `GID-297:`
	conn.ScanRow(nil, nil, "DELETE FROM inbox WHERE id=@id", nil)                                 // want `GID-297:`
	conn.ScanRow(nil, nil, "WITH inbox AS (SELECT 1) SELECT * FROM inbox", nil)                   // want `GID-297:`
	conn.ScanRow(nil, nil, "WITH RECURSIVE inbox(id) AS (SELECT 1) SELECT * FROM inbox", nil)     // want `GID-297:`
	conn.ScanRow(nil, nil, "UPDATE \"work order\" SET id=@id", nil)                               // want `GID-297:`
}

func lowercaseStatements(conn connection) {
	conn.ScanRow(nil, nil, "select 1", nil)                                     // want `GID-297:`
	conn.ScanRow(nil, nil, "insert into inbox (id) values (@id)", nil)          // want `GID-297:`
	conn.ScanRow(nil, nil, "update inbox set id=@id", nil)                      // want `GID-297:`
	conn.ScanRow(nil, nil, "delete from inbox where id=@id", nil)               // want `GID-297:`
	conn.ScanRow(nil, nil, "with inbox as (select 1) select * from inbox", nil) // want `GID-297:`
	conn.ScanRow(nil, nil, "iNsErT iNtO inbox (id) vAlUeS (@id)", nil)          // want `GID-297:`
}
