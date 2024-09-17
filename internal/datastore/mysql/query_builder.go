package mysql

import (
	"github.com/authzed/spicedb/internal/datastore/mysql/migrations"

	sq "github.com/Masterminds/squirrel"
)

// QueryBuilder captures all parameterizable queries used
// by the MySQL datastore implementation
type QueryBuilder struct {
	GetLastRevision   sq.SelectBuilder
	LoadRevisionRange sq.SelectBuilder

	WriteNamespaceQuery        sq.InsertBuilder
	ReadNamespaceQuery         sq.SelectBuilder
	DeleteNamespaceQuery       sq.UpdateBuilder
	DeleteNamespaceTuplesQuery sq.UpdateBuilder

	ReadCounterQuery   sq.SelectBuilder
	InsertCounterQuery sq.InsertBuilder
	DeleteCounterQuery sq.UpdateBuilder
	UpdateCounterQuery sq.UpdateBuilder

	QueryTuplesWithIdsQuery      sq.SelectBuilder
	QueryTuplesQuery             sq.SelectBuilder
	DeleteTupleQuery             sq.UpdateBuilder
	QueryRelationshipExistsQuery sq.SelectBuilder
	WriteTupleQuery              sq.InsertBuilder
	QueryChangedQuery            sq.SelectBuilder
	CountTupleQuery              sq.SelectBuilder

	WriteCaveatQuery  sq.InsertBuilder
	ReadCaveatQuery   sq.SelectBuilder
	ListCaveatsQuery  sq.SelectBuilder
	DeleteCaveatQuery sq.UpdateBuilder
}

// NewQueryBuilder returns a new QueryBuilder instance. The migration
// driver is used to determine the names of the tables.
func NewQueryBuilder(driver *migrations.MySQLDriver) *QueryBuilder {
	builder := QueryBuilder{}

	// transaction builders
	builder.GetLastRevision = getLastRevision(driver.RelationTupleTransaction())
	builder.LoadRevisionRange = loadRevisionRange(driver.RelationTupleTransaction())

	// namespace builders
	builder.WriteNamespaceQuery = writeNamespace(driver.Namespace())
	builder.ReadNamespaceQuery = readNamespace(driver.Namespace())
	builder.DeleteNamespaceQuery = deleteNamespace(driver.Namespace())

	// counters builders
	builder.ReadCounterQuery = readCounter(driver.RelationshipCounters())
	builder.InsertCounterQuery = insertCounter(driver.RelationshipCounters())
	builder.DeleteCounterQuery = deleteCounter(driver.RelationshipCounters())
	builder.UpdateCounterQuery = updateCounter(driver.RelationshipCounters())

	// tuple builders
	builder.QueryTuplesWithIdsQuery = queryTuplesWithIds(driver.RelationTuple())
	builder.DeleteNamespaceTuplesQuery = deleteNamespaceTuples(driver.RelationTuple())
	builder.QueryTuplesQuery = queryTuples(driver.RelationTuple())
	builder.DeleteTupleQuery = deleteTuple(driver.RelationTuple())
	builder.QueryRelationshipExistsQuery = queryRelationshipExists(driver.RelationTuple())
	builder.WriteTupleQuery = writeTuple(driver.RelationTuple())
	builder.QueryChangedQuery = queryChanged(driver.RelationTuple())
	builder.CountTupleQuery = countRels(driver.RelationTuple())

	// caveat builders
	builder.ReadCaveatQuery = readCaveat(driver.Caveat())
	builder.ListCaveatsQuery = listCaveats(driver.Caveat())
	builder.WriteCaveatQuery = writeCaveat(driver.Caveat())
	builder.DeleteCaveatQuery = deleteCaveat(driver.Caveat())

	return &builder
}

func listCaveats(tableCaveat string) sq.SelectBuilder {
	return sb.Select(colCaveatDefinition, colCreatedTxn).From(tableCaveat).OrderBy(colName)
}

func deleteCaveat(tableCaveat string) sq.UpdateBuilder {
	return sb.Update(tableCaveat).Where(sq.Eq{colDeletedTxn: liveDeletedTxnID})
}

func writeCaveat(tableCaveat string) sq.InsertBuilder {
	return sb.Insert(tableCaveat).Columns(
		colName,
		colCaveatDefinition,
		colCreatedTxn,
	)
}

func readCaveat(tableCaveat string) sq.SelectBuilder {
	return sb.Select(colCaveatDefinition, colCreatedTxn).From(tableCaveat)
}

func getLastRevision(tableTransaction string) sq.SelectBuilder {
	return sb.Select("MAX(id)").From(tableTransaction).Limit(1)
}

func loadRevisionRange(tableTransaction string) sq.SelectBuilder {
	return sb.Select(colID, colMetadata).From(tableTransaction)
}

func readCounter(tableRelationshipCounters string) sq.SelectBuilder {
	return sb.Select(
		colCounterName,
		colCounterSerializedFilter,
		colCounterCurrentCount,
		colCounterUpdatedAtRevision,
	).From(tableRelationshipCounters)
}

func insertCounter(tableRelationshipCounters string) sq.InsertBuilder {
	return sb.Insert(tableRelationshipCounters).Columns(
		colCounterName,
		colCounterSerializedFilter,
		colCounterCurrentCount,
		colCounterUpdatedAtRevision,
		colCreatedTxn,
	)
}

func deleteCounter(tableRelationshipCounters string) sq.UpdateBuilder {
	return sb.Update(tableRelationshipCounters).Where(sq.Eq{colDeletedTxn: liveDeletedTxnID})
}

func updateCounter(tableRelationshipCounters string) sq.UpdateBuilder {
	return sb.Update(tableRelationshipCounters).Where(sq.Eq{colDeletedTxn: liveDeletedTxnID})
}

func writeNamespace(tableNamespace string) sq.InsertBuilder {
	return sb.Insert(tableNamespace).Columns(
		colNamespace,
		colConfig,
		colCreatedTxn,
	)
}

func readNamespace(tableNamespace string) sq.SelectBuilder {
	return sb.Select(colConfig, colCreatedTxn).From(tableNamespace)
}

func deleteNamespace(tableNamespace string) sq.UpdateBuilder {
	return sb.Update(tableNamespace).Where(sq.Eq{colDeletedTxn: liveDeletedTxnID})
}

func deleteNamespaceTuples(tableTuple string) sq.UpdateBuilder {
	return sb.Update(tableTuple).Where(sq.Eq{colDeletedTxn: liveDeletedTxnID})
}

func queryTuplesWithIds(tableTuple string) sq.SelectBuilder {
	return sb.Select(
		colID,
		colNamespace,
		colObjectID,
		colRelation,
		colUsersetNamespace,
		colUsersetObjectID,
		colUsersetRelation,
		colCaveatName,
		colCaveatContext,
	).From(tableTuple)
}

func queryTuples(tableTuple string) sq.SelectBuilder {
	return sb.Select(
		colNamespace,
		colObjectID,
		colRelation,
		colUsersetNamespace,
		colUsersetObjectID,
		colUsersetRelation,
		colCaveatName,
		colCaveatContext,
	).From(tableTuple)
}

func countRels(tableTuple string) sq.SelectBuilder {
	return sb.Select(
		"count(*)",
	).From(tableTuple)
}

func deleteTuple(tableTuple string) sq.UpdateBuilder {
	return sb.Update(tableTuple).Where(sq.Eq{colDeletedTxn: liveDeletedTxnID})
}

func queryRelationshipExists(tableTuple string) sq.SelectBuilder {
	return sb.Select(colID).From(tableTuple)
}

func writeTuple(tableTuple string) sq.InsertBuilder {
	return sb.Insert(tableTuple).Columns(
		colNamespace,
		colObjectID,
		colRelation,
		colUsersetNamespace,
		colUsersetObjectID,
		colUsersetRelation,
		colCaveatName,
		colCaveatContext,
		colCreatedTxn,
	)
}

func queryChanged(tableTuple string) sq.SelectBuilder {
	return sb.Select(
		colNamespace,
		colObjectID,
		colRelation,
		colUsersetNamespace,
		colUsersetObjectID,
		colUsersetRelation,
		colCaveatName,
		colCaveatContext,
		colCreatedTxn,
		colDeletedTxn,
	).From(tableTuple)
}
