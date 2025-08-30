package query

import (
	"context"
	"strings"
	"testing"

	"com.raunlo.checklist/internal/core/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// stripSQL collapses consecutive whitespace so comparisons are unaffected by formatting.
func stripSQL(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// mockRow implements pgx.Row and returns predetermined values.
type mockRow struct {
	scan func(dest ...any) error
}

func (r mockRow) Scan(dest ...any) error { return r.scan(dest...) }

// mockTx collects executed SQL commands and serves queued QueryRow results.
type mockTx struct {
	execs    []string
	queries  []string
	rowFuncs []func(dest ...any) error
}

func newMockTx(rowFuncs ...func(dest ...any) error) *mockTx {
	return &mockTx{rowFuncs: rowFuncs}
}

func (m *mockTx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	m.execs = append(m.execs, stripSQL(sql))
	return pgconn.CommandTag{}, nil
}

func (m *mockTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	m.queries = append(m.queries, stripSQL(sql))
	fn := m.rowFuncs[0]
	m.rowFuncs = m.rowFuncs[1:]
	return mockRow{fn}
}

// Unused TransactionWrapper methods.
func (m *mockTx) Begin(ctx context.Context) (pgx.Tx, error) { return nil, nil }
func (m *mockTx) Commit(ctx context.Context) error          { return nil }
func (m *mockTx) Rollback(ctx context.Context) error        { return nil }
func (m *mockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (m *mockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults { return nil }
func (m *mockTx) LargeObjects() pgx.LargeObjects                               { return pgx.LargeObjects{} }
func (m *mockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (m *mockTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, nil
}
func (m *mockTx) Conn() *pgx.Conn { return nil }
func (m *mockTx) QueryOne(ctx context.Context, sql string, dest interface{}, args pgx.NamedArgs) error {
	return nil
}
func (m *mockTx) QueryList(ctx context.Context, sql string, dest interface{}, args pgx.NamedArgs) error {
	return nil
}

func TestChangeChecklistItemOrder_MoveToEnd(t *testing.T) {
	tx := newMockTx(
		// First QueryRow: fetch current prev and next ids for the moving item.
		func(dest ...any) error {
			prev := uint(1)
			next := uint(3)
			*(dest[0].(**uint)) = &prev
			*(dest[1].(**uint)) = &next
			return nil
		},
		// Second QueryRow: find target position (tail item with no next).
		func(dest ...any) error {
			*(dest[0].(*uint)) = 3
			*(dest[1].(**uint)) = nil
			*(dest[2].(**uint)) = nil
			return nil
		},
	)

	fn := NewChangeChecklistItemOrderQueryFunction(domain.ChangeOrderRequest{
		ChecklistId:     1,
		ChecklistItemId: 2,
		NewOrderNumber:  1,
		SortOrder:       domain.AscSort,
	}).GetTransactionalQueryFunction()

	ok, err := fn(tx)
	if err != nil || !ok {
		t.Fatalf("change order failed: %v", err)
	}

	expectedQueries := []string{
		stripSQL(`SELECT PREV_ITEM_ID, NEXT_ITEM_ID FROM CHECKLIST_ITEM
                                WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @checklistItemId FOR UPDATE`),
		stripSQL(`WITH RECURSIVE CHECKLIST_ITEMS_CTE as (
                                SELECT CHECKLIST_ITEM_ID, NEXT_ITEM_ID, PREV_ITEM_ID,1 as ORDER_NUMBER
                                FROM CHECKLIST_ITEM WHERE CHECKLIST_ID = @checklistId  AND NEXT_ITEM_ID IS NULL AND CHECKLIST_ITEM_ID <> @itemToMoveId

                                UNION ALL

                                SELECT CHECKLIST_ITEM.CHECKLIST_ITEM_ID, CHECKLIST_ITEM.NEXT_ITEM_ID, CHECKLIST_ITEM.PREV_ITEM_ID, ORDER_NUMBER + 1  as ORDER_NUMBER
                                FROM CHECKLIST_ITEM, CHECKLIST_ITEMS_CTE
                                WHERE CHECKLIST_ITEM.CHECKLIST_ID = @checklistId  AND CHECKLIST_ITEMS_CTE.CHECKLIST_ITEM_ID =  CHECKLIST_ITEM.NEXT_ITEM_ID)

                                SELECT CHECKLIST_ITEM_ID, PREV_ITEM_ID, NEXT_ITEM_ID from CHECKLIST_ITEMS_CTE where ORDER_NUMBER = @orderNumber`),
	}
	expectedExecs := []string{
		stripSQL(`UPDATE CHECKLIST_ITEM SET NEXT_ITEM_ID = NULL, PREV_ITEM_ID = NULL
                                WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @checklistItemId`),
		stripSQL(`UPDATE CHECKLIST_ITEM
                                        SET NEXT_ITEM_ID = @nextItemId
                                        WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @prevItemId`),
		stripSQL(`UPDATE CHECKLIST_ITEM
                                        SET PREV_ITEM_ID = @prevItemId
                                        WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @nextItemId`),
		stripSQL(`UPDATE CHECKLIST_ITEM SET NEXT_ITEM_ID = @itemToMoveId
                                WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @newPrevItemId`),
		stripSQL(`UPDATE CHECKLIST_ITEM SET NEXT_ITEM_ID = @newNextItemId
                        WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @itemToMoveId`),
		stripSQL(`UPDATE CHECKLIST_ITEM SET PREV_ITEM_ID = @newPrevItemId
                        WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @itemToMoveId`),
	}
	if len(tx.execs) != len(expectedExecs) {
		t.Fatalf("expected %d execs got %d", len(expectedExecs), len(tx.execs))
	}
	for i := range expectedExecs {
		if tx.execs[i] != expectedExecs[i] {
			t.Errorf("exec %d = %q, want %q", i, tx.execs[i], expectedExecs[i])
		}
	}
	if len(tx.queries) != len(expectedQueries) {
		t.Fatalf("expected %d queries got %d", len(expectedQueries), len(tx.queries))
	}
	for i := range expectedQueries {
		if tx.queries[i] != expectedQueries[i] {
			t.Errorf("query %d = %q, want %q", i, tx.queries[i], expectedQueries[i])
		}
	}
}

func TestChangeChecklistItemOrder_MoveBetweenItems(t *testing.T) {
	tx := newMockTx(
		// Initial prev and next of moving item.
		func(dest ...any) error {
			prev := uint(1)
			next := uint(3)
			*(dest[0].(**uint)) = &prev
			*(dest[1].(**uint)) = &next
			return nil
		},
		// Target position: after item3 and before item4.
		func(dest ...any) error {
			*(dest[0].(*uint)) = 3
			*(dest[1].(**uint)) = nil
			next := uint(4)
			*(dest[2].(**uint)) = &next
			return nil
		},
	)

	fn := NewChangeChecklistItemOrderQueryFunction(domain.ChangeOrderRequest{
		ChecklistId:     1,
		ChecklistItemId: 2,
		NewOrderNumber:  2,
		SortOrder:       domain.AscSort,
	}).GetTransactionalQueryFunction()

	ok, err := fn(tx)
	if err != nil || !ok {
		t.Fatalf("change order failed: %v", err)
	}

	expectedQueries := []string{
		stripSQL(`SELECT PREV_ITEM_ID, NEXT_ITEM_ID FROM CHECKLIST_ITEM
                                WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @checklistItemId FOR UPDATE`),
		stripSQL(`WITH RECURSIVE CHECKLIST_ITEMS_CTE as (
                                SELECT CHECKLIST_ITEM_ID, NEXT_ITEM_ID, PREV_ITEM_ID,1 as ORDER_NUMBER
                                FROM CHECKLIST_ITEM WHERE CHECKLIST_ID = @checklistId  AND NEXT_ITEM_ID IS NULL AND CHECKLIST_ITEM_ID <> @itemToMoveId

                                UNION ALL

                                SELECT CHECKLIST_ITEM.CHECKLIST_ITEM_ID, CHECKLIST_ITEM.NEXT_ITEM_ID, CHECKLIST_ITEM.PREV_ITEM_ID, ORDER_NUMBER + 1  as ORDER_NUMBER
                                FROM CHECKLIST_ITEM, CHECKLIST_ITEMS_CTE
                                WHERE CHECKLIST_ITEM.CHECKLIST_ID = @checklistId  AND CHECKLIST_ITEMS_CTE.CHECKLIST_ITEM_ID =  CHECKLIST_ITEM.NEXT_ITEM_ID)

                                SELECT CHECKLIST_ITEM_ID, PREV_ITEM_ID, NEXT_ITEM_ID from CHECKLIST_ITEMS_CTE where ORDER_NUMBER = @orderNumber`),
	}
	expectedExecs := []string{
		stripSQL(`UPDATE CHECKLIST_ITEM SET NEXT_ITEM_ID = NULL, PREV_ITEM_ID = NULL
                                WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @checklistItemId`),
		stripSQL(`UPDATE CHECKLIST_ITEM
                                        SET NEXT_ITEM_ID = @nextItemId
                                        WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @prevItemId`),
		stripSQL(`UPDATE CHECKLIST_ITEM
                                        SET PREV_ITEM_ID = @prevItemId
                                        WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @nextItemId`),
		stripSQL(`UPDATE CHECKLIST_ITEM SET NEXT_ITEM_ID = @itemToMoveId
                                WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @newPrevItemId`),
		stripSQL(`UPDATE CHECKLIST_ITEM SET PREV_ITEM_ID = @itemToMoveId
                                WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @newNextItemId`),
		stripSQL(`UPDATE CHECKLIST_ITEM SET NEXT_ITEM_ID = @newNextItemId
                        WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @itemToMoveId`),
		stripSQL(`UPDATE CHECKLIST_ITEM SET PREV_ITEM_ID = @newPrevItemId
                        WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @itemToMoveId`),
	}
	if len(tx.execs) != len(expectedExecs) {
		t.Fatalf("expected %d execs got %d", len(expectedExecs), len(tx.execs))
	}
	for i := range expectedExecs {
		if tx.execs[i] != expectedExecs[i] {
			t.Errorf("exec %d = %q, want %q", i, tx.execs[i], expectedExecs[i])
		}
	}
	if len(tx.queries) != len(expectedQueries) {
		t.Fatalf("expected %d queries got %d", len(expectedQueries), len(tx.queries))
	}
	for i := range expectedQueries {
		if tx.queries[i] != expectedQueries[i] {
			t.Errorf("query %d = %q, want %q", i, tx.queries[i], expectedQueries[i])
		}
	}
}
