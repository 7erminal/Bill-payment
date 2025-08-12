package main

import (
	"github.com/beego/beego/v2/client/orm/migration"
)

// DO NOT MODIFY
type Billers_20250628_052741 struct {
	migration.Migration
}

// DO NOT MODIFY
func init() {
	m := &Billers_20250628_052741{}
	m.Created = "20250628_052741"

	migration.Register("Billers_20250628_052741", m)
}

// Run the migrations
func (m *Billers_20250628_052741) Up() {
	// use m.SQL("CREATE TABLE ...") to make schema update
	m.SQL("CREATE TABLE billers(`biller_id` int(11) NOT NULL AUTO_INCREMENT,`biller_name` varchar(80) NOT NULL,`biller_code` varchar(80) NOT NULL,`biller_reference_id` varchar(250) DEFAULT NULL,`description` varchar(255) DEFAULT NULL,`operator_id` int(11) NOT NULL,`date_created` datetime DEFAULT CURRENT_TIMESTAMP,`date_modified` datetime ON UPDATE CURRENT_TIMESTAMP,`created_by` int(11) DEFAULT NULL,`modified_by` int(11) DEFAULT NULL,`active` int(11) DEFAULT 1,PRIMARY KEY (`biller_id`), FOREIGN KEY (operator_id) REFERENCES operators(operator_id) ON UPDATE CASCADE ON DELETE NO ACTION)")
}

// Reverse the migrations
func (m *Billers_20250628_052741) Down() {
	// use m.SQL("DROP TABLE ...") to reverse schema update
	m.SQL("DROP TABLE `billers`")
}
