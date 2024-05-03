package datastore

import (
	"database/sql"
	"errors"
	"github.com/PretendoNetwork/nex-go/v2"
	nextypes "github.com/PretendoNetwork/nex-go/v2/types"
	"github.com/PretendoNetwork/nex-protocols-go/v2/datastore/types"
	"github.com/PretendoNetwork/puyo-puyo-tetris/globals"
	"github.com/lib/pq"
	"time"
)

var selectByDataIdStmt *sql.Stmt
var selectByNameAndOwnerStmt *sql.Stmt
var insertObjectStmt *sql.Stmt
var updateMetaBinaryStmt *sql.Stmt

// === GETTERS ===

// TODO totalCount
func getObjects(stmt *sql.Stmt, args ...any) ([]*types.DataStoreMetaInfo, error) {
	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*types.DataStoreMetaInfo

	for rows.Next() {
		result := types.NewDataStoreMetaInfo()

		var createdTime time.Time
		var updatedTime time.Time
		var ownerID uint64
		var permissionRecipients []uint64
		var delPermissionRecipients []uint64
		var tags []string

		result.ExpireTime = nextypes.NewDateTime(0x9C3f3E0000) // * 9999-12-31T00:00:00.000Z. This is what the real server sends

		err := rows.Scan(
			&result.DataID.Value,
			&ownerID,
			&result.Size.Value,
			&result.Name.Value,
			&result.DataType.Value,
			&result.MetaBinary.Value,
			&result.Permission.Permission.Value,
			pq.Array(&permissionRecipients),
			&result.DelPermission.Permission.Value,
			pq.Array(&delPermissionRecipients),
			&result.Flag.Value,
			&result.Period.Value,
			&result.ReferDataID.Value,
			pq.Array(&tags),
			&createdTime,
			&updatedTime,
		)
		if err != nil {
			return nil, err
			//globals.Logger.Error(err.Error())
			//continue
		}

		result.OwnerID = nextypes.NewPID(ownerID)
		result.Permission.RecipientIDs = createPIDList(&permissionRecipients)
		result.DelPermission.RecipientIDs = createPIDList(&delPermissionRecipients)
		result.Tags = createStringList(&tags)
		result.CreatedTime.FromTimestamp(createdTime)
		result.UpdatedTime.FromTimestamp(updatedTime)
		result.ReferredTime.FromTimestamp(createdTime)

		results = append(results, result)
	}

	return results, rows.Err()
}

// part of ChangeMeta
func GetObjectInfoByDataID(dataID *nextypes.PrimitiveU64) (*types.DataStoreMetaInfo, *nex.Error) {
	objects, err := getObjects(selectByDataIdStmt, 1, dataID.Value)
	if errors.Is(err, sql.ErrNoRows) || len(objects) < 1 {
		return nil, nex.NewError(nex.ResultCodes.DataStore.NotFound, "Object not found")
	} else if err != nil {
		return nil, nex.NewError(nex.ResultCodes.DataStore.SystemFileError, err.Error())
	}

	return objects[0], nil
}

// part of GetMeta
func GetObjectInfoByDataIDWithPassword(dataID *nextypes.PrimitiveU64, _ *nextypes.PrimitiveU64) (*types.DataStoreMetaInfo, *nex.Error) {
	// TODO check password?
	return GetObjectInfoByDataID(dataID)
}

// part of GetMeta, not used
func GetObjectInfoByPersistenceTargetWithPassword(_ *types.DataStorePersistenceTarget, _ *nextypes.PrimitiveU64) (*types.DataStoreMetaInfo, *nex.Error) {
	return nil, nex.NewError(nex.ResultCodes.Core.NotImplemented, "GetObjectInfoByPersistenceTargetWithPassword unimplemented")
}

// get own gamedata (among other things)
func GetObjectInfosByDataStoreSearchParam(param *types.DataStoreSearchParam, pid *nextypes.PID) ([]*types.DataStoreMetaInfo, uint32, *nex.Error) {
	// TODO refactor this according to Jon's notes
	if param.DataType.Value == 65535 && param.SearchTarget.Value == 10 {
		// TODO "gamedata" hardcoded
		objects, err := getObjects(selectByNameAndOwnerStmt, 1, "gamedata", pid.Value())
		if err != nil {
			return nil, 0, nex.NewError(nex.ResultCodes.DataStore.SystemFileError, err.Error())
		}

		return objects, uint32(len(objects)), nil
	}

	globals.Logger.Warning("Unknown GetObjectInfosByDataStoreSearchParam!")
	return nil, 0, nil
}

// === INSERTERS === (long-handed)

// uploading gamedata meta object
func InitializeObjectByPreparePostParam(ownerPID *nextypes.PID, param *types.DataStorePreparePostParam) (uint64, *nex.Error) {
	now := time.Now()

	var dataID uint64

	err := insertObjectStmt.QueryRow(
		ownerPID.Value(),
		param.Size.Value,
		param.Name.Value,
		param.DataType.Value,
		param.MetaBinary.Value,
		param.Permission.Permission.Value,
		pq.Array(convertPIDList(param.Permission.RecipientIDs)),
		param.DelPermission.Permission.Value,
		pq.Array(convertPIDList(param.DelPermission.RecipientIDs)),
		param.Flag.Value,
		param.Period.Value,
		param.ReferDataID.Value,
		pq.Array(convertStringList(param.Tags)),
		param.PersistenceInitParam.PersistenceSlotID.Value, // todo DeleteLastObject
		pq.Array(convertStringList(param.ExtraData)),
		now,
		now,
	).Scan(&dataID)
	if err != nil {
		return 0, nex.NewError(nex.ResultCodes.DataStore.SystemFileError, err.Error())
	}

	return dataID, nil
}

// maybe unused?
func InitializeObjectRatingWithSlot(_ uint64, _ *types.DataStoreRatingInitParamWithSlot) *nex.Error {
	return nex.NewError(nex.ResultCodes.Core.NotImplemented, "InitializeObjectRatingWithSlot unimplemented")
}

// === UPDATERS ===

// Parts of ChangeMeta
func UpdateObjectPeriodByDataIDWithPassword(_ *nextypes.PrimitiveU64, _ *nextypes.PrimitiveU16, _ *nextypes.PrimitiveU64) *nex.Error {
	return nex.NewError(nex.ResultCodes.Core.NotImplemented, "UpdateObjectPeriodByDataIDWithPassword unimplemented")
}

// Parts of ChangeMeta
func UpdateObjectMetaBinaryByDataIDWithPassword(dataID *nextypes.PrimitiveU64, metaBinary *nextypes.QBuffer, _ *nextypes.PrimitiveU64) *nex.Error {
	result, err := updateMetaBinaryStmt.Exec(metaBinary.Value, dataID.Value)
	if err != nil {
		return nex.NewError(nex.ResultCodes.DataStore.SystemFileError, err.Error())
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nex.NewError(nex.ResultCodes.DataStore.SystemFileError, err.Error())
	} else if rows < 1 {
		return nex.NewError(nex.ResultCodes.DataStore.NotFound, "Object not found")
	}

	return nil
}

// Parts of ChangeMeta
func UpdateObjectDataTypeByDataIDWithPassword(_ *nextypes.PrimitiveU64, _ *nextypes.PrimitiveU16, _ *nextypes.PrimitiveU64) *nex.Error {
	return nex.NewError(nex.ResultCodes.Core.NotImplemented, "UpdateObjectDataTypeByDataIDWithPassword unimplemented")
}

// === HELPERS ===

// Prepared statements
func initDatastore() {
	const selectObject = `
	SELECT
    	data_id,
    	owner,
    	size,
    	name,
    	data_type,
		meta_binary,
        permission,
     	permission_recipients,
     	delete_permission,
     	delete_permission_recipients,
     	flag,
        period,
     	refer_data_id,
     	tags,
     	creation_date,
     	update_date
	FROM datastore.objects`

	stmt, err := Postgres.Prepare(selectObject + ` WHERE name = $2 AND owner = $3 LIMIT $1`)
	if err != nil {
		panic(err)
	}
	selectByNameAndOwnerStmt = stmt

	stmt, err = Postgres.Prepare(selectObject + ` WHERE data_id = $2 LIMIT $1`)
	if err != nil {
		panic(err)
	}
	selectByDataIdStmt = stmt

	stmt, err = Postgres.Prepare(`INSERT INTO datastore.objects 
    (
    	owner,
     	size,
     	name,
     	data_type,
     	meta_binary,
        permission,
     	permission_recipients,
     	delete_permission,
     	delete_permission_recipients,
     	flag,
        period,
     	refer_data_id,
     	tags, 
     	persistence_slot_id,
     	extra_data,
     	creation_date,
     	update_date
    )
	VALUES (
	        $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
	) RETURNING data_id`)
	if err != nil {
		panic(err)
	}
	insertObjectStmt = stmt

	stmt, err = Postgres.Prepare(`UPDATE datastore.objects SET meta_binary = $1 WHERE data_id = $2`)
	if err != nil {
		panic(err)
	}
	updateMetaBinaryStmt = stmt
}

// Helpers for nex types
func convertPIDList(list *nextypes.List[*nextypes.PID]) []uint64 {
	result := make([]uint64, list.Length())

	for i, pid := range list.Slice() {
		result[i] = pid.Value()
	}

	return result
}

func createPIDList(list *[]uint64) *nextypes.List[*nextypes.PID] {
	result := make([]*nextypes.PID, len(*list))

	for i, u := range *list {
		result[i] = nextypes.NewPID(u)
	}

	nexlist := nextypes.NewList[*nextypes.PID]()
	nexlist.SetFromData(result)

	return nexlist
}

func convertStringList(list *nextypes.List[*nextypes.String]) []string {
	result := make([]string, list.Length())

	for i, n := range list.Slice() {
		result[i] = n.Value
	}

	return result
}

func createStringList(list *[]string) *nextypes.List[*nextypes.String] {
	result := make([]*nextypes.String, len(*list))

	for i, u := range *list {
		result[i] = nextypes.NewString(u)
	}

	nexlist := nextypes.NewList[*nextypes.String]()
	nexlist.SetFromData(result)

	return nexlist
}
