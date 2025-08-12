package main

import (
	"github.com/beego/beego/v2/client/orm/migration"
)

// DO NOT MODIFY
type AddColumnToRequestsTable_20250802_155036 struct {
	migration.Migration
}

// DO NOT MODIFY
func init() {
	m := &AddColumnToRequestsTable_20250802_155036{}
	m.Created = "20250802_155036"

	migration.Register("AddColumnToRequestsTable_20250802_155036", m)
}

// Run the migrations
func (m *AddColumnToRequestsTable_20250802_155036) Up() {
	// use m.SQL("CREATE TABLE ...") to make schema update
	m.SQL("ALTER TABLE `requests` ADD COLUMN `api_request_id` INT(11) NULL DEFAULT NULL COMMENT 'API Request ID' AFTER `request_id`")
}

// Reverse the migrations
func (m *AddColumnToRequestsTable_20250802_155036) Down() {
	// use m.SQL("DROP TABLE ...") to reverse schema update

}
