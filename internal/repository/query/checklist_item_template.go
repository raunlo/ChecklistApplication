package query

import (
	"context"
	"errors"

	"com.raunlo.checklist/internal/core/domain"
	"github.com/jackc/pgx/v5"
	"github.com/raunlo/pgx-with-automapper/pool"
)

type PersistChecklistItemTemplateQueryFunction struct {
	checklistItemTemplate domain.ChecklistItemTemplate
}

func (p *PersistChecklistItemTemplateQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (domain.ChecklistItemTemplate, error) {
	return func(tx pool.TransactionWrapper) (domain.ChecklistItemTemplate, error) {
		sql := `INSERT INTO TEMPLATE_ITEM(ID, NAME) VALUES(nextval('template_item_id'), @name) RETURNING ID`
		row := tx.QueryRow(context.Background(), sql, pgx.NamedArgs{"name": "__template__"})
		err := row.Scan(&p.checklistItemTemplate.Id)
		return p.checklistItemTemplate, err
	}

}

type UpdateChecklistItemTemplateQueryFunction struct {
	checklistItemTemplate domain.ChecklistItemTemplate
}

func (u *UpdateChecklistItemTemplateQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (bool, error) {
	return func(tx pool.TransactionWrapper) (bool, error) {
		sql := `UPDATE TEMPLATE_ITEM SET NAME = NAME WHERE ID = @id`
		tag, err := tx.Exec(context.Background(), sql, pgx.NamedArgs{"id": u.checklistItemTemplate.Id})
		return tag.RowsAffected() == 1, err
	}
}

type DeleteChecklistItemTemplateByIdQueryFunction struct {
	checklistItemTemplateId uint
}

func (d *DeleteChecklistItemTemplateByIdQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (bool, error) {
	return func(tx pool.TransactionWrapper) (bool, error) {
		_, err := tx.Exec(context.Background(), `DELETE FROM TEMPLATE_ITEM_ROW WHERE TEMPLATE_ITEM_ID = @id`, pgx.NamedArgs{"id": d.checklistItemTemplateId})
		if err != nil {
			return false, err
		}
		tag, err := tx.Exec(context.Background(), `DELETE FROM TEMPLATE_ITEM WHERE ID = @id`, pgx.NamedArgs{"id": d.checklistItemTemplateId})
		if err != nil {
			return false, err
		}
		if tag.RowsAffected() > 1 {
			return false, errors.New("deleteChecklistItemTemplateById affected more than one row")
		}
		return tag.RowsAffected() == 1, nil
	}

}

type FindChecklistItemTemplateByIdQueryFunction struct {
	checklistItemTemplateId uint
}

func (f FindChecklistItemTemplateByIdQueryFunction) GetQueryFunction() func(connection pool.Conn) (*domain.ChecklistItemTemplate, error) {
	return func(connection pool.Conn) (*domain.ChecklistItemTemplate, error) {
		sql := `SELECT t.ID, r.ID FROM TEMPLATE_ITEM t LEFT JOIN TEMPLATE_ITEM_ROW r ON t.ID = r.TEMPLATE_ITEM_ID WHERE t.ID = @id`
		rows, err := connection.Query(context.Background(), sql, pgx.NamedArgs{"id": f.checklistItemTemplateId})
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var template *domain.ChecklistItemTemplate
		for rows.Next() {
			var tId uint
			var rId *uint
			if err := rows.Scan(&tId, &rId); err != nil {
				return nil, err
			}
			if template == nil {
				template = &domain.ChecklistItemTemplate{Id: tId}
			}
			if rId != nil {
				template.Rows = append(template.Rows, domain.ChecklistItemTemplateRow{Id: *rId})
			}
		}
		if template == nil {
			return nil, nil
		}
		return template, rows.Err()
	}

}

type GetAllChecklistItemTemplatesQueryFunction struct {
}

func (g *GetAllChecklistItemTemplatesQueryFunction) GetQueryFunction() func(connection pool.Conn) ([]domain.ChecklistItemTemplate, error) {
	return func(connection pool.Conn) ([]domain.ChecklistItemTemplate, error) {
		sql := `SELECT t.ID, r.ID FROM TEMPLATE_ITEM t LEFT JOIN TEMPLATE_ITEM_ROW r ON t.ID = r.TEMPLATE_ITEM_ID ORDER BY t.ID, r.ID`
		rows, err := connection.Query(context.Background(), sql)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		templatesMap := map[uint]*domain.ChecklistItemTemplate{}
		for rows.Next() {
			var tId uint
			var rId *uint
			if err := rows.Scan(&tId, &rId); err != nil {
				return nil, err
			}
			tpl := templatesMap[tId]
			if tpl == nil {
				tpl = &domain.ChecklistItemTemplate{Id: tId}
				templatesMap[tId] = tpl
			}
			if rId != nil {
				tpl.Rows = append(tpl.Rows, domain.ChecklistItemTemplateRow{Id: *rId})
			}
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		templates := make([]domain.ChecklistItemTemplate, 0, len(templatesMap))
		for _, t := range templatesMap {
			templates = append(templates, *t)
		}
		return templates, nil
	}
}
