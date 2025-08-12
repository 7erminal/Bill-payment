package main

import (
	"github.com/beego/beego/v2/client/orm/migration"
)

// DO NOT MODIFY
type BilInsTransactions_20250728_211056 struct {
	migration.Migration
}

// DO NOT MODIFY
func init() {
	m := &BilInsTransactions_20250728_211056{}
	m.Created = "20250728_211056"

	migration.Register("BilInsTransactions_20250728_211056", m)
}

// Run the migrations
func (m *BilInsTransactions_20250728_211056) Up() {
	// use m.SQL("CREATE TABLE ...") to make schema update
	m.SQL("CREATE TABLE bil_ins_transactions(`bil_ins_transaction_id` int(11) NOT NULL AUTO_INCREMENT,`bil_transaction_id` int(11) DEFAULT NULL,`amount` float DEFAULT NULL,`biller_id` int(11) DEFAULT NULL,`sender_account_number` varchar(255) DEFAULT NULL,`recipient_account_number` varchar(255) DEFAULT NULL,`network` varchar(150) DEFAULT NULL,`request` text DEFAULT NULL,`response` text DEFAULT NULL,`date_created` datetime DEFAULT CURRENT_TIMESTAMP,`date_modified` datetime ON UPDATE CURRENT_TIMESTAMP,`created_by` int(11) DEFAULT 1,`modified_by` int(11) DEFAULT 1,`active` int(11) DEFAULT 1,PRIMARY KEY (`bil_ins_transaction_id`), FOREIGN KEY (bil_transaction_id) REFERENCES bil_transactions(transaction_id) ON UPDATE CASCADE ON DELETE NO ACTION, FOREIGN KEY (biller_id) REFERENCES billers(biller_id) ON UPDATE CASCADE ON DELETE NO ACTION)")
}

// Reverse the migrations
func (m *BilInsTransactions_20250728_211056) Down() {
	// use m.SQL("DROP TABLE ...") to reverse schema update
	m.SQL("DROP TABLE `bil_ins_transactions`")
}
