package cloudmigrationimpl

import (
	"context"
	"strconv"
	"testing"

	"github.com/grafana/grafana/pkg/infra/db"
	"github.com/grafana/grafana/pkg/tests/testsuite"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	testsuite.Run(m)
}

func TestMigrateDatasources(t *testing.T) {
	// TODO: Write this test
}

func TestGetAllCloudMigrations(t *testing.T) {
	testDB := db.InitTestDB(t)
	s := &sqlStore{db: testDB}
	ctx := context.Background()

	t.Run("get all cloud_migrations", func(t *testing.T) {
		// replace this with proper method when created
		_, err := testDB.GetSqlxSession().Exec(ctx, `
			INSERT INTO cloud_migration (id, auth_token, stack, created, updated)
			VALUES (1, '12345', 'stack1', '2024-03-25 15:30:36.000', '2024-03-27 15:30:43.000'),
 				(2, '6789', 'stack2', '2024-03-25 15:30:36.000', '2024-03-27 15:30:43.000'),
 				(3, '777', 'stack3', '2024-03-25 15:30:36.000', '2024-03-27 15:30:43.000');
		`)
		require.NoError(t, err)

		value, err := s.GetAllCloudMigrations(ctx)
		require.NoError(t, err)
		require.Equal(t, 3, len(value))
		for _, m := range value {
			switch m.ID {
			case 1:
				require.Equal(t, "stack1", m.Stack)
				require.Equal(t, "12345", m.AuthToken)
			case 2:
				require.Equal(t, "stack2", m.Stack)
				require.Equal(t, "6789", m.AuthToken)
			case 3:
				require.Equal(t, "stack3", m.Stack)
				require.Equal(t, "777", m.AuthToken)
			default:
				require.Fail(t, "ID value not expected: "+strconv.FormatInt(m.ID, 10))
			}
		}
	})
}
