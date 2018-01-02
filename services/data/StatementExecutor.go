package data

import (
	"github.com/influxdata/influxdb/coordinator"
	"github.com/influxdata/influxql"
	"errors"
	"github.com/influxdata/influxdb/query"
)

type StatementExecutor struct {
	coordinator.StatementExecutor

	MetaExecutor *MetaExecutor
}

// ExecuteStatement executes the given statement with the given execution context.
func (e *StatementExecutor) ExecuteStatement(stmt influxql.Statement, ctx query.ExecutionContext) error {
	// Select statements are handled separately so that they can be streamed.
	if stmt, ok := stmt.(*influxql.SelectStatement); ok {
		return e.StatementExecutor.ExecuteStatement(stmt, ctx)
	}

	var messages []*query.Message
	var err error
	switch stmt := stmt.(type) {
	case *influxql.AlterRetentionPolicyStatement:
		if ctx.ReadOnly {
			messages = append(messages, query.ReadOnlyWarning(stmt.String()))
		}
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.CreateContinuousQueryStatement:
		if ctx.ReadOnly {
			messages = append(messages, query.ReadOnlyWarning(stmt.String()))
		}
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.CreateDatabaseStatement:
		if ctx.ReadOnly {
			messages = append(messages, query.ReadOnlyWarning(stmt.String()))
		}
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.CreateRetentionPolicyStatement:
		if ctx.ReadOnly {
			messages = append(messages, query.ReadOnlyWarning(stmt.String()))
		}
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.CreateSubscriptionStatement:
		if ctx.ReadOnly {
			messages = append(messages, query.ReadOnlyWarning(stmt.String()))
		}
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.CreateUserStatement:
		if ctx.ReadOnly {
			messages = append(messages, query.ReadOnlyWarning(stmt.String()))
		}
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.DeleteSeriesStatement:
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.DropContinuousQueryStatement:
		if ctx.ReadOnly {
			messages = append(messages, query.ReadOnlyWarning(stmt.String()))
		}
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.DropDatabaseStatement:
		if ctx.ReadOnly {
			messages = append(messages, query.ReadOnlyWarning(stmt.String()))
		}
		err = e.executeDropDatabaseStatement(stmt)
	case *influxql.DropMeasurementStatement:
		if ctx.ReadOnly {
			messages = append(messages, query.ReadOnlyWarning(stmt.String()))
		}
		err = e.executeDropMeasurementStatement(stmt, ctx.Database)
	case *influxql.DropSeriesStatement:
		if ctx.ReadOnly {
			messages = append(messages, query.ReadOnlyWarning(stmt.String()))
		}
		err = e.executeDropSeriesStatement(stmt, ctx.Database)
	case *influxql.DropRetentionPolicyStatement:
		if ctx.ReadOnly {
			messages = append(messages, query.ReadOnlyWarning(stmt.String()))
		}
		err = e.executeDropRetentionPolicyStatement(stmt)
	case *influxql.DropShardStatement:
		if ctx.ReadOnly {
			messages = append(messages, query.ReadOnlyWarning(stmt.String()))
		}
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.DropSubscriptionStatement:
		if ctx.ReadOnly {
			messages = append(messages, query.ReadOnlyWarning(stmt.String()))
		}
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.DropUserStatement:
		if ctx.ReadOnly {
			messages = append(messages, query.ReadOnlyWarning(stmt.String()))
		}
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.ExplainStatement:
		if stmt.Analyze {
			err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
		} else {
			err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
		}
	case *influxql.GrantStatement:
		if ctx.ReadOnly {
			messages = append(messages, query.ReadOnlyWarning(stmt.String()))
		}
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.GrantAdminStatement:
		if ctx.ReadOnly {
			messages = append(messages, query.ReadOnlyWarning(stmt.String()))
		}
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.RevokeStatement:
		if ctx.ReadOnly {
			messages = append(messages, query.ReadOnlyWarning(stmt.String()))
		}
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.RevokeAdminStatement:
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.ShowContinuousQueriesStatement:
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.ShowDatabasesStatement:
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.ShowDiagnosticsStatement:
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.ShowGrantsForUserStatement:
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.ShowMeasurementsStatement:
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.ShowMeasurementCardinalityStatement:
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.ShowRetentionPoliciesStatement:
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.ShowSeriesCardinalityStatement:
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.ShowShardsStatement:
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.ShowShardGroupsStatement:
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.ShowStatsStatement:
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.ShowSubscriptionsStatement:
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.ShowTagKeysStatement:
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.ShowTagValuesStatement:
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.ShowUsersStatement:
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.SetPasswordUserStatement:
		if ctx.ReadOnly {
			messages = append(messages, query.ReadOnlyWarning(stmt.String()))
		}
		err = e.StatementExecutor.ExecuteStatement(stmt, ctx)
	case *influxql.ShowQueriesStatement, *influxql.KillQueryStatement:
		// Send query related statements to the task manager.
		return e.TaskManager.ExecuteStatement(stmt, ctx)
	default:
		return query.ErrInvalidQuery
	}

	return err
}

func (e *StatementExecutor) executeDropDatabaseStatement(stmt *influxql.DropDatabaseStatement) error {
	if e.MetaClient.Database(stmt.Name) == nil {
		return nil
	}

	// Locally delete the datababse.
	if err := e.TSDBStore.DeleteDatabase(stmt.Name); err != nil {
		return err
	}

	// Remove the database from the Meta Store.
	e.MetaClient.DropDatabase(stmt.Name)

	return e.MetaExecutor.ExecuteStatement(stmt,"")
}

func (e *StatementExecutor)  executeDropMeasurementStatement(stmt *influxql.DropMeasurementStatement, database string) error {
	if e.MetaClient.Database(database) == nil {
		return nil
	}

	// Locally drop the measurement
	if err := e.TSDBStore.DeleteMeasurement(database, stmt.Name); err != nil {
		return err
	}

	// Execute the statement on the other data nodes in the cluster.
	return e.MetaExecutor.ExecuteStatement(stmt, database)
}

func (e *StatementExecutor)  executeDropRetentionPolicyStatement(stmt *influxql.DropRetentionPolicyStatement) error {
	if err := e.MetaClient.DropRetentionPolicy(stmt.Database, stmt.Name); err != nil {
		return err
	}

	// Locally drop the retention policy.
	if err := e.TSDBStore.DeleteRetentionPolicy(stmt.Database, stmt.Name); err != nil {
		return err
	}

	// Execute the statement on the other data nodes in the cluster.
	return e.MetaExecutor.ExecuteStatement(stmt, stmt.Database)
}

func (e *StatementExecutor)  executeDropSeriesStatement(stmt *influxql.DropSeriesStatement, database string) error {
	if  e.MetaClient.Database(database) == nil {
		return nil
	}

	// Check for time in WHERE clause (not supported).
	if influxql.HasTimeExpr(stmt.Condition) {
		return errors.New("DROP SERIES doesn't support time in WHERE clause")
	}

	// Locally drop the series.
	if err := e.TSDBStore.DeleteSeries(database, stmt.Sources, stmt.Condition); err != nil {
		return err
	}

	// Execute the statement on the other data nodes in the cluster.
	return e.MetaExecutor.ExecuteStatement(stmt, database)
}