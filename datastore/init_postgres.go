package datastore

import (
	"github.com/PretendoNetwork/puyo-puyo-tetris/globals"
	"os"
)

func initPostgres() {
	_, err := Postgres.Exec(`CREATE SCHEMA IF NOT EXISTS datastore`)
	if err != nil {
		globals.Logger.Critical(err.Error())
		os.Exit(0)
	}

	globals.Logger.Success("datastore Postgres schema created")

	_, err = Postgres.Exec(`CREATE SEQUENCE IF NOT EXISTS datastore.object_data_id_seq
		INCREMENT 1
		MINVALUE 1
		MAXVALUE 281474976710656
		START 1
		CACHE 1`, // * Honestly I don't know what CACHE does but I saw it recommended so here it is
	)
	if err != nil {
		globals.Logger.Critical(err.Error())
		os.Exit(0)
	}

	_, err = Postgres.Exec(`CREATE TABLE IF NOT EXISTS datastore.objects (
		data_id bigint NOT NULL DEFAULT nextval('datastore.object_data_id_seq') PRIMARY KEY,
		upload_completed boolean NOT NULL DEFAULT FALSE,
		deleted boolean NOT NULL DEFAULT FALSE,
		owner bigint,
		size int,
		name text,
		data_type int,
		meta_binary bytea,
		permission int,
		permission_recipients int[],
		delete_permission int,
		delete_permission_recipients int[],
		flag int,
		period int,
		refer_data_id bigint,
		tags text[],
		persistence_slot_id int,
		extra_data text[],
		access_password bigint NOT NULL DEFAULT 0,
		update_password bigint NOT NULL DEFAULT 0,
		creation_date timestamp,
		update_date timestamp
	)`)
	if err != nil {
		globals.Logger.Critical(err.Error())
		os.Exit(0)
	}

	//// * Unsure what like half of this is but the client sends it so we saves it
	//_, err = Postgres.Exec(`CREATE TABLE IF NOT EXISTS datastore.object_ratings (
	//	data_id bigint,
	//	slot smallint,
	//	flag smallint,
	//	internal_flag smallint,
	//	lock_type smallint,
	//	initial_value bigint,
	//	range_min int,
	//	range_max int,
	//	period_hour smallint,
	//	period_duration int,
	//	total_value bigint,
	//	count int NOT NULL DEFAULT 0,
	//	PRIMARY KEY(data_id, slot)
	//)`)
	//if err != nil {
	//	globals.Logger.Critical(err.Error())
	//	os.Exit(0)
	//}

	globals.Logger.Success("Postgres tables created")
}
