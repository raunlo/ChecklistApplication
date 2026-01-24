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

// mockRows implements pgx.Rows for returning multiple position values.
type mockRows struct {
	positions []float64
	index     int
	closed    bool
}

func (r *mockRows) Next() bool {
	if r.index >= len(r.positions) {
		return false
	}
	return true
}

func (r *mockRows) Scan(dest ...any) error {
	if r.index < len(r.positions) {
		*(dest[0].(*float64)) = r.positions[r.index]
		r.index++
	}
	return nil
}

func (r *mockRows) Close()                                       { r.closed = true }
func (r *mockRows) Err() error                                   { return nil }
func (r *mockRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (r *mockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *mockRows) Values() ([]any, error)                       { return nil, nil }
func (r *mockRows) RawValues() [][]byte                          { return nil }
func (r *mockRows) Conn() *pgx.Conn                              { return nil }

// mockTx collects executed SQL commands and serves queued QueryRow results.
type mockTx struct {
	execs       []string
	queries     []string
	rowFuncs    []func(dest ...any) error
	rowsResults []*mockRows
	rowsIndex   int
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

func (m *mockTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	m.queries = append(m.queries, stripSQL(sql))
	if m.rowsIndex < len(m.rowsResults) {
		rows := m.rowsResults[m.rowsIndex]
		m.rowsIndex++
		return rows, nil
	}
	return &mockRows{positions: []float64{}}, nil
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

func (m *mockTx) Conn() *pgx.Conn { return nil }
func (m *mockTx) QueryOne(ctx context.Context, sql string, dest interface{}, args pgx.NamedArgs) error {
	return nil
}

func (m *mockTx) QueryList(ctx context.Context, sql string, dest interface{}, args pgx.NamedArgs) error {
	return nil
}

func TestChangeChecklistItemOrder_InsertBetween(t *testing.T) {
	// Setup: items at positions 1000, 2000, 3000
	// Move item at position 3000 to position 2 (between 1000 and 2000)
	// Expected new position: (1000 + 2000) / 2 = 1500

	tx := newMockTx(
		// First QueryRow: get item's completed status
		func(dest ...any) error {
			*(dest[0].(*bool)) = false // uncompleted
			return nil
		},
		// Third QueryRow: check min gap for rebalancing
		func(dest ...any) error {
			*(dest[0].(*float64)) = 500.0 // gap is big, no rebalance needed
			return nil
		},
	)
	// Second Query: get positions (excluding moving item)
	tx.rowsResults = []*mockRows{
		{positions: []float64{1000.0, 2000.0}}, // other items
	}

	fn := NewChangeChecklistItemOrderQueryFunction(domain.ChangeOrderRequest{
		ChecklistId:     1,
		ChecklistItemId: 3,
		NewOrderNumber:  2, // Insert at position 2 (between items at index 0 and 1)
		SortOrder:       domain.AscSort,
	}).GetTransactionalQueryFunction()

	response, err := fn(tx)
	if err != nil {
		t.Fatalf("change order failed: %v", err)
	}

	// Expected position: (1000 + 2000) / 2 = 1500
	expectedPosition := 1500.0
	if response.Position != expectedPosition {
		t.Errorf("expected position %f, got %f", expectedPosition, response.Position)
	}
	if response.RebalanceNeeded {
		t.Error("expected no rebalance needed")
	}
}

func TestChangeChecklistItemOrder_InsertAtStart(t *testing.T) {
	// Setup: items at positions 1000, 2000
	// Move an item to position 1 (at the start)
	// Expected new position: 1000 - 1000 = 0

	tx := newMockTx(
		// First QueryRow: get item's completed status
		func(dest ...any) error {
			*(dest[0].(*bool)) = false // uncompleted
			return nil
		},
		// Third QueryRow: check min gap for rebalancing
		func(dest ...any) error {
			*(dest[0].(*float64)) = 1000.0
			return nil
		},
	)
	tx.rowsResults = []*mockRows{
		{positions: []float64{1000.0, 2000.0}},
	}

	fn := NewChangeChecklistItemOrderQueryFunction(domain.ChangeOrderRequest{
		ChecklistId:     1,
		ChecklistItemId: 3,
		NewOrderNumber:  1, // Insert at start
		SortOrder:       domain.AscSort,
	}).GetTransactionalQueryFunction()

	response, err := fn(tx)
	if err != nil {
		t.Fatalf("change order failed: %v", err)
	}

	// Expected position: 1000 - DefaultGapSize = 0
	expectedPosition := 0.0
	if response.Position != expectedPosition {
		t.Errorf("expected position %f, got %f", expectedPosition, response.Position)
	}
}

func TestChangeChecklistItemOrder_InsertAtEnd(t *testing.T) {
	// Setup: items at positions 1000, 2000
	// Move an item to position 3 (at the end)
	// Expected new position: 2000 + 1000 = 3000

	tx := newMockTx(
		// First QueryRow: get item's completed status
		func(dest ...any) error {
			*(dest[0].(*bool)) = false // uncompleted
			return nil
		},
		// Third QueryRow: check min gap for rebalancing
		func(dest ...any) error {
			*(dest[0].(*float64)) = 1000.0
			return nil
		},
	)
	tx.rowsResults = []*mockRows{
		{positions: []float64{1000.0, 2000.0}},
	}

	fn := NewChangeChecklistItemOrderQueryFunction(domain.ChangeOrderRequest{
		ChecklistId:     1,
		ChecklistItemId: 3,
		NewOrderNumber:  3, // Insert at end
		SortOrder:       domain.AscSort,
	}).GetTransactionalQueryFunction()

	response, err := fn(tx)
	if err != nil {
		t.Fatalf("change order failed: %v", err)
	}

	// Expected position: 2000 + DefaultGapSize = 3000
	expectedPosition := 3000.0
	if response.Position != expectedPosition {
		t.Errorf("expected position %f, got %f", expectedPosition, response.Position)
	}
}

func TestChangeChecklistItemOrder_EmptyList(t *testing.T) {
	// Setup: no other items in checklist
	// Expected new position: FirstItemPosition (1000)

	tx := newMockTx(
		// First QueryRow: get item's completed status
		func(dest ...any) error {
			*(dest[0].(*bool)) = false // uncompleted
			return nil
		},
		// Third QueryRow: check min gap for rebalancing
		func(dest ...any) error {
			*(dest[0].(*float64)) = 1000.0
			return nil
		},
	)
	tx.rowsResults = []*mockRows{
		{positions: []float64{}}, // no other items
	}

	fn := NewChangeChecklistItemOrderQueryFunction(domain.ChangeOrderRequest{
		ChecklistId:     1,
		ChecklistItemId: 1,
		NewOrderNumber:  1,
		SortOrder:       domain.AscSort,
	}).GetTransactionalQueryFunction()

	response, err := fn(tx)
	if err != nil {
		t.Fatalf("change order failed: %v", err)
	}

	// Expected position: FirstItemPosition = 1000
	expectedPosition := domain.FirstItemPosition
	if response.Position != expectedPosition {
		t.Errorf("expected position %f, got %f", expectedPosition, response.Position)
	}
}

func TestChangeChecklistItemOrder_TriggersRebalance(t *testing.T) {
	// Setup: items with very small gaps
	// Expected: RebalanceNeeded = true

	tx := newMockTx(
		// First QueryRow: get item's completed status
		func(dest ...any) error {
			*(dest[0].(*bool)) = false
			return nil
		},
		// Third QueryRow: check min gap - very small
		func(dest ...any) error {
			*(dest[0].(*float64)) = 0.0001 // smaller than MinGapThreshold (0.001)
			return nil
		},
	)
	tx.rowsResults = []*mockRows{
		{positions: []float64{1000.0, 1000.0001, 1000.0002}},
	}

	fn := NewChangeChecklistItemOrderQueryFunction(domain.ChangeOrderRequest{
		ChecklistId:     1,
		ChecklistItemId: 4,
		NewOrderNumber:  2,
		SortOrder:       domain.AscSort,
	}).GetTransactionalQueryFunction()

	response, err := fn(tx)
	if err != nil {
		t.Fatalf("change order failed: %v", err)
	}

	if !response.RebalanceNeeded {
		t.Error("expected rebalance needed due to small gap")
	}
}
